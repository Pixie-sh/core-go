package dates

import (
	"fmt"
	"sort"
	"time"
)

func ValidateWeekdays(weekdays []int32) bool {
	seen := make(map[int32]bool)
	for _, day := range weekdays {
		if day < 0 || day > 6 {
			return false
		}

		if seen[day] {
			return false
		}
		seen[day] = true
	}
	return true
}

// ConvertWeekdaysToUTC transforms an array of weekdays from a specified timezone to UTC.
// Weekdays are represented as integers: 0 (Sunday) through 6 (Saturday).
// We assume each weekday refers to midnight (00:00) in the specified timezone.
//
// Parameters:
//   - weekdays: Array of weekdays in the local timezone (0-6)
//   - timezone: IANA timezone string (e.g., "Europe/Lisbon")
//
// Returns:
//   - Array of weekdays in UTC
//   - Error if any occurs
func ConvertWeekdaysToUTC(weekdays []int32, timezone string) ([]int32, error) {
	if len(weekdays) == 0 {
		return nil, fmt.Errorf("weekdays array cannot be empty")
	}

	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, fmt.Errorf("invalid timezone: %v", err)
	}

	// Use a fixed reference week in July 2023 (during DST for most timezones)
	// July 2, 2023 was a Sunday, so we can easily calculate each weekday
	utcWeekdaysMap := make(map[int32]bool)

	for _, weekday := range weekdays {
		// Validate weekday
		if weekday < 0 || weekday > 6 {
			return nil, fmt.Errorf("weekday must be between 0 (Sunday) and 6 (Saturday)")
		}

		// Create midnight on the given weekday in the reference week (July 2-8, 2023)
		// in the specified timezone
		day := 2 + int(weekday) // July 2 is Sunday (0), July 3 is Monday (1), etc.
		midnight := time.Date(2023, 7, day, 0, 0, 0, 0, loc)

		// Convert to UTC and get the weekday
		utcTime := midnight.UTC()
		utcWeekday := int32(utcTime.Weekday())
		utcWeekdaysMap[utcWeekday] = true
	}

	utcWeekdays := make([]int32, 0, len(utcWeekdaysMap))
	for weekday := range utcWeekdaysMap {
		utcWeekdays = append(utcWeekdays, weekday)
	}

	sort.Slice(utcWeekdays, func(i, j int) bool {
		return utcWeekdays[i] < utcWeekdays[j]
	})

	return utcWeekdays, nil
}
