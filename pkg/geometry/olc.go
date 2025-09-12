package geometry

import (
	"fmt"
	"math"
	"sort"
	"strings"

	olc "github.com/google/open-location-code/go"
)

func EncodeOLC(lat, lng float64) string {
	olcPrefix := olc.Encode(lat, lng, 8)
	return olcPrefix[:8]
}

// GetOLCPrefixesInRadius returns OLC prefixes within radiusKms of the given lat/lng at codeLength.
func GetOLCPrefixesInRadius(lat, lng float64, radiusKms uint64) string {
	radiusMeters := float64(radiusKms * 1000)
	codeLength := getCodeLengthForRadius(radiusMeters)
	prefixes := GetOLCPrefixesInRadiusCustomCodeLength(lat, lng, codeLength, radiusMeters)

	return compressOLCPrefixes(prefixes)
}

func compressOLCPrefixes(prefixes []string) string {
	if len(prefixes) == 0 {
		return ""
	}

	// Sort prefixes to group similar ones together
	sort.Strings(prefixes)

	var patterns []string
	currentPrefix := prefixes[0][:len(prefixes[0])-1]
	currentChars := []byte{prefixes[0][len(prefixes[0])-1]}

	for i := 1; i < len(prefixes); i++ {
		prefix := prefixes[i][:len(prefixes[i])-1]
		lastChar := prefixes[i][len(prefixes[i])-1]

		if prefix == currentPrefix {
			currentChars = append(currentChars, lastChar)
		} else {
			patterns = append(patterns, formatPattern(currentPrefix, currentChars))
			currentPrefix = prefix
			currentChars = []byte{lastChar}
		}
	}

	patterns = append(patterns, formatPattern(currentPrefix, currentChars))

	return fmt.Sprintf("^(%s)", strings.Join(patterns, "|"))
}

func formatPattern(prefix string, chars []byte) string {
	if len(chars) == 1 {
		return prefix + string(chars[0])
	}
	return prefix + "[" + string(chars) + "]"
}

// GetOLCPrefixesInRadiusCustomCodeLength returns OLC prefixes within radiusMeters of the given lat/lng at codeLength.
func GetOLCPrefixesInRadiusCustomCodeLength(lat, lng float64, codeLength int, radiusMeters float64) []string {
	// Calculate cell size in degrees based on code length (OLC specification)
	cellSizeDeg := cellSizeInDegrees(codeLength)

	// Calculate latitude and longitude spans for the radius
	latSpan := radiusMeters / 111319.0                               // 1 degree ≈ 111,319 meters
	lonSpan := radiusMeters / (111319.0 * math.Cos(lat*math.Pi/180)) // Adjust for latitude

	// Number of cells to offset in each direction
	cellsLat := int(math.Ceil(latSpan / cellSizeDeg))
	cellsLon := int(math.Ceil(lonSpan / cellSizeDeg))

	var prefixes []string
	unique := make(map[string]struct{})

	// Iterate over all possible cell offsets
	for dlat := -cellsLat; dlat <= cellsLat; dlat++ {
		for dlon := -cellsLon; dlon <= cellsLon; dlon++ {
			// Calculate new coordinates
			newLat := lat + float64(dlat)*cellSizeDeg
			newLon := normalizeLongitude(lng + float64(dlon)*cellSizeDeg)

			// Skip invalid latitudes
			if newLat < -90 || newLat > 90 {
				continue
			}

			// Encode new OLC prefix
			code := olc.Encode(newLat, newLon, codeLength)
			prefix := code[:codeLength]

			// Check if already added
			if _, exists := unique[prefix]; !exists {
				// Calculate distance from original point
				distance := haversineDistance(lat, lng, newLat, newLon)
				if distance <= radiusMeters {
					unique[prefix] = struct{}{}
					prefixes = append(prefixes, prefix)
				}
			}
		}
	}

	return prefixes
}

func getCodeLengthForRadius(radiusMeters float64) int {
	switch {
	case radiusMeters > 100000: // > 100 km
		return 4
	case radiusMeters > 10000: // 10-100 km
		return 6
	case radiusMeters > 1000: // 1-10 km
		return 8
	default: // < 1 km
		return 10
	}
}

// cellSizeInDegrees returns the cell size in degrees for a given code length.
func cellSizeInDegrees(codeLength int) float64 {
	// OLC cell size per code length (from specification)
	switch {
	case codeLength <= 2:
		return 20.0
	case codeLength <= 4:
		return 1.0
	case codeLength <= 6:
		return 0.05
	case codeLength <= 8:
		return 0.0025
	case codeLength <= 10:
		return 0.000125
	default:
		return 0.000125 / math.Pow(20, float64(codeLength-10)/2)
	}
}

// normalizeLongitude wraps longitude to [-180, 180)
func normalizeLongitude(lon float64) float64 {
	lon = math.Mod(lon, 360)
	if lon < -180 {
		lon += 360
	} else if lon >= 180 {
		lon -= 360
	}
	return lon
}

// Helper function to calculate distance between two points using the Haversine formula
func haversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadius = 6371000 // Earth's radius in meters

	// Convert to radians
	lat1 = lat1 * math.Pi / 180
	lon1 = lon1 * math.Pi / 180
	lat2 = lat2 * math.Pi / 180
	lon2 = lon2 * math.Pi / 180

	dlat := lat2 - lat1
	dlon := lon2 - lon1

	a := math.Sin(dlat/2)*math.Sin(dlat/2) +
		math.Cos(lat1)*math.Cos(lat2)*
			math.Sin(dlon/2)*math.Sin(dlon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}
