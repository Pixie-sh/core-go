package dates

import "time"

// Compare only the date part
func IsBetweenIgnoreTime(date, start, end time.Time) bool {
	dateOnly := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	startOnly := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
	endOnly := time.Date(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0, end.Location())

	return !dateOnly.Before(startOnly) && !dateOnly.After(endOnly)
}

// Compare only the time part
func IsBetweenIgnoreDate(t, start, end time.Time) bool {
	refYear, refMonth, refDay := 1, time.January, 1

	tOnly := time.Date(refYear, refMonth, refDay, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), time.UTC)
	startOnly := time.Date(refYear, refMonth, refDay, start.Hour(), start.Minute(), start.Second(), start.Nanosecond(), time.UTC)
	endOnly := time.Date(refYear, refMonth, refDay, end.Hour(), end.Minute(), end.Second(), end.Nanosecond(), time.UTC)

	return !tOnly.Before(startOnly) && !tOnly.After(endOnly)
}

// Compare only the date part
func IsAfterIgnoreTime(date, start time.Time) bool {
	dateOnly := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	startOnly := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())

	return dateOnly.After(startOnly)
}

// Compare only the date part
func IsBeforeIgnoreTime(date, start time.Time) bool {
	dateOnly := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	startOnly := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())

	return dateOnly.Before(startOnly)
}

// Compare only the time part
func IsBeforeIgnoreDate(t, compare time.Time) bool {
	refYear, refMonth, refDay := 1, time.January, 1

	tOnly := time.Date(refYear, refMonth, refDay, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), time.UTC)
	compareOnly := time.Date(refYear, refMonth, refDay, compare.Hour(), compare.Minute(), compare.Second(), compare.Nanosecond(), time.UTC)

	return tOnly.Before(compareOnly)
}

// Compare only the time part
func IsAfterIgnoreDate(t, compare time.Time) bool {
	refYear, refMonth, refDay := 1, time.January, 1

	tOnly := time.Date(refYear, refMonth, refDay, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), time.UTC)
	compareOnly := time.Date(refYear, refMonth, refDay, compare.Hour(), compare.Minute(), compare.Second(), compare.Nanosecond(), time.UTC)

	return tOnly.After(compareOnly)
}
