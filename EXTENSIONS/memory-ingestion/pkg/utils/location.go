package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// LocationEnricher provides location enrichment capabilities:
// - Reverse geocoding (lat/lon → place names)
// - Timezone lookup
// - Country/region detection
type LocationEnricher struct {
	httpClient      *http.Client
	cacheEnabled    bool
	cache           map[string]*EnrichedLocation
	nominatimURL    string // OpenStreetMap Nominatim API
	timezoneAPIKey  string // Optional: for timezone API
}

// EnrichedLocation contains enriched location data
type EnrichedLocation struct {
	Latitude    float64
	Longitude   float64
	PlaceName   string // Human-readable place name
	City        string
	State       string
	Country     string
	CountryCode string // ISO 3166-1 alpha-2
	Timezone    string // IANA timezone identifier
	Geohash     string
}

// NewLocationEnricher creates a new location enricher
func NewLocationEnricher(cacheEnabled bool) *LocationEnricher {
	return &LocationEnricher{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		cacheEnabled: cacheEnabled,
		cache:        make(map[string]*EnrichedLocation),
		// Using OpenStreetMap Nominatim (free, no API key required)
		nominatimURL: "https://nominatim.openstreetmap.org/reverse",
	}
}

// SetTimezoneAPIKey sets the API key for timezone lookups
// If not set, timezone will be estimated from coordinates
func (le *LocationEnricher) SetTimezoneAPIKey(apiKey string) {
	le.timezoneAPIKey = apiKey
}

// SetNominatimURL sets a custom Nominatim API URL
// Useful for self-hosted instances
func (le *LocationEnricher) SetNominatimURL(url string) {
	le.nominatimURL = url
}

// Enrich performs complete location enrichment
func (le *LocationEnricher) Enrich(lat, lon float64, precision int) (*EnrichedLocation, error) {
	if lat < -90 || lat > 90 {
		return nil, ErrInvalidLatitude
	}
	if lon < -180 || lon > 180 {
		return nil, ErrInvalidLongitude
	}

	// Check cache first
	cacheKey := fmt.Sprintf("%.6f,%.6f", lat, lon)
	if le.cacheEnabled {
		if cached, ok := le.cache[cacheKey]; ok {
			return cached, nil
		}
	}

	result := &EnrichedLocation{
		Latitude:  lat,
		Longitude: lon,
	}

	// Generate geohash
	geohash, err := Encode(lat, lon, precision)
	if err != nil {
		return nil, fmt.Errorf("failed to generate geohash: %w", err)
	}
	result.Geohash = geohash

	// Reverse geocode
	place, err := le.reverseGeocode(lat, lon)
	if err != nil {
		// Non-fatal: continue without place name
		result.PlaceName = fmt.Sprintf("%.6f, %.6f", lat, lon)
	} else {
		result.PlaceName = place.DisplayName
		result.City = place.Address.City
		if result.City == "" {
			result.City = place.Address.Town
		}
		if result.City == "" {
			result.City = place.Address.Village
		}
		result.State = place.Address.State
		result.Country = place.Address.Country
		result.CountryCode = place.Address.CountryCode
	}

	// Lookup timezone
	timezone, err := le.lookupTimezone(lat, lon)
	if err != nil {
		// Non-fatal: use estimated timezone
		timezone = estimateTimezone(lon)
	}
	result.Timezone = timezone

	// Cache result
	if le.cacheEnabled {
		le.cache[cacheKey] = result
	}

	return result, nil
}

// reverseGeocode converts coordinates to place information using Nominatim
func (le *LocationEnricher) reverseGeocode(lat, lon float64) (*nominatimResponse, error) {
	// Build request URL
	params := url.Values{}
	params.Set("lat", fmt.Sprintf("%.6f", lat))
	params.Set("lon", fmt.Sprintf("%.6f", lon))
	params.Set("format", "json")
	params.Set("addressdetails", "1")
	params.Set("zoom", "18") // Most detailed

	reqURL := fmt.Sprintf("%s?%s", le.nominatimURL, params.Encode())

	// Create request
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Nominatim requires User-Agent
	req.Header.Set("User-Agent", "Memory-Connector/1.0")

	// Execute request
	resp, err := le.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("nominatim returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var result nominatimResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// lookupTimezone looks up IANA timezone for coordinates
func (le *LocationEnricher) lookupTimezone(lat, lon float64) (string, error) {
	// If API key is set, use timezone API
	// Otherwise, estimate from longitude
	if le.timezoneAPIKey != "" {
		return le.lookupTimezoneAPI(lat, lon)
	}

	return estimateTimezone(lon), nil
}

// lookupTimezoneAPI queries a timezone API (placeholder for actual implementation)
// Suggested API: https://timezoneapi.io or https://timezonefinder API
func (le *LocationEnricher) lookupTimezoneAPI(lat, lon float64) (string, error) {
	// TODO: Implement actual API integration
	// For now, fall back to estimation
	return estimateTimezone(lon), errors.New("timezone API not implemented")
}

// estimateTimezone estimates timezone from longitude
// This is a rough approximation and may be incorrect near timezone boundaries
func estimateTimezone(lon float64) string {
	// Calculate UTC offset based on longitude
	// 15 degrees = 1 hour
	offset := int(lon / 15.0)

	// Map offset to common timezone
	// This is a simplification - real timezones are complex!
	switch {
	case offset >= -12 && offset < -8:
		return "America/Los_Angeles" // PST
	case offset >= -8 && offset < -6:
		return "America/Denver" // MST
	case offset >= -6 && offset < -5:
		return "America/Chicago" // CST
	case offset >= -5 && offset <= 0:
		return "America/New_York" // EST
	case offset > 0 && offset <= 2:
		return "Europe/Paris" // CET
	case offset > 2 && offset <= 4:
		return "Europe/Moscow" // MSK
	case offset > 4 && offset <= 6:
		return "Asia/Kolkata" // IST
	case offset > 6 && offset <= 9:
		return "Asia/Shanghai" // CST
	case offset > 9 && offset <= 12:
		return "Asia/Tokyo" // JST
	default:
		return "UTC"
	}
}

// nominatimResponse matches the Nominatim API response structure
type nominatimResponse struct {
	PlaceID     int             `json:"place_id"`
	Licence     string          `json:"licence"`
	DisplayName string          `json:"display_name"`
	Address     nominatimAddress `json:"address"`
	Lat         string          `json:"lat"`
	Lon         string          `json:"lon"`
}

type nominatimAddress struct {
	HouseNumber string `json:"house_number"`
	Road        string `json:"road"`
	Suburb      string `json:"suburb"`
	Village     string `json:"village"`
	Town        string `json:"town"`
	City        string `json:"city"`
	County      string `json:"county"`
	State       string `json:"state"`
	Postcode    string `json:"postcode"`
	Country     string `json:"country"`
	CountryCode string `json:"country_code"`
}

// BuildPlaceName creates a human-readable place name from components
func BuildPlaceName(city, state, country string) string {
	parts := []string{}

	if city != "" {
		parts = append(parts, city)
	}
	if state != "" {
		parts = append(parts, state)
	}
	if country != "" {
		parts = append(parts, country)
	}

	if len(parts) == 0 {
		return "Unknown Location"
	}

	result := parts[0]
	for i := 1; i < len(parts); i++ {
		result += ", " + parts[i]
	}

	return result
}

// GetCacheStats returns cache statistics
func (le *LocationEnricher) GetCacheStats() (size int, enabled bool) {
	return len(le.cache), le.cacheEnabled
}

// ClearCache clears the location cache
func (le *LocationEnricher) ClearCache() {
	le.cache = make(map[string]*EnrichedLocation)
}
