package utils

import (
	"testing"
	"time"
)

func TestFormatDuration(t *testing.T) {
	testCases := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{"Zero", 0, "0s"},
		{"OnlySeconds", 45 * time.Second, "45s"},
		{"OnlyMinutes", 5 * time.Minute, "5m"},
		{"OnlyHours", 3 * time.Hour, "3h"},
		{"OnlyDays", 2 * 24 * time.Hour, "2d"},
		{"MinutesAndSeconds", 1*time.Minute + 30*time.Second, "1m 30s"},
		{"HoursAndMinutes", 2*time.Hour + 45*time.Minute, "2h 45m"},
		{"HoursMinutesSeconds", 1*time.Hour + 30*time.Minute + 15*time.Second, "1h 30m 15s"},
		{"DaysAndHours", 3*24*time.Hour + 5*time.Hour, "3d 5h"},
		{"DaysHoursMinutes", 1*24*time.Hour + 12*time.Hour + 30*time.Minute, "1d 12h 30m"},
		{"DaysHoursMinutesSeconds", 1*24*time.Hour + 23*time.Hour + 59*time.Minute + 59*time.Second, "1d 23h 59m 59s"},
		{"LargeNumber", 365*24*time.Hour + 5*time.Hour + 48*time.Minute + 56*time.Second, "365d 5h 48m 56s"},
		{"SingleUnits", 1*24*time.Hour + 1*time.Hour + 1*time.Minute + 1*time.Second, "1d 1h 1m 1s"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := FormatDuration(tc.duration)
			if result != tc.expected {
				t.Errorf("FormatDuration(%v) = %s; want %s", tc.duration, result, tc.expected)
			}
		})
	}
}

func TestOrderStringSlices(t *testing.T) {
	testCases := []struct {
		name      string
		source    []string
		reference []string
		expected  []string
	}{
		{
			name:      "Basic ordering with some items not in reference",
			source:    []string{"test2", "test3", "test1"},
			reference: []string{"test1", "test3"},
			expected:  []string{"test1", "test3", "test2"},
		},
		{
			name:      "All source items in reference",
			source:    []string{"test2", "test3", "test1"},
			reference: []string{"test1", "test3", "test2"},
			expected:  []string{"test1", "test3", "test2"},
		},
		{
			name:      "Some reference items not in source",
			source:    []string{"test2", "test3", "test1"},
			reference: []string{"test4", "test1", "test3", "test5"},
			expected:  []string{"test1", "test3", "test2"},
		},
		{
			name:      "Empty source",
			source:    []string{},
			reference: []string{"test1", "test2"},
			expected:  []string{},
		},
		{
			name:      "Empty reference",
			source:    []string{"test1", "test2", "test3"},
			reference: []string{},
			expected:  []string{"test1", "test2", "test3"},
		},
		{
			name:      "Duplicate items in source",
			source:    []string{"test1", "test2", "test1", "test3"},
			reference: []string{"test3", "test1"},
			expected:  []string{"test3", "test1", "test2", "test1"},
		},
		{
			name:      "Duplicate items in reference",
			source:    []string{"test1", "test2", "test3"},
			reference: []string{"test1", "test1", "test3"},
			expected:  []string{"test1", "test3", "test2"},
		},
		{
			name:      "No common items",
			source:    []string{"test1", "test2", "test3"},
			reference: []string{"test4", "test5", "test6"},
			expected:  []string{"test1", "test2", "test3"},
		},
		{
			name:      "Case sensitivity",
			source:    []string{"Test", "test", "TEST"},
			reference: []string{"test", "TEST"},
			expected:  []string{"test", "TEST", "Test"},
		},
		{
			name:      "One reference",
			source:    []string{"test1", "test2", "test3"},
			reference: []string{"test3"},
			expected:  []string{"test3", "test1", "test2"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := OrderStringsSlices(tc.source, tc.reference)
			isOk := true
			for i := range result {
				if result[i] != tc.expected[i] {
					isOk = false
					break
				}
			}
			if !isOk {
				t.Errorf("OrderStringsSlices(%v,%v) = %s; want %s", tc.source, tc.reference, result, tc.expected)
			}

		})
	}
}
