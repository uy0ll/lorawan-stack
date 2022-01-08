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

package band

import "go.thethings.network/lorawan-stack/v3/pkg/ttnpb"

// CN_470_510_26_A_RP2_v1_0_0 is the band definition for CN470-510 26MHz antenna, type A in the RP002-1.0.0 specification.
var CN_470_510_26_A_RP2_v1_0_0 = Band{
	ID: CN_470_510_26_A,

	EnableADR: true,

	MaxUplinkChannels: 48,
	UplinkChannels:    cn47051026AUplinkChannels,

	MaxDownlinkChannels: 24,
	DownlinkChannels:    cn47051026ADownlinkChannels,

	// See IEEE 11-11/0972r0
	SubBands: []SubBandParameters{
		{
			MinFrequency: 470000000,
			MaxFrequency: 510000000,
			DutyCycle:    1,
			MaxEIRP:      17.0 + eirpDelta,
		},
	},

	DataRates: map[ttnpb.DataRateIndex]DataRate{
		ttnpb.DATA_RATE_0: makeLoRaDataRate(12, 125000, makeConstMaxMACPayloadSizeFunc(59)),
		ttnpb.DATA_RATE_1: makeLoRaDataRate(11, 125000, makeConstMaxMACPayloadSizeFunc(59)),
		ttnpb.DATA_RATE_2: makeLoRaDataRate(10, 125000, makeConstMaxMACPayloadSizeFunc(59)),
		ttnpb.DATA_RATE_3: makeLoRaDataRate(9, 125000, makeConstMaxMACPayloadSizeFunc(123)),
		ttnpb.DATA_RATE_4: makeLoRaDataRate(8, 125000, makeConstMaxMACPayloadSizeFunc(230)),
		ttnpb.DATA_RATE_5: makeLoRaDataRate(7, 125000, makeConstMaxMACPayloadSizeFunc(230)),
	},
	MaxADRDataRateIndex: ttnpb.DATA_RATE_5,

	ReceiveDelay1:        defaultReceiveDelay1,
	ReceiveDelay2:        defaultReceiveDelay2,
	JoinAcceptDelay1:     defaultJoinAcceptDelay1,
	JoinAcceptDelay2:     defaultJoinAcceptDelay2,
	MaxFCntGap:           defaultMaxFCntGap,
	ADRAckLimit:          defaultADRAckLimit,
	ADRAckDelay:          defaultADRAckDelay,
	MinRetransmitTimeout: defaultRetransmitTimeout - defaultRetransmitTimeoutMargin,
	MaxRetransmitTimeout: defaultRetransmitTimeout + defaultRetransmitTimeoutMargin,

	DefaultMaxEIRP: 19.15,
	TxOffset: []float32{
		0,
		-2,
		-4,
		-6,
		-8,
		-10,
		-12,
		-14,
	},

	Rx1Channel: channelIndexModulo(24),
	Rx1DataRate: func(idx ttnpb.DataRateIndex, offset ttnpb.DataRateOffset, _ bool) (ttnpb.DataRateIndex, error) {
		if idx > ttnpb.DATA_RATE_5 {
			return 0, errDataRateIndexTooHigh.WithAttributes("max", 5)
		}
		if offset > 5 {
			return 0, errDataRateOffsetTooHigh.WithAttributes("max", 5)
		}
		// Unchanged from the pre-RP2 CN470-510 band definition.
		return cn470510DownlinkDRTable[idx][offset], nil
	},

	GenerateChMasks: generateChMask48,
	ParseChMask:     parseChMask48,

	DefaultRx2Parameters: Rx2Parameters{ttnpb.DATA_RATE_0, 492500000},

	Beacon: Beacon{
		DataRateIndex: ttnpb.DATA_RATE_2,
		CodingRate:    "4/5",
		// The frequency of the beacon depends on the Common Join Channel Index,
		// which is not available at band level.
		ComputeFrequency: nil,
	},

	LoRaCodingRate: "4/5",

	FreqMultiplier:   100,
	ImplementsCFList: true,
	CFListType:       ttnpb.CFListType_CHANNEL_MASKS,
}
