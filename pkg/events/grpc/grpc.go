// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

// Package grpc contains an implementation of the EventsServer, which is used to
// stream all events published for a set of identifiers.
package grpc

import (
	"context"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	grpc_runtime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights/rightsutil"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/gogoproto"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmiddleware/warning"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const workersPerCPU = 2

// NewEventsServer returns a new EventsServer on the given PubSub.
func NewEventsServer(ctx context.Context, pubsub events.PubSub) *EventsServer {
	if _, ok := pubsub.(events.Store); ok {
		log.FromContext(ctx).Infof("Events PubSub: %T is also a Store!", pubsub)
	}
	definedNames := make(map[string]struct{})
	for _, def := range events.All().Definitions() {
		definedNames[def.Name()] = struct{}{}
	}
	return &EventsServer{
		ctx:          ctx,
		pubsub:       pubsub,
		definedNames: definedNames,
	}
}

// EventsServer streams events from a PubSub over gRPC.
type EventsServer struct {
	ctx          context.Context
	pubsub       events.PubSub
	definedNames map[string]struct{}
}

var (
	errInvalidRegexp    = errors.DefineInvalidArgument("invalid_regexp", "invalid regexp")
	errNoMatchingEvents = errors.DefineInvalidArgument("no_matching_events", "no matching events for regexp `{regexp}`")
	errUnknownEventName = errors.DefineInvalidArgument("unknown_event_name", "unknown event `{name}`")
)

func (srv *EventsServer) processNames(names ...string) ([]string, error) {
	if len(names) == 0 {
		return nil, nil
	}
	nameMap := make(map[string]struct{})
	for _, name := range names {
		if strings.HasPrefix(name, "/") && strings.HasSuffix(name, "/") {
			re, err := regexp.Compile(strings.Trim(name, "/"))
			if err != nil {
				return nil, errInvalidRegexp.WithCause(err)
			}
			var found bool
			for defined := range srv.definedNames {
				if re.MatchString(defined) {
					nameMap[defined] = struct{}{}
					found = true
				}
			}
			if !found {
				return nil, errNoMatchingEvents.WithAttributes("regexp", re.String())
			}
		} else {
			var found bool
			for defined := range srv.definedNames {
				if name == defined {
					nameMap[name] = struct{}{}
					found = true
					break
				}
			}
			if !found {
				return nil, errUnknownEventName.WithAttributes("name", name)
			}
		}
	}
	if len(nameMap) == 0 {
		return nil, nil
	}
	out := make([]string, 0, len(nameMap))
	for name := range nameMap {
		out = append(out, name)
	}
	sort.Strings(out)
	return out, nil
}

var errNoIdentifiers = errors.DefineInvalidArgument("no_identifiers", "no identifiers")

// Stream implements the EventsServer interface.
func (srv *EventsServer) Stream(req *ttnpb.StreamEventsRequest, stream ttnpb.Events_StreamServer) error {
	if len(req.Identifiers) == 0 {
		return errNoIdentifiers.New()
	}

	names, err := srv.processNames(req.Names...)
	if err != nil {
		return err
	}

	ctx := stream.Context()

	if err := rights.RequireAny(ctx, req.Identifiers...); err != nil {
		return err
	}

	ch := make(events.Channel, 8)
	handler := events.ContextHandler(ctx, ch)

	store, hasStore := srv.pubsub.(events.Store)
	var group *errgroup.Group
	if hasStore {
		if req.After == nil && req.Tail == 0 {
			now := time.Now()
			req.After = ttnpb.ProtoTimePtr(now)
		}
		group, ctx = errgroup.WithContext(ctx)
		group.Go(func() error {
			return store.SubscribeWithHistory(ctx, names, req.Identifiers, ttnpb.StdTime(req.After), int(req.Tail), handler)
		})
	} else {
		if req.Tail > 0 || req.After != nil {
			warning.Add(ctx, "Events storage is not enabled")
		}
		if err := srv.pubsub.Subscribe(ctx, names, req.Identifiers, handler); err != nil {
			return err
		}
	}

	if err := stream.SendHeader(metadata.MD{}); err != nil {
		return err
	}

	hostname, err := os.Hostname()
	if err != nil {
		return err
	}
	startEvent := &ttnpb.Event{
		UniqueId:       events.NewCorrelationID(),
		Name:           "events.stream.start",
		Time:           ttnpb.ProtoTimePtr(time.Now()),
		Identifiers:    req.Identifiers,
		Origin:         hostname,
		CorrelationIds: events.CorrelationIDsFromContext(ctx),
	}

	if len(names) > 0 {
		value, err := gogoproto.Value(names)
		if err != nil {
			return err
		}
		startEvent.Data, err = pbtypes.MarshalAny(value)
		if err != nil {
			return err
		}
	}

	if err := stream.Send(startEvent); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			if group != nil {
				return group.Wait()
			}
			return ctx.Err()
		case evt := <-ch:
			isVisible, err := rightsutil.EventIsVisible(ctx, evt)
			if err != nil {
				if err := rights.RequireAny(ctx, req.Identifiers...); err != nil {
					return err
				}
				log.FromContext(ctx).WithError(err).Warn("Failed to check event visibility")
				continue
			}
			if !isVisible {
				continue
			}
			proto, err := events.Proto(evt)
			if err != nil {
				log.FromContext(ctx).WithError(err).Warn("Failed to convert event to proto")
				continue
			}
			if err := stream.Send(proto); err != nil {
				return err
			}
		}
	}
}

var errStorageDisabled = errors.DefineFailedPrecondition("storage_disabled", "events storage is not not enabled")

// FindRelated implements the EventsServer interface.
func (srv *EventsServer) FindRelated(ctx context.Context, req *ttnpb.FindRelatedEventsRequest) (*ttnpb.FindRelatedEventsResponse, error) {
	store, hasStore := srv.pubsub.(events.Store)
	if !hasStore {
		return nil, errStorageDisabled.New()
	}
	_, err := rights.AuthInfo(ctx)
	if err != nil {
		return nil, err
	}

	evts, err := store.FindRelated(ctx, req.GetCorrelationId())
	if err != nil {
		return nil, err
	}

	var res ttnpb.FindRelatedEventsResponse

	for _, evt := range evts {
		evtProto, err := events.Proto(evt)
		if err != nil {
			log.FromContext(ctx).WithError(err).Warn("Failed to convert event to proto")
			continue
		}
		isVisible, err := rightsutil.EventIsVisible(ctx, evt)
		if err != nil {
			log.FromContext(ctx).WithError(err).Warn("Failed to check event visibility")
			continue
		}
		if isVisible {
			res.Events = append(res.Events, evtProto)
		} else {
			res.Events = append(res.Events, &ttnpb.Event{
				Name:        evtProto.Name,
				Time:        evtProto.Time,
				Identifiers: evtProto.Identifiers,
				// Data is private
				// CorrelationIDs is private
				Origin: evtProto.Origin,
				// Context is private
				Visibility: evtProto.Visibility,
				// Authentication is private
				// RemoteIP is private
				// UserAgent is private
				UniqueId: evtProto.UniqueId,
			})
		}
	}
	return &res, nil
}

// Roles implements rpcserver.Registerer.
func (srv *EventsServer) Roles() []ttnpb.ClusterRole {
	return nil
}

// RegisterServices implements rpcserver.Registerer.
func (srv *EventsServer) RegisterServices(s *grpc.Server) {
	ttnpb.RegisterEventsServer(s, srv)
}

// RegisterHandlers implements rpcserver.Registerer.
func (srv *EventsServer) RegisterHandlers(s *grpc_runtime.ServeMux, conn *grpc.ClientConn) {
	ttnpb.RegisterEventsHandler(srv.ctx, s, conn)
}
