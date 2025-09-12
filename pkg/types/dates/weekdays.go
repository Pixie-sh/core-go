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

	now := time.Now()

	utcWeekdaysMap := make(map[int32]bool)

	for _, weekday := range weekdays {
		// Validate weekday
		if weekday < 0 || weekday > 6 {
			return nil, fmt.Errorf("weekday must be between 0 (Sunday) and 6 (Saturday)")
		}

		// Calculate the date of the next occurrence of this weekday
		daysToAdd := (int(weekday) - int(now.Weekday()) + 7) % 7
		targetDate := now.AddDate(0, 0, daysToAdd)

		// Create a time at midnight in the specified timezone
		midnight := time.Date(
			targetDate.Year(),
			targetDate.Month(),
			targetDate.Day(),
			0, 0, 0, 0,
			loc,
		)

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
