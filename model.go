// Copyright 2026 Dean Stalker
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package adsbexchange

// Config represents the configuration for the ADS-B Exchange feed
type Config struct {
	Host string `json:"host" mapstructure:"host"`
	Key  string `json:"key" mapstructure:"key"`
}

// Result represents the result of a feed query
type Result struct {
	Aircraft []*Aircraft `json:"ac"`
	Message  string      `json:"msg"`
	Now      int64       `json:"now"`
	Total    int64       `json:"total"`
	CTime    int64       `json:"ctime"`
	PTime    int64       `json:"ptime"`
}

// Aircraft represents an aircraft in the feed
type Aircraft struct {
	Hex          string           `json:"hex"`
	RecordType   RecordType       `json:"type"`
	Flight       string           `json:"flight"`
	Registration string           `json:"r"`
	Type         string           `json:"t"`
	Emergency    *Emergency       `json:"emergency"`
	Category     *EmitterCategory `json:"category"`

	DBFlags int64 `json:"dbFlags"`

	AltimeterBarometric float64 `json:"alt_baro"`
	AltimeterGeometric  float64 `json:"alt_geom"`

	GroundSpeed       float64  `json:"gs"`
	IndicatedAirSpeed *float64 `json:"ias"`
	TrueAirSpeed      *float64 `json:"tas"`
	Mach              *float64 `json:"mach"`

	WindDirection  *float64 `json:"wd"`
	WindSpeed      *float64 `json:"ws"`
	OutsideAirTemp *float64 `json:"oat"`
	TotalAirTemp   *float64 `json:"tat"`

	Track           *float64 `json:"track"`
	TrackRate       *float64 `json:"track_rate"`
	Roll            *float64 `json:"roll"`
	MagneticHeading *float64 `json:"mag_heading"`
	TrueHeading     *float64 `json:"true_heading"`
	BarometricRate  *float64 `json:"baro_rate"`
	GeometricRate   *float64 `json:"geom_rate"`

	NavQNH         *float64   `json:"nav_qnh"`
	NavAltitudeMCP *float64   `json:"nav_altitude_mcp"`
	NavAltitudeFMS *float64   `json:"nav_altitude_fms"`
	NavHeading     *float64   `json:"nav_heading"`
	NavModes       []*NavMode `json:"nav_modes"`
	Squawk         *string    `json:"squawk"`

	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lon"`

	NavigationIntegrityCategory     int64   `json:"nic"`
	RadiusContainment               int64   `json:"rc"`
	LastUpdatedPosition             float64 `json:"seen_pos"`
	Version                         int64   `json:"version"`
	NavigationIntegrityCategoryBaro int64   `json:"nic_baro"`
	NavigationAccuracyPosition      int64   `json:"nac_p"`
	NavigationAccuracyVertical      int64   `json:"nac_v"`
	SourceIntegrityLevel            int64   `json:"sil"`
	SourceIntegrityLevelType        string  `json:"sil_type"`
	GeometricVerticalAccuracy       int64   `json:"gva"`
	SystemDesignAssurance           int64   `json:"sda"`
	FlightStatusAlertBit            int64   `json:"alert"`

	MLAT []*string `json:"mlat"`
	TISB []*string `json:"tisb"`

	Messages           int64   `json:"messages"`
	LastUpdatedSeconds float64 `json:"seen"`
	RSSI               float64 `json:"rssi"`
}

type Emergency string
type EmitterCategory string
type RecordType string

type NavMode string

const (
	EmergencyNone      Emergency = "none"
	EmergencyGeneral   Emergency = "general"
	EmergencyLifeguard Emergency = "lifeguard"
	EmergencyMinfuel   Emergency = "minfuel"
	EmergencyNordo     Emergency = "nordo"
	EmergencyUnlawful  Emergency = "unlawful"
	EmergencyDowned    Emergency = "downed"
	EmergencyReserved  Emergency = "reserved"

	NavModeAltHold   NavMode = "althold"
	NavModeApproach  NavMode = "approach"
	NavModeAutopilot NavMode = "autopilot"
	NavModeLNav      NavMode = "lnav"
	NavModeTCAS      NavMode = "tcas"
	NavModeVNAV      NavMode = "vnav"

	RecordTypeADSBICAO      RecordType = "adsb_icao"      // messages from a Mode S or ADS-B transponder, using a 24-bit ICAO address
	RecordTypeADSBICAONT    RecordType = "adsb_icao_nt"   // messages from an ADS-B equipped “non-transponder” emitter e.g., a ground vehicle, using a 24-bit ICAO address
	RecordTypeADSRICAO      RecordType = "adsr_icao"      // rebroadcast of ADS-B messages originally sent via another data link e.g., UAT, using a 24-bit ICAO address
	RecordTypeTISBICAO      RecordType = "tisb_icao"      // traffic information about a non-ADS-B target identified by a 24-bit ICAO address, e.g., a Mode S target tracked by secondary radar
	RecordTypeADSC          RecordType = "adsc"           // ADS-C (received by monitoring satellite downlinks)
	RecordTypeMLAT          RecordType = "mlat"           // MLAT, position-calculated arrival time differences using multiple receivers, outliers, and varying accuracy are expected.
	RecordTypeOther         RecordType = "other"          // miscellaneous data received via Basestation / SBS format, quality / source is unknown.
	RecordTypeModeS         RecordType = "mode_s"         // ModeS data from the plane transponder (no position transmitted)
	RecordTypeADSBOther     RecordType = "adsb_other"     // messages from an ADS-B transponder using a non-ICAO address, e.g., anonymized address
	RecordTypeADSROther     RecordType = "adsr_other"     // rebroadcast of ADS-B messages originally sent via another data link e.g., UAT, using a non-ICAO address
	RecordTypeTISBOther     RecordType = "tisb_other"     // traffic information about a non-ADS-B target identified by a non-ICAO address, e.g., a Mode S target tracked by secondary radar
	RecordTypeTISBTrackFile RecordType = "tisb_trackfile" // traffic information about a non-ADS-B target using a track/file identifier, typically from primary or Mode A/C radar

	CategoryNone            EmitterCategory = "A0" // No ADS-B Emitter category
	CategoryLight           EmitterCategory = "A1" // < 15,500 lbs
	CategorySmall           EmitterCategory = "A2" // 15,500 - 75,000 lbs
	CategoryLarge           EmitterCategory = "A3" // 75,000 - 300,000 lbs
	CategoryHVLarge         EmitterCategory = "A4" // High wake vortex aircraft >= 75,000 lbs <= 300,000 lbs
	CategoryHeavy           EmitterCategory = "A5" // > 300,000 lbs
	CategoryHighPerformance EmitterCategory = "A6" // High performance aircraft
	CategoryRotorcraft      EmitterCategory = "A7" // Rotorcraft
	CategoryGlider          EmitterCategory = "B1" // Glider / Sailplane
	CategoryLighterThanAir  EmitterCategory = "B2" //
	CategoryParachutist     EmitterCategory = "B3" // Parachutist / Skydiver
	CategoryUltraLight      EmitterCategory = "B4" // Ultra-light aircraft a vehicle that meets the requirements of 14 CFR 103.1
	CategoryUAV             EmitterCategory = "B6" // Unmanned Aerial Vehicle
	CategorySpacecraft      EmitterCategory = "B7" // Spacecraft
)

var CategoryDescription = map[EmitterCategory]string{
	CategoryNone:            "No ADS-B Emitter category information",
	CategoryLight:           "Light aircraft",
	CategorySmall:           "Small aircraft",
	CategoryLarge:           "Large aircraft",
	CategoryHVLarge:         "High wake vortex aircraft",
	CategoryHeavy:           "Heavy aircraft",
	CategoryHighPerformance: "High Performance aircraft",
	CategoryRotorcraft:      "Rotorcraft",
	CategoryGlider:          "Glider",
	CategoryLighterThanAir:  "Lighter than air",
	CategoryParachutist:     "Parachutist / Skydiver",
	CategoryUltraLight:      "Ultra-light aircraft a vehicle that meets the requirements of 14 CFR 103.1",
	CategoryUAV:             "Unmanned Aerial Vehicle",
	CategorySpacecraft:      "Spacecraft",
}
