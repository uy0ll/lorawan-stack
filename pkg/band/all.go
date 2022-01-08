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

import (
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// All contains all the bands available.
var All = map[string]map[ttnpb.PHYVersion]Band{
	AS_923: {
		ttnpb.RP001_V1_0_2:       AS_923_RP1_v1_0_2,
		ttnpb.RP001_V1_0_2_REV_B: AS_923_RP1_v1_0_2_RevB,
		ttnpb.RP001_V1_0_3_REV_A: AS_923_RP1_v1_0_3_RevA,
		ttnpb.RP001_V1_1_REV_A:   AS_923_RP1_v1_1_RevA,
		ttnpb.RP001_V1_1_REV_B:   AS_923_RP1_v1_1_RevB,
		ttnpb.RP002_V1_0_0:       AS_923_RP2_v1_0_0,
	},
	AU_915_928: {
		ttnpb.TS001_V1_0_1:       AU_915_928_TS1_v1_0_1,
		ttnpb.RP001_V1_0_2:       AU_915_928_RP1_v1_0_2,
		ttnpb.RP001_V1_0_2_REV_B: AU_915_928_RP1_v1_0_2_RevB,
		ttnpb.RP001_V1_0_3_REV_A: AU_915_928_RP1_v1_0_3_RevA,
		ttnpb.RP001_V1_1_REV_A:   AU_915_928_RP1_v1_1_RevA,
		ttnpb.RP001_V1_1_REV_B:   AU_915_928_RP1_v1_1_RevB,
		ttnpb.RP002_V1_0_0:       AU_915_928_RP2_v1_0_0,
	},
	CN_470_510: {
		ttnpb.TS001_V1_0_1:       CN_470_510_TS1_v1_0_1,
		ttnpb.RP001_V1_0_2:       CN_470_510_RP1_v1_0_2,
		ttnpb.RP001_V1_0_2_REV_B: CN_470_510_RP1_v1_0_2_RevB,
		ttnpb.RP001_V1_0_3_REV_A: CN_470_510_RP1_v1_0_3_RevA,
		ttnpb.RP001_V1_1_REV_A:   CN_470_510_RP1_v1_1_RevA,
		ttnpb.RP001_V1_1_REV_B:   CN_470_510_RP1_v1_1_RevB,
	},
	CN_470_510_20_A: {
		ttnpb.RP002_V1_0_0: CN_470_510_20_A_RP2_v1_0_0,
	},
	CN_470_510_20_B: {
		ttnpb.RP002_V1_0_0: CN_470_510_20_B_RP2_v1_0_0,
	},
	CN_470_510_26_A: {
		ttnpb.RP002_V1_0_0: CN_470_510_26_A_RP2_v1_0_0,
	},
	CN_470_510_26_B: {
		ttnpb.RP002_V1_0_0: CN_470_510_26_B_RP2_v1_0_0,
	},
	CN_779_787: {
		ttnpb.TS001_V1_0:         CN_779_787_RP1_V1_0,
		ttnpb.TS001_V1_0_1:       CN_779_787_RP1_V1_0_1,
		ttnpb.RP001_V1_0_2:       CN_779_787_RP1_V1_0_2,
		ttnpb.RP001_V1_0_2_REV_B: CN_779_787_RP1_V1_0_2_RevB,
		ttnpb.RP001_V1_0_3_REV_A: CN_779_787_RP1_V1_0_3_RevA,
		ttnpb.RP001_V1_1_REV_A:   CN_779_787_RP1_V1_1_RevA,
		ttnpb.RP001_V1_1_REV_B:   CN_779_787_RP1_V1_1_RevB,
		ttnpb.RP002_V1_0_0:       CN_779_787_RP2_V1_0_0,
	},
	EU_433: {
		ttnpb.TS001_V1_0:         EU_433_TS1_V1_0,
		ttnpb.TS001_V1_0_1:       EU_433_TS1_V1_0_1,
		ttnpb.RP001_V1_0_2:       EU_433_RP1_V1_0_2,
		ttnpb.RP001_V1_0_2_REV_B: EU_433_RP1_V1_0_2_Rev_B,
		ttnpb.RP001_V1_0_3_REV_A: EU_433_RP1_V1_0_3_Rev_A,
		ttnpb.RP001_V1_1_REV_A:   EU_433_RP1_V1_1_Rev_A,
		ttnpb.RP001_V1_1_REV_B:   EU_433_RP1_V1_1_Rev_B,
		ttnpb.RP002_V1_0_0:       EU_433_RP2_V1_0_0,
	},
	EU_863_870: {
		ttnpb.TS001_V1_0:         EU_863_870_TS1_V1_0,
		ttnpb.TS001_V1_0_1:       EU_863_870_TS1_V1_0_1,
		ttnpb.RP001_V1_0_2:       EU_863_870_RP1_V1_0_2,
		ttnpb.RP001_V1_0_2_REV_B: EU_863_870_RP1_V1_0_2_Rev_B,
		ttnpb.RP001_V1_0_3_REV_A: EU_863_870_RP1_V1_0_3_Rev_A,
		ttnpb.RP001_V1_1_REV_A:   EU_863_870_RP1_V1_1_Rev_A,
		ttnpb.RP001_V1_1_REV_B:   EU_863_870_RP1_V1_1_Rev_B,
		ttnpb.RP002_V1_0_0:       EU_863_870_RP2_V1_0_0,
	},
	IN_865_867: {
		ttnpb.RP001_V1_0_2_REV_B: IN_865_867_RP1_V1_0_2_Rev_B,
		ttnpb.RP001_V1_0_3_REV_A: IN_865_867_RP1_V1_0_3_Rev_A,
		ttnpb.RP001_V1_1_REV_A:   IN_865_867_RP1_V1_1_Rev_A,
		ttnpb.RP001_V1_1_REV_B:   IN_865_867_RP1_V1_1_Rev_B,
		ttnpb.RP002_V1_0_0:       IN_865_867_RP2_V1_0_0,
	},
	ISM_2400: {
		ttnpb.TS001_V1_0:         ISM_2400_Universal,
		ttnpb.TS001_V1_0_1:       ISM_2400_Universal,
		ttnpb.RP001_V1_0_2:       ISM_2400_Universal,
		ttnpb.RP001_V1_0_2_REV_B: ISM_2400_Universal,
		ttnpb.RP001_V1_0_3_REV_A: ISM_2400_Universal,
		ttnpb.RP001_V1_1_REV_A:   ISM_2400_Universal,
		ttnpb.RP001_V1_1_REV_B:   ISM_2400_Universal,
		ttnpb.RP002_V1_0_0:       ISM_2400_Universal,
	},
	KR_920_923: {
		ttnpb.RP001_V1_0_2:       KR_920_923_RP1_V1_0_2,
		ttnpb.RP001_V1_0_2_REV_B: KR_920_923_RP1_V1_0_2_Rev_B,
		ttnpb.RP001_V1_0_3_REV_A: KR_920_923_RP1_V1_0_3_Rev_A,
		ttnpb.RP001_V1_1_REV_A:   KR_920_923_RP1_V1_1_Rev_A,
		ttnpb.RP001_V1_1_REV_B:   KR_920_923_RP1_V1_1_Rev_B,
		ttnpb.RP002_V1_0_0:       KR_920_923_RP2_V1_0_0,
	},
	RU_864_870: {
		ttnpb.RP001_V1_0_3_REV_A: RU_864_870_RP1_V1_0_3_Rev_A,
		ttnpb.RP001_V1_1_REV_A:   RU_864_870_RP1_V1_1_Rev_A,
		ttnpb.RP001_V1_1_REV_B:   RU_864_870_RP1_V1_1_Rev_B,
		ttnpb.RP002_V1_0_0:       RU_864_870_RP2_V1_0_0,
	},
	US_902_928: {
		ttnpb.TS001_V1_0:         US_902_928_TS1_V1_0,
		ttnpb.TS001_V1_0_1:       US_902_928_TS1_V1_0_1,
		ttnpb.RP001_V1_0_2:       US_902_928_RP1_V1_0_2,
		ttnpb.RP001_V1_0_2_REV_B: US_902_928_RP1_V1_0_2_Rev_B,
		ttnpb.RP001_V1_0_3_REV_A: US_902_928_RP1_V1_0_3_Rev_A,
		ttnpb.RP001_V1_1_REV_A:   US_902_928_RP1_V1_1_Rev_A,
		ttnpb.RP001_V1_1_REV_B:   US_902_928_RP1_V1_1_Rev_B,
		ttnpb.RP002_V1_0_0:       US_902_928_RP2_V1_0_0,
	},
}

// Get returns the band if it was found, and returns an error otherwise.
func Get(id string, version ttnpb.PHYVersion) (Band, error) {
	versions, ok := All[id]
	if !ok {
		return Band{}, errBandNotFound.WithAttributes("id", id, "version", version)
	}
	band, ok := versions[version]
	if !ok {
		return Band{}, errBandNotFound.WithAttributes("id", id, "version", version)
	}
	return band, nil
}

const latestSupportedVersion = ttnpb.RP001_V1_1_REV_B

// GetLatest returns the latest version of the band if it was found,
// and returns an error otherwise.
func GetLatest(id string) (Band, error) {
	return Get(id, latestSupportedVersion)
}
