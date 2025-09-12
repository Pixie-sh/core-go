package geometry

import (
	"reflect"
	"testing"
)

func TestGeohash_String(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{
			name:  "basic string representation",
			value: "u4pruydqqvj",
			want:  "u4pruydqqvj",
		},
		{
			name:  "empty string",
			value: "",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := Geohash{
				value: tt.value,
			}
			if got := g.String(); got != tt.want {
				t.Errorf("Geohash.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGeohash_Precision(t *testing.T) {
	tests := []struct {
		name      string
		precision uint
		want      uint
	}{
		{
			name:      "default precision",
			precision: 6,
			want:      6,
		},
		{
			name:      "custom precision",
			precision: 12,
			want:      12,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := Geohash{
				precision: tt.precision,
			}
			if got := g.Precision(); got != tt.want {
				t.Errorf("Geohash.Precision() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGeohash_LatLon(t *testing.T) {
	tests := []struct {
		name string
		lat  float64
		lon  float64
		want []float64
	}{
		{
			name: "San Francisco",
			lat:  37.7749,
			lon:  -122.4194,
			want: []float64{37.7749, -122.4194},
		},
		{
			name: "New York",
			lat:  40.7128,
			lon:  -74.0060,
			want: []float64{40.7128, -74.0060},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := Geohash{
				lat: tt.lat,
				lon: tt.lon,
			}
			if lat, log := g.LatLon(); !reflect.DeepEqual([]float64{lat, log}, tt.want) {
				t.Errorf("Geohash.LatLon() = %v, want %v", []float64{lat, log}, tt.want)
			}
		})
	}
}

func TestGeohash_FromLatLon(t *testing.T) {
	tests := []struct {
		name           string
		lat            float64
		lon            float64
		precision      uint
		expectedValue  string
		expectedLat    float64
		expectedLon    float64
		expectedPrecis uint
	}{
		{
			name:           "San Francisco default precision",
			lat:            37.7749,
			lon:            -122.4194,
			precision:      0,        // Use default
			expectedValue:  "9q8yyk", // This is the actual geohash for SF with precision 6
			expectedLat:    37.7749,
			expectedLon:    -122.4194,
			expectedPrecis: 0, // The method doesn't update this field
		},
		{
			name:           "New York with custom precision",
			lat:            40.7128,
			lon:            -74.0060,
			precision:      8,
			expectedValue:  "dr5regw3", // This is the actual geohash for NYC with precision 8
			expectedLat:    40.7128,
			expectedLon:    -74.0060,
			expectedPrecis: 0, // The method doesn't update this field
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Geohash{}
			var result *Geohash

			if tt.precision == 0 {
				result = g.FromLatLon(tt.lat, tt.lon)
			} else {
				result = g.FromLatLon(tt.lat, tt.lon, tt.precision)
			}

			// Check the returned result
			if result.value != tt.expectedValue {
				t.Errorf("Geohash.FromLatLon() value = %v, want %v", result.value, tt.expectedValue)
			}

			// Check the modified original struct
			if g.value != tt.expectedValue {
				t.Errorf("Geohash.value after FromLatLon() = %v, want %v", g.value, tt.expectedValue)
			}

			if g.lat != tt.expectedLat {
				t.Errorf("Geohash.lat after FromLatLon() = %v, want %v", g.lat, tt.expectedLat)
			}

			if g.lon != tt.expectedLon {
				t.Errorf("Geohash.lon after FromLatLon() = %v, want %v", g.lon, tt.expectedLon)
			}
		})
	}
}

func TestGeohash_From(t *testing.T) {
	tests := []struct {
		name     string
		source   Geohash
		expected Geohash
	}{
		{
			name: "Copy from another geohash",
			source: Geohash{
				value:     "9q8yyk",
				precision: 6,
				lat:       37.7749,
				lon:       -122.4194,
			},
			expected: Geohash{
				value:     "9q8yyk",
				precision: 6,
				lat:       37.7749,
				lon:       -122.4194,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Geohash{}
			result := g.From(tt.source)

			// Check the returned result
			if !reflect.DeepEqual(*result, tt.expected) {
				t.Errorf("Geohash.From() result = %v, want %v", result, tt.expected)
			}

			// Check the modified original struct
			if !reflect.DeepEqual(*g, tt.expected) {
				t.Errorf("Geohash after From() = %v, want %v", *g, tt.expected)
			}
		})
	}
}

func TestGeohash_PrecisionVariations(t *testing.T) {
	tests := []struct {
		name     string
		geohash  Geohash
		expected []Geohash
	}{
		{
			name: "3-character geohash",
			geohash: Geohash{
				value:     "9q8",
				precision: 3,
			},
			expected: []Geohash{
				{value: "9q8", precision: 3},
				{value: "9q", precision: 2},
			},
		},
		{
			name: "5-character geohash",
			geohash: Geohash{
				value:     "9q8yy",
				precision: 5,
			},
			expected: []Geohash{
				{value: "9q8yy", precision: 5},
				{value: "9q8y", precision: 4},
				{value: "9q8", precision: 3},
				{value: "9q", precision: 2},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.geohash.PrecisionVariations()

			if len(result) != len(tt.expected) {
				t.Errorf("Geohash.PrecisionVariations() length = %v, want %v", len(result), len(tt.expected))
				return
			}

			for i, v := range result {
				if v.value != tt.expected[i].value || v.precision != tt.expected[i].precision {
					t.Errorf("Geohash.PrecisionVariations()[%d] = {value: %v, precision: %v}, want {value: %v, precision: %v}",
						i, v.value, v.precision, tt.expected[i].value, tt.expected[i].precision)
				}
			}
		})
	}
}

func Test_convertToGeohash(t *testing.T) {
	tests := []struct {
		name      string
		lat       float64
		lon       float64
		precision uint
		want      string
	}{
		{
			name:      "San Francisco precision 6",
			lat:       37.7749,
			lon:       -122.4194,
			precision: 6,
			want:      "9q8yyk",
		},
		{
			name:      "New York precision 8",
			lat:       40.7128,
			lon:       -74.0060,
			precision: 8,
			want:      "dr5regw3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := geohashFromLatLon(tt.lat, tt.lon, tt.precision); got != tt.want {
				t.Errorf("convertToGeohash() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_geohashPrecisionVariations(t *testing.T) {
	tests := []struct {
		name    string
		geohash string
		want    []string
	}{
		{
			name:    "5-character geohash",
			geohash: "9q8yy",
			want:    []string{"9q8yy", "9q8y", "9q8", "9q"},
		},
		{
			name:    "3-character geohash",
			geohash: "9q8",
			want:    []string{"9q8", "9q"},
		},
		{
			name:    "1-character geohash",
			geohash: "9",
			want:    []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := geohashPrecisionVariations(tt.geohash); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("geohashPrecisionVariations() = %v, want %v", got, tt.want)
			}
		})
	}
}
