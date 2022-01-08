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

package lbslns

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/band"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io/ws"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/scheduling"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestFromDownlinkMessage(t *testing.T) {
	_, ctx := test.New(t)
	ctx = ws.NewContextWithSession(ctx, &ws.Session{})
	updateSessionID(ctx, 0x11)
	var lbsLNS lbsLNS
	for _, tc := range []struct {
		BandID,
		Name string
		DownlinkMessage         ttnpb.DownlinkMessage
		ExpectedDownlinkMessage DownlinkMessage
	}{
		{
			BandID: band.EU_863_870,
			Name:   "SampleDownlink",
			DownlinkMessage: ttnpb.DownlinkMessage{
				RawPayload: []byte("Ymxhamthc25kJ3M=="),
				EndDeviceIds: &ttnpb.EndDeviceIdentifiers{
					DeviceId: "testdevice",
				},
				Settings: &ttnpb.DownlinkMessage_Scheduled{
					Scheduled: &ttnpb.TxSettings{
						DataRate: ttnpb.DataRate{
							Modulation: &ttnpb.DataRate_Lora{
								Lora: &ttnpb.LoRaDataRate{
									SpreadingFactor: 10,
									Bandwidth:       125000,
								},
							},
						},
						Frequency: 868500000,
						Downlink: &ttnpb.TxSettings_Downlink{
							AntennaIndex: 2,
						},
						Timestamp: 1553300787,
					},
				},
				CorrelationIds: []string{"correlation1"},
			},
			ExpectedDownlinkMessage: DownlinkMessage{
				DevEUI:      "00-00-00-00-00-00-00-01",
				DeviceClass: 0,
				Diid:        1,
				Pdu:         "596d7868616d74686332356b4a334d3d3d",
				RxDelay:     1,
				Rx1DR:       2,
				Rx1Freq:     868500000,
				RCtx:        2,
				Priority:    25,
				MuxTime:     1554300787.123456,
			},
		},
		{
			BandID: band.EU_863_870,
			Name:   "WithAbsoluteTime",
			DownlinkMessage: ttnpb.DownlinkMessage{
				RawPayload: []byte("Ymxhamthc25kJ3M=="),
				EndDeviceIds: &ttnpb.EndDeviceIdentifiers{
					DeviceId: "testdevice",
				},
				Settings: &ttnpb.DownlinkMessage_Scheduled{
					Scheduled: &ttnpb.TxSettings{
						DataRate: ttnpb.DataRate{
							Modulation: &ttnpb.DataRate_Lora{
								Lora: &ttnpb.LoRaDataRate{
									SpreadingFactor: 10,
									Bandwidth:       125000,
								},
							},
						},
						Frequency: 869525000,
						Downlink: &ttnpb.TxSettings_Downlink{
							AntennaIndex: 2,
						},
					},
				},
				CorrelationIds: []string{"correlation2"},
			},
			ExpectedDownlinkMessage: DownlinkMessage{
				DevEUI:      "00-00-00-00-00-00-00-01",
				DeviceClass: 0,
				Diid:        2,
				Pdu:         "596d7868616d74686332356b4a334d3d3d",
				RxDelay:     1,
				Rx1DR:       2,
				Rx1Freq:     869525000,
				RCtx:        2,
				Priority:    25,
				MuxTime:     1554300787.123456,
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)
			raw, err := lbsLNS.FromDownlink(ctx, tc.DownlinkMessage, tc.BandID, 1554300787, time.Unix(1554300787, 123456000))
			if !a.So(err, should.BeNil) {
				t.FailNow()
			}
			var dnmsg DownlinkMessage
			err = dnmsg.unmarshalJSON(raw)
			if !a.So(err, should.BeNil) {
				t.FailNow()
			}
			dnmsg.XTime = tc.ExpectedDownlinkMessage.XTime
			if !a.So(dnmsg, should.Resemble, tc.ExpectedDownlinkMessage) {
				t.Fatalf("Invalid DownlinkMessage: %v", dnmsg)
			}
		})
	}
}

func TestToDownlinkMessage(t *testing.T) {
	for _, tc := range []struct {
		BandID,
		Name string
		DownlinkMessage         DownlinkMessage
		ExpectedDownlinkMessage *ttnpb.DownlinkMessage
	}{
		{
			BandID: band.EU_863_870,
			Name:   "SampleDownlink",
			DownlinkMessage: DownlinkMessage{
				DeviceClass: 0,
				Pdu:         "Ymxhamthc25kJ3M==",
				RxDelay:     1,
				Rx1DR:       2,
				Rx1Freq:     868500000,
				RCtx:        2,
				Priority:    25,
				XTime:       1554300785,
			},
			ExpectedDownlinkMessage: &ttnpb.DownlinkMessage{
				RawPayload: []byte("Ymxhamthc25kJ3M=="),
				Settings: &ttnpb.DownlinkMessage_Scheduled{
					Scheduled: &ttnpb.TxSettings{
						DataRate: ttnpb.DataRate{
							Modulation: &ttnpb.DataRate_Lora{
								Lora: &ttnpb.LoRaDataRate{
									SpreadingFactor: 10,
									Bandwidth:       125000,
								},
							},
						},
						Frequency: 868500000,
						Downlink: &ttnpb.TxSettings_Downlink{
							AntennaIndex: 2,
						},
						Timestamp: 1554300785,
					},
				},
			},
		},
		{
			BandID: band.EU_863_870,
			Name:   "WithAbsoluteTime",
			DownlinkMessage: DownlinkMessage{
				DeviceClass: 1,
				Pdu:         "Ymxhamthc25kJ3M==",
				RxDelay:     1,
				Rx1DR:       2,
				Rx1Freq:     869525000,
				RCtx:        2,
				Priority:    25,
			},
			ExpectedDownlinkMessage: &ttnpb.DownlinkMessage{
				RawPayload: []byte("Ymxhamthc25kJ3M=="),
				Settings: &ttnpb.DownlinkMessage_Scheduled{
					Scheduled: &ttnpb.TxSettings{
						DataRate: ttnpb.DataRate{
							Modulation: &ttnpb.DataRate_Lora{
								Lora: &ttnpb.LoRaDataRate{
									SpreadingFactor: 10,
									Bandwidth:       125000,
								},
							},
						},
						Frequency: 869525000,
						Downlink: &ttnpb.TxSettings_Downlink{
							AntennaIndex: 2,
						},
					},
				},
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)
			dlMesg, err := tc.DownlinkMessage.ToDownlinkMessage(tc.BandID)
			if !a.So(err, should.BeNil) {
				t.FailNow()
			}
			if !a.So(dlMesg, should.Resemble, tc.ExpectedDownlinkMessage) {
				t.Fatalf("Invalid DownlinkMessage: %v", dlMesg)
			}
		})
	}
}

func TestTransferTime(t *testing.T) {
	a, ctx := test.New(t)

	ctx = ws.NewContextWithSession(ctx, &ws.Session{})

	f := (*lbsLNS)(nil)
	now := time.Unix(123, 456)

	// No timesync settings available in the session.
	b, err := f.TransferTime(ctx, now, nil, nil)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	a.So(b, should.BeNil)

	// Enable timesync for the session.
	updateSessionTimeSync(ctx, true)

	// No GPSTime / ConcentratorTime - expect only MuxTime.
	b, err = f.TransferTime(ctx, now, nil, nil)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	if a.So(b, should.NotBeNil) {
		var res TimeSyncResponse
		if err := json.Unmarshal(b, &res); !a.So(err, should.BeNil) {
			t.FailNow()
		}
		a.So(res.TxTime, should.Equal, 0.0)
		a.So(res.XTime, should.Equal, 0)
		a.So(res.GPSTime, should.Equal, 0)
		a.So(res.MuxTime, should.Equal, TimeToUnixSeconds(now))
	}

	// Add fictional session ID.
	updateSessionID(ctx, 0x42)

	gpsTime := time.Unix(456, 678)
	concentratorTime := scheduling.ConcentratorTime(890 * time.Microsecond)

	// Attempt to transfer time.
	b, err = f.TransferTime(ctx, now, &gpsTime, &concentratorTime)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	if a.So(b, should.NotBeNil) {
		var res TimeSyncResponse
		if err := json.Unmarshal(b, &res); !a.So(err, should.BeNil) {
			t.FailNow()
		}
		a.So(res.TxTime, should.Equal, 0.0)
		a.So(SessionIDFromXTime(res.XTime), should.Equal, 0x42)
		a.So(ConcentratorTimeFromXTime(res.XTime), should.Equal, 890*time.Microsecond)
		a.So(res.GPSTime, should.Equal, TimeToGPSTime(gpsTime))
		a.So(res.MuxTime, should.Equal, TimeToUnixSeconds(now))
	}
}
