package utils

import (
	"errors"
	"math"
)

// Geohash encodes latitude/longitude coordinates into a string
// for spatial indexing and proximity queries.

const (
	// Base32 alphabet for geohash encoding
	base32 = "0123456789bcdefghjkmnpqrstuvwxyz"

	// Default precision (8 characters = ~19m x 19m)
	DefaultPrecision = 8
)

var (
	ErrInvalidLatitude  = errors.New("latitude must be between -90 and 90")
	ErrInvalidLongitude = errors.New("longitude must be between -180 and 180")
	ErrInvalidPrecision = errors.New("precision must be between 1 and 12")
)

// Encode converts latitude and longitude to a geohash string
// precision determines the length of the resulting string (1-12)
//
// Precision levels:
//   1: ±2500 km
//   2: ±630 km
//   3: ±78 km
//   4: ±20 km
//   5: ±2.4 km
//   6: ±610 m
//   7: ±76 m
//   8: ±19 m (default)
//   9: ±2.4 m
//   10: ±60 cm
//   11: ±7.4 cm
//   12: ±1.9 cm
func Encode(lat, lon float64, precision int) (string, error) {
	if lat < -90 || lat > 90 {
		return "", ErrInvalidLatitude
	}
	if lon < -180 || lon > 180 {
		return "", ErrInvalidLongitude
	}
	if precision < 1 || precision > 12 {
		return "", ErrInvalidPrecision
	}

	latRange := [2]float64{-90, 90}
	lonRange := [2]float64{-180, 180}

	geohash := make([]byte, precision)
	bits := 0
	bit := 0
	ch := 0
	even := true

	for i := 0; i < precision; i++ {
		for bits < 5 {
			var mid float64
			if even {
				mid = (lonRange[0] + lonRange[1]) / 2
				if lon > mid {
					ch |= (1 << (4 - bits))
					lonRange[0] = mid
				} else {
					lonRange[1] = mid
				}
			} else {
				mid = (latRange[0] + latRange[1]) / 2
				if lat > mid {
					ch |= (1 << (4 - bits))
					latRange[0] = mid
				} else {
					latRange[1] = mid
				}
			}
			even = !even
			bits++
		}
		geohash[i] = base32[ch]
		bits = 0
		ch = 0
		bit = 0
	}

	return string(geohash), nil
}

// EncodeWithDefault encodes with default precision (8 characters)
func EncodeWithDefault(lat, lon float64) (string, error) {
	return Encode(lat, lon, DefaultPrecision)
}

// Decode converts a geohash string back to latitude and longitude
// Returns the center point of the geohash box
func Decode(geohash string) (lat, lon float64, err error) {
	if len(geohash) == 0 {
		return 0, 0, errors.New("geohash cannot be empty")
	}

	latRange := [2]float64{-90, 90}
	lonRange := [2]float64{-180, 180}
	even := true

	for _, char := range geohash {
		idx := indexInBase32(byte(char))
		if idx == -1 {
			return 0, 0, errors.New("invalid character in geohash")
		}

		for i := 4; i >= 0; i-- {
			bit := (idx >> i) & 1
			if even {
				mid := (lonRange[0] + lonRange[1]) / 2
				if bit == 1 {
					lonRange[0] = mid
				} else {
					lonRange[1] = mid
				}
			} else {
				mid := (latRange[0] + latRange[1]) / 2
				if bit == 1 {
					latRange[0] = mid
				} else {
					latRange[1] = mid
				}
			}
			even = !even
		}
	}

	lat = (latRange[0] + latRange[1]) / 2
	lon = (lonRange[0] + lonRange[1]) / 2
	return lat, lon, nil
}

// BoundingBox returns the lat/lon bounding box for a geohash
func BoundingBox(geohash string) (minLat, maxLat, minLon, maxLon float64, err error) {
	if len(geohash) == 0 {
		return 0, 0, 0, 0, errors.New("geohash cannot be empty")
	}

	latRange := [2]float64{-90, 90}
	lonRange := [2]float64{-180, 180}
	even := true

	for _, char := range geohash {
		idx := indexInBase32(byte(char))
		if idx == -1 {
			return 0, 0, 0, 0, errors.New("invalid character in geohash")
		}

		for i := 4; i >= 0; i-- {
			bit := (idx >> i) & 1
			if even {
				mid := (lonRange[0] + lonRange[1]) / 2
				if bit == 1 {
					lonRange[0] = mid
				} else {
					lonRange[1] = mid
				}
			} else {
				mid := (latRange[0] + latRange[1]) / 2
				if bit == 1 {
					latRange[0] = mid
				} else {
					latRange[1] = mid
				}
			}
			even = !even
		}
	}

	return latRange[0], latRange[1], lonRange[0], lonRange[1], nil
}

// Neighbors returns the 8 adjacent geohashes
func Neighbors(geohash string) ([]string, error) {
	if len(geohash) == 0 {
		return nil, errors.New("geohash cannot be empty")
	}

	lat, lon, err := Decode(geohash)
	if err != nil {
		return nil, err
	}

	minLat, maxLat, minLon, maxLon, err := BoundingBox(geohash)
	if err != nil {
		return nil, err
	}

	latStep := maxLat - minLat
	lonStep := maxLon - minLon

	precision := len(geohash)
	neighbors := make([]string, 0, 8)

	// North
	if n, err := Encode(lat+latStep, lon, precision); err == nil {
		neighbors = append(neighbors, n)
	}
	// South
	if n, err := Encode(lat-latStep, lon, precision); err == nil {
		neighbors = append(neighbors, n)
	}
	// East
	if n, err := Encode(lat, lon+lonStep, precision); err == nil {
		neighbors = append(neighbors, n)
	}
	// West
	if n, err := Encode(lat, lon-lonStep, precision); err == nil {
		neighbors = append(neighbors, n)
	}
	// Northeast
	if n, err := Encode(lat+latStep, lon+lonStep, precision); err == nil {
		neighbors = append(neighbors, n)
	}
	// Northwest
	if n, err := Encode(lat+latStep, lon-lonStep, precision); err == nil {
		neighbors = append(neighbors, n)
	}
	// Southeast
	if n, err := Encode(lat-latStep, lon+lonStep, precision); err == nil {
		neighbors = append(neighbors, n)
	}
	// Southwest
	if n, err := Encode(lat-latStep, lon-lonStep, precision); err == nil {
		neighbors = append(neighbors, n)
	}

	return neighbors, nil
}

// Distance calculates the approximate distance in kilometers between two geohashes
func Distance(geohash1, geohash2 string) (float64, error) {
	lat1, lon1, err := Decode(geohash1)
	if err != nil {
		return 0, err
	}

	lat2, lon2, err := Decode(geohash2)
	if err != nil {
		return 0, err
	}

	return haversineDistance(lat1, lon1, lat2, lon2), nil
}

// haversineDistance calculates distance between two points using the Haversine formula
func haversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadiusKm = 6371.0

	dLat := degToRad(lat2 - lat1)
	dLon := degToRad(lon2 - lon1)

	lat1Rad := degToRad(lat1)
	lat2Rad := degToRad(lat2)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Sin(dLon/2)*math.Sin(dLon/2)*math.Cos(lat1Rad)*math.Cos(lat2Rad)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusKm * c
}

// degToRad converts degrees to radians
func degToRad(deg float64) float64 {
	return deg * math.Pi / 180
}

// indexInBase32 returns the index of a character in the base32 alphabet
func indexInBase32(char byte) int {
	for i := 0; i < len(base32); i++ {
		if base32[i] == char {
			return i
		}
	}
	return -1
}

// PrecisionForRadius returns the appropriate geohash precision for a given radius in kilometers
func PrecisionForRadius(radiusKm float64) int {
	// Approximate geohash cell sizes at equator
	precisions := []struct {
		precision int
		sizeKm    float64
	}{
		{1, 5000},
		{2, 1250},
		{3, 156},
		{4, 39},
		{5, 4.9},
		{6, 1.2},
		{7, 0.15},
		{8, 0.038},
		{9, 0.0047},
		{10, 0.0012},
		{11, 0.00015},
		{12, 0.000037},
	}

	for _, p := range precisions {
		if radiusKm >= p.sizeKm {
			return p.precision
		}
	}

	return 12 // Maximum precision
}
