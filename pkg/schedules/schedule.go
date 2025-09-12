package schedules

import (
	"time"

	"github.com/pixie-sh/core-go/pkg/types"
)

// ScheduleData represents the schedule information to be stored
type ScheduleData struct {
	Year     int    `json:"year"`
	Month    int    `json:"month"`
	Day      int    `json:"day"`
	Hours    int    `json:"hours"`
	Minutes  int    `json:"minutes"`
	Timezone string `json:"timezone"` // e.g., "America/New_York"
	IsDST    bool   `json:"isDST"`    // true if daylight saving is active for this date/time in the given timezone
}

// Schedule stores and manipulates timezone-aware schedule data
type Schedule struct {
	schedule ScheduleData
}

func NewScheduleFromTime(t time.Time, targetTimezone string) (Schedule, error) {
	// Load the target location
	targetLoc, err := time.LoadLocation(targetTimezone)
	if err != nil {
		return Schedule{}, err
	}

	// Convert time to the target timezone
	localTime := t.In(targetLoc)

	// Extract components in the target timezone
	year := localTime.Year()
	month := int(localTime.Month())
	day := localTime.Day()
	hours := localTime.Hour()
	minutes := localTime.Minute()

	// Check if DST is in effect
	_, isDST := localTime.Zone()

	schedule := ScheduleData{
		Year:     year,
		Month:    month,
		Day:      day,
		Hours:    hours,
		Minutes:  minutes,
		Timezone: targetTimezone,
		IsDST:    isDST != 0, // In Go, isDST is seconds offset from standard time
	}

	return Schedule{schedule: schedule}, nil
}

// NewSchedule creates a new schedule from local date/time and a timezone.
// The function determines if DST is active for the given date/time and timezone.
func NewSchedule(year, month, day, hours, minutes int, timezone string) (Schedule, error) {
	// Create a time.Time from the provided components
	// Go month is 1-12, so we don't need to adjust it
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return Schedule{}, err
	}

	// Create the time in the specified timezone
	dt := time.Date(year, time.Month(month), day, hours, minutes, 0, 0, loc)

	// Check if DST is in effect
	_, isDST := dt.Zone()

	schedule := ScheduleData{
		Year:     year,
		Month:    month,
		Day:      day,
		Hours:    hours,
		Minutes:  minutes,
		Timezone: timezone,
		IsDST:    isDST != 0, // In Go, isDST is seconds offset from standard time
	}

	return Schedule{schedule: schedule}, nil
}

func NewScheduleDataNowUTC() ScheduleData {
	now := time.Now().UTC()
	return ScheduleData{
		Year:     now.Year(),
		Month:    int(now.Month()),
		Day:      now.Day(),
		Hours:    now.Hour(),
		Minutes:  now.Minute(),
		Timezone: "UTC",
		IsDST:    false,
	}
}

func (s *ScheduleData) SetToStartOfDay() {
	s.Hours = 0
	s.Minutes = 0
}

func (s *ScheduleData) SetToEndOfDay() {
	s.Hours = 23
	s.Minutes = 59
}

// ToUTC converts the stored schedule (local to the user's timezone) into UTC time.
// This is useful for saving a canonical time to the server.
func (s Schedule) ToUTC() time.Time {
	loc, err := time.LoadLocation(s.schedule.Timezone)
	if err != nil {
		// Handle error, falling back to UTC
		loc = time.UTC
	}
	// Create the time in the original timezone
	dt := time.Date(
		s.schedule.Year,
		time.Month(s.schedule.Month),
		s.schedule.Day,
		s.schedule.Hours,
		s.schedule.Minutes,
		0, 0, loc)

	// Return it converted to UTC
	return dt.UTC()
}

// ToOriginal returns the time in the original schedule's timezone
func (s Schedule) ToOriginal() time.Time {
	loc, err := time.LoadLocation(s.schedule.Timezone)
	if err != nil {
		// Handle error, falling back to UTC
		loc = time.UTC
	}

	// Create and return the time in the original timezone
	return time.Date(
		s.schedule.Year,
		time.Month(s.schedule.Month),
		s.schedule.Day,
		s.schedule.Hours,
		s.schedule.Minutes,
		0, 0, loc)
}

// ToLocal converts the stored schedule to a target timezone.
// This is useful when displaying the schedule to another user who is in a different timezone.
func (s Schedule) ToLocal(targetZone string) (time.Time, error) {
	// First, get the time in its original timezone
	originalTime := s.ToOriginal()

	// Then convert to the target timezone
	targetLoc, err := time.LoadLocation(targetZone)
	if err != nil {
		return time.Time{}, err
	}

	return originalTime.In(targetLoc), nil
}

// GetScheduleData returns the schedule data.
// This data is what you would send to your server via REST.
func (s Schedule) GetScheduleData() ScheduleData {
	return s.schedule
}

// IsEmpty check if internal data has zero values
func (s Schedule) IsEmpty() bool {
	return types.IsEmpty(s.schedule)
}

// FromScheduleData is a factory function to re-create a Schedule instance from stored schedule data.
func FromScheduleData(data ScheduleData) (Schedule, error) {
	return NewSchedule(
		data.Year,
		data.Month,
		data.Day,
		data.Hours,
		data.Minutes,
		data.Timezone,
	)
}
