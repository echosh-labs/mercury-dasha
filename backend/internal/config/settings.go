package config

import (
	"strings"
	"time"
)

// SystemSettings encapsulates user-configured or auto-detected temporal and geographic parameters.
type SystemSettings struct {
	Timezone       string    `json:"timezone"`         // IANA identifier, e.g. "America/New_York"
	UTCOffsetHours float64   `json:"utc_offset_hours"` // Hours offset from UTC (e.g. -4.0 for EDT)
	Latitude       float64   `json:"latitude"`         // Geographic latitude in decimal degrees
	Longitude      float64   `json:"longitude"`        // Geographic longitude in decimal degrees
	AutoDetect     bool      `json:"auto_detect"`      // True if client auto-detection is active
	LocationName   string    `json:"location_name"`    // Descriptive label, e.g. "New York, USA"
	UpdatedAt      time.Time `json:"updated_at"`
}

// DefaultSystemSettings returns the initial default configuration (St. Catharines / Niagara Region).
func DefaultSystemSettings() SystemSettings {
	return SystemSettings{
		Timezone:       "America/Toronto",
		UTCOffsetHours: -4.0,
		Latitude:       43.1594,
		Longitude:      -79.2469,
		AutoDetect:     false,
		LocationName:   "St. Catharines, ON",
		UpdatedAt:      time.Now().UTC(),
	}
}

type CentroidInfo struct {
	Latitude  float64
	Longitude float64
	Name      string
}

// IANACentroids maps standard world timezones to geographic coordinates.
var IANACentroids = map[string]CentroidInfo{
	"america/new_york":    {Latitude: 40.7128, Longitude: -74.0060, Name: "New York, USA"},
	"america/chicago":     {Latitude: 41.8781, Longitude: -87.6298, Name: "Chicago, USA"},
	"america/denver":      {Latitude: 39.7392, Longitude: -104.9903, Name: "Denver, USA"},
	"america/los_angeles": {Latitude: 34.0522, Longitude: -118.2437, Name: "Los Angeles, USA"},
	"america/phoenix":     {Latitude: 33.4484, Longitude: -112.0740, Name: "Phoenix, USA"},
	"america/toronto":     {Latitude: 43.1594, Longitude: -79.2469, Name: "St. Catharines / Niagara, Canada"},
	"america/vancouver":   {Latitude: 49.2827, Longitude: -123.1207, Name: "Vancouver, Canada"},
	"america/mexico_city": {Latitude: 19.4326, Longitude: -99.1332, Name: "Mexico City, Mexico"},
	"america/sao_paulo":   {Latitude: -23.5505, Longitude: -46.6333, Name: "São Paulo, Brazil"},
	"america/buenos_aires":{Latitude: -34.6037, Longitude: -58.3816, Name: "Buenos Aires, Argentina"},
	"europe/london":       {Latitude: 51.5074, Longitude: -0.1278, Name: "London, UK"},
	"europe/paris":        {Latitude: 48.8566, Longitude: 2.3522, Name: "Paris, France"},
	"europe/berlin":       {Latitude: 52.5200, Longitude: 13.4050, Name: "Berlin, Germany"},
	"europe/rome":         {Latitude: 41.9028, Longitude: 12.4964, Name: "Rome, Italy"},
	"europe/madrid":       {Latitude: 40.4168, Longitude: -3.7038, Name: "Madrid, Spain"},
	"europe/amsterdam":    {Latitude: 52.3676, Longitude: 4.9041, Name: "Amsterdam, Netherlands"},
	"europe/zurich":       {Latitude: 47.3769, Longitude: 8.5417, Name: "Zurich, Switzerland"},
	"europe/athens":       {Latitude: 37.9838, Longitude: 23.7275, Name: "Athens, Greece"},
	"asia/tokyo":          {Latitude: 35.6762, Longitude: 139.6503, Name: "Tokyo, Japan"},
	"asia/shanghai":       {Latitude: 31.2304, Longitude: 121.4737, Name: "Shanghai, China"},
	"asia/hong_kong":      {Latitude: 22.3193, Longitude: 114.1694, Name: "Hong Kong"},
	"asia/singapore":      {Latitude: 1.3521, Longitude: 103.8198, Name: "Singapore"},
	"asia/kolkata":        {Latitude: 28.6139, Longitude: 77.2090, Name: "New Delhi, India"},
	"asia/dubai":          {Latitude: 25.2048, Longitude: 55.2708, Name: "Dubai, UAE"},
	"asia/seoul":          {Latitude: 37.5665, Longitude: 126.9780, Name: "Seoul, South Korea"},
	"asia/bangkok":        {Latitude: 13.7563, Longitude: 100.5018, Name: "Bangkok, Thailand"},
	"australia/sydney":    {Latitude: -33.8688, Longitude: 151.2093, Name: "Sydney, Australia"},
	"australia/melbourne": {Latitude: -37.8136, Longitude: 144.9631, Name: "Melbourne, Australia"},
	"pacific/auckland":    {Latitude: -36.8485, Longitude: 174.7633, Name: "Auckland, New Zealand"},
	"pacific/honolulu":    {Latitude: 21.3069, Longitude: -157.8583, Name: "Honolulu, USA"},
	"utc":                 {Latitude: 0.0, Longitude: 0.0, Name: "Equatorial Prime Meridian (UTC)"},
}

// ResolveCentroid returns approximate coordinates for a given IANA timezone identifier.
func ResolveCentroid(tz string) (CentroidInfo, bool) {
	norm := strings.ToLower(strings.TrimSpace(tz))
	info, ok := IANACentroids[norm]
	return info, ok
}
