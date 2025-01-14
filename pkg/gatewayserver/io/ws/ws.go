// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ws

import (
	"context"
	"encoding/hex"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	echo "github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/random"
	"go.thethings.network/lorawan-stack/v3/pkg/ratelimit"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
	"go.thethings.network/lorawan-stack/v3/pkg/web"
	"go.thethings.network/lorawan-stack/v3/pkg/web/middleware"
	"google.golang.org/grpc/metadata"
)

var (
	errGatewayID      = errors.DefineInvalidArgument("invalid_gateway_id", "invalid gateway ID `{id}`")
	errNoAuthProvided = errors.DefineUnauthenticated("no_auth_provided", "no auth provided `{uid}`")
)

type srv struct {
	ctx       context.Context
	server    io.Server
	webServer *echo.Echo
	upgrader  *websocket.Upgrader
	cfg       Config
	formatter Formatter
}

func (s *srv) Protocol() string            { return "ws" }
func (s *srv) SupportsDownlinkClaim() bool { return false }

// New creates a new WebSocket frontend.
func New(ctx context.Context, server io.Server, formatter Formatter, cfg Config) *echo.Echo {
	ctx = log.NewContextWithField(ctx, "namespace", "gatewayserver/io/ws")

	webServer := echo.New()
	webServer.Logger = web.NewNoopLogger()
	webServer.HTTPErrorHandler = errorHandler
	webServer.Use(
		middleware.ID(""),
		echomiddleware.BodyLimit("16M"),
		middleware.Log(log.FromContext(ctx)),
		ratelimit.EchoMiddleware(server.RateLimiter(), "gs:accept:ws"),
		middleware.Recover(),
	)

	s := &srv{
		ctx:       ctx,
		server:    server,
		upgrader:  &websocket.Upgrader{},
		webServer: webServer,
		formatter: formatter,
		cfg:       cfg,
	}

	eps := s.formatter.Endpoints()
	webServer.GET(eps.ConnectionInfo, s.handleConnectionInfo)
	webServer.GET(eps.Traffic, s.handleTraffic)

	go func() {
		<-ctx.Done()
		webServer.Close()
	}()

	return webServer
}

func (s *srv) handleConnectionInfo(c echo.Context) error {
	ctx := c.Request().Context()
	eps := s.formatter.Endpoints()
	logger := log.FromContext(ctx).WithFields(log.Fields(
		"endpoint", eps.ConnectionInfo,
		"remote_addr", c.Request().RemoteAddr,
	))
	ws, err := s.upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		logger.WithError(err).Debug("Failed to upgrade request to websocket connection")
		return err
	}
	defer ws.Close()

	_, data, err := ws.ReadMessage()
	if err != nil {
		logger.WithError(err).Debug("Failed to read message")
		return err
	}

	scheme := "ws"
	if c.IsTLS() || s.cfg.UseTrafficTLSAddress {
		scheme = "wss"
	}

	info := ServerInfo{
		Scheme:  scheme,
		Address: c.Request().Host,
	}

	resp := s.formatter.HandleConnectionInfo(ctx, data, s.server, info, time.Now())
	if err := ws.WriteMessage(websocket.TextMessage, resp); err != nil {
		logger.WithError(err).Warn("Failed to write connection info response message")
		return err
	}
	logger.Debug("Sent connection info response message")
	return nil
}

var euiHexPattern = regexp.MustCompile("^eui-([a-f0-9A-F]{16})$")

func (s *srv) handleTraffic(c echo.Context) (err error) {
	id := c.Param("id")
	auth := c.Request().Header.Get(echo.HeaderAuthorization)
	ctx := c.Request().Context()
	eps := s.formatter.Endpoints()

	ctx = log.NewContextWithFields(ctx, log.Fields(
		"endpoint", eps.Traffic,
		"remote_addr", c.Request().RemoteAddr,
	))
	ctx = NewContextWithSession(ctx, &Session{})

	// Convert the ID to EUI.
	str := euiHexPattern.FindStringSubmatch(id)
	if len(str) != 2 {
		return errGatewayID.WithAttributes("id", id)
	}
	hexValue, err := hex.DecodeString(str[1])
	if err != nil {
		return errGatewayID.WithAttributes("id", id)
	}
	var eui types.EUI64
	eui.UnmarshalBinary(hexValue)

	ctx, ids, err := s.server.FillGatewayContext(ctx, ttnpb.GatewayIdentifiers{Eui: &eui})
	if err != nil {
		return err
	}

	uid := unique.ID(ctx, ids)
	ctx = log.NewContextWithField(ctx, "gateway_uid", uid)

	var md metadata.MD

	if auth != "" {
		if !strings.HasPrefix(auth, "Bearer ") {
			auth = fmt.Sprintf("Bearer %s", auth)
		}
		md = metadata.New(map[string]string{
			"id":            ids.GatewayId,
			"authorization": auth,
		})
	}

	if ctxMd, ok := metadata.FromIncomingContext(ctx); ok {
		md = metadata.Join(ctxMd, md)
	}
	ctx = metadata.NewIncomingContext(ctx, md)
	// If a fallback frequency is defined in the server context, inject it into local the context.
	if fallback, ok := frequencyplans.FallbackIDFromContext(s.ctx); ok {
		ctx = frequencyplans.WithFallbackID(ctx, fallback)
	}

	if auth == "" {
		// If the server allows unauthenticated connections (for local testing), we provide the link rights ourselves.
		if s.cfg.AllowUnauthenticated {
			ctx = rights.NewContext(ctx, rights.Rights{
				GatewayRights: map[string]*ttnpb.Rights{
					uid: {
						Rights: []ttnpb.Right{ttnpb.RIGHT_GATEWAY_LINK},
					},
				},
			})
		} else {
			// We error here directly as there is no need make an RPC call to the IS to get a failed rights check due to no Auth.
			return errNoAuthProvided.WithAttributes("uid", uid)
		}
	}

	logger := log.FromContext(ctx)

	conn, err := s.server.Connect(ctx, s, ids)
	if err != nil {
		logger.WithError(err).Warn("Failed to connect")
		return err
	}

	ws, err := s.upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		logger.WithError(err).Debug("Failed to upgrade request to websocket connection")
		conn.Disconnect(err)
		return err
	}
	defer ws.Close()
	wsWriteMu := &sync.Mutex{}

	defer func() {
		conn.Disconnect(err)
		err = nil // Errors are sent over the websocket connection that is established by this point.
	}()

	pingTicker := time.NewTicker(random.Jitter(s.cfg.WSPingInterval, 0.1))
	defer pingTicker.Stop()

	ws.SetPingHandler(func(data string) error {
		logger.Debug("Received ping from gateway, send pong")
		wsWriteMu.Lock()
		defer wsWriteMu.Unlock()
		if err := ws.WriteMessage(websocket.PongMessage, nil); err != nil {
			logger.WithError(err).Warn("Failed to send pong")
			return err
		}
		return nil
	})

	// Not all gateways support pongs to the server's pings.
	ws.SetPongHandler(func(data string) error {
		logger.Debug("Received pong from gateway")
		return nil
	})

	var timeSyncTickerC <-chan time.Time
	if s.cfg.TimeSyncInterval > 0 {
		ticker := time.NewTicker(random.Jitter(s.cfg.TimeSyncInterval, 0.1))
		timeSyncTickerC = ticker.C
		defer ticker.Stop()
	}

	go func() {
		defer ws.Close()
		for {
			select {
			case <-conn.Context().Done():
				return
			case <-pingTicker.C:
				wsWriteMu.Lock()
				err := ws.WriteMessage(websocket.PingMessage, nil)
				wsWriteMu.Unlock()
				if err != nil {
					logger.WithError(err).Warn("Failed to send ping message")
					conn.Disconnect(err)
					return
				}
			case <-timeSyncTickerC:
				b, err := s.formatter.TransferTime(ctx, time.Now(), conn)
				if err != nil {
					logger.WithError(err).Warn("Failed to generate time transfer")
					conn.Disconnect(err)
					return
				}
				if b == nil {
					continue
				}

				wsWriteMu.Lock()
				err = ws.WriteMessage(websocket.TextMessage, b)
				wsWriteMu.Unlock()
				if err != nil {
					logger.WithError(err).Warn("Failed to transfer time")
					conn.Disconnect(err)
					return
				}
			case down := <-conn.Down():
				concentratorTime, ok := conn.TimeFromTimestampTime(down.GetScheduled().Timestamp)
				if !ok {
					logger.Warn("No clock synchronization")
					continue
				}
				dnmsg, err := s.formatter.FromDownlink(ctx, *down, conn.BandID(), concentratorTime, time.Now())
				if err != nil {
					logger.WithError(err).Warn("Failed to marshal downlink message")
					continue
				}

				wsWriteMu.Lock()
				err = ws.WriteMessage(websocket.TextMessage, dnmsg)
				wsWriteMu.Unlock()
				if err != nil {
					logger.WithError(err).Warn("Failed to send downlink message")
					conn.Disconnect(err)
					return
				}
			}
		}
	}()

	resource := ratelimit.GatewayUpResource(ctx, ids)
	for {
		if err := ratelimit.Require(s.server.RateLimiter(), resource); err != nil {
			logger.WithError(err).Warn("Terminate connection")
			conn.Disconnect(err)
			return err
		}
		_, data, err := ws.ReadMessage()
		if err != nil {
			logger.WithError(err).Debug("Failed to read message")
			return err
		}
		downstream, err := s.formatter.HandleUp(ctx, data, ids, conn, time.Now())
		if err != nil {
			return err
		}
		if downstream != nil {
			logger.Info("Send downstream message")
			wsWriteMu.Lock()
			err = ws.WriteMessage(websocket.TextMessage, downstream)
			wsWriteMu.Unlock()
			if err != nil {
				logger.WithError(err).Warn("Failed to send message downstream")
				conn.Disconnect(err)
				return err
			}
		}
	}
}

type errorMessage struct {
	Message string `json:"message"`
}

// errorHandler is an echo.HTTPErrorHandler.
func errorHandler(err error, c echo.Context) {
	if httpErr, ok := err.(*echo.HTTPError); ok {
		c.JSON(httpErr.Code, httpErr.Message)
		return
	}

	statusCode, description := http.StatusInternalServerError, ""
	if ttnErr, ok := errors.From(err); ok {
		if !errors.IsInternal(ttnErr) {
			description = ttnErr.Error()
		}
		statusCode = errors.ToHTTPStatusCode(ttnErr)
	}
	if description != "" {
		c.JSON(statusCode, errorMessage{description})
	} else {
		c.NoContent(statusCode)
	}
}
