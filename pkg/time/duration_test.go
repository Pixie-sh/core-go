package time

import (
	"testing"
	gotime "time"
)

func TestDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration Duration
		hours    float64
		minutes  float64
		seconds  float64
	}{
		{"zero", Duration(0), 0, 0, 0},
		{"one hour", Duration(gotime.Hour), 1, 60, 3600},
		{"one minute", Duration(gotime.Minute), 1.0 / 60, 1, 60},
		{"one second", Duration(gotime.Second), 1.0 / 3600, 1.0 / 60, 1},
		{"complex", Duration(gotime.Hour) + 30*Duration(gotime.Minute) + 45*Duration(gotime.Second), 1.5125, 90.75, 5445},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.duration.Hours(); got != tt.hours {
				t.Errorf("Duration.Hours() = %v, want %v", got, tt.hours)
			}
			if got := tt.duration.Minutes(); got != tt.minutes {
				t.Errorf("Duration.Minutes() = %v, want %v", got, tt.minutes)
			}
			if got := tt.duration.Seconds(); got != tt.seconds {
				t.Errorf("Duration.Seconds() = %v, want %v", got, tt.seconds)
			}
		})
	}
}

func TestDurationString(t *testing.T) {
	tests := []struct {
		duration Duration
		want     string
	}{
		{Duration(0), "0s"},
		{Duration(gotime.Hour), "1h0m0s"},
		{Duration(gotime.Minute), "1m0s"},
		{Duration(gotime.Second), "1s"},
		{Duration(gotime.Hour) + 30*Duration(gotime.Minute) + 45*Duration(gotime.Second), "1h30m45s"},
	}

	for _, tt := range tests {
		if got := tt.duration.String(); got != tt.want {
			t.Errorf("Duration.String() = %v, want %v", got, tt.want)
		}
	}
}

func TestDurationParseValid(t *testing.T) {
	tests := []struct {
		str  string
		want Duration
	}{
		{"0s", Duration(0)},
		{"1h", Duration(gotime.Hour)},
		{"1m", Duration(gotime.Minute)},
		{"1s", Duration(gotime.Second)},
		{"1h30m45s", Duration(gotime.Hour) + 30*Duration(gotime.Minute) + 45*Duration(gotime.Second)},
	}

	for _, tt := range tests {
		got, err := ParseDuration(tt.str)
		if err != nil {
			t.Errorf("ParseDuration(%q) unexpected error: %v", tt.str, err)
		}
		if got != tt.want {
			t.Errorf("ParseDuration(%q) = %v, want %v", tt.str, got, tt.want)
		}
	}
}

func TestDurationParseInvalid(t *testing.T) {
	tests := []string{
		"",
		"3",
		"-",
		"1h30m45",
		"1h30m45x",
		"1.5x",
		"1hh",
	}

	for _, tt := range tests {
		_, err := ParseDuration(tt)
		if err == nil {
			t.Errorf("ParseDuration(%q) expected error, got nil", tt)
		}
	}
}
