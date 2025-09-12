package dates

import (
	"testing"
	"time"

	"github.com/pixie-sh/core-go/pkg/schedules"
)

func TestIsBetweenIgnoreTimeScheduleData(t *testing.T) {
	type testCase struct {
		name     string
		date     schedules.ScheduleData
		start    schedules.ScheduleData
		end      schedules.ScheduleData
		expected bool
	}

	testCases := []testCase{
		{
			name: "date is between start and end",
			date: schedules.ScheduleData{
				Year:     2025,
				Month:    4,
				Day:      20,
				Hours:    12,
				Minutes:  0,
				Timezone: "Europe/Paris",
				IsDST:    true,
			},
			start: schedules.ScheduleData{
				Year:     2025,
				Month:    4,
				Day:      18,
				Hours:    0,
				Minutes:  0,
				Timezone: "Europe/Paris",
				IsDST:    true,
			},
			end: schedules.ScheduleData{
				Year:     2025,
				Month:    4,
				Day:      20,
				Hours:    23,
				Minutes:  59,
				Timezone: "Europe/Paris",
				IsDST:    true,
			},
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			askSchedule, err := schedules.FromScheduleData(tc.date)
			if err != nil {
				panic(err)
			}
			utcAskDate := askSchedule.ToUTC()

			dealScheduleStartAt, err := schedules.FromScheduleData(tc.start)
			if err != nil {
				panic(err)
			}
			utcDealDateStartAt := dealScheduleStartAt.ToUTC()

			dealScheduleEndAt, err := schedules.FromScheduleData(tc.end)
			if err != nil {
				panic(err)
			}
			utcDealDateEndAt := dealScheduleEndAt.ToUTC()

			result := IsBetweenIgnoreTime(utcAskDate, utcDealDateStartAt, utcDealDateEndAt)
			if result != tc.expected {
				t.Errorf("IsBetweenIgnoreTime(%v, %v, %v) = %v; want %v",
					utcAskDate, utcDealDateStartAt, utcDealDateEndAt, result, tc.expected)
			}
		})
	}
}

func TestIsBetweenIgnoreTime(t *testing.T) {
	type testCase struct {
		name     string
		date     time.Time
		start    time.Time
		end      time.Time
		expected bool
	}

	testCases := []testCase{
		{
			name:     "date is between start and end",
			date:     time.Date(2023, 5, 15, 14, 30, 0, 0, time.UTC),
			start:    time.Date(2023, 5, 10, 8, 0, 0, 0, time.UTC),
			end:      time.Date(2023, 5, 20, 18, 0, 0, 0, time.UTC),
			expected: true,
		},
		{
			name:     "date is equal to start",
			date:     time.Date(2023, 5, 10, 14, 30, 0, 0, time.UTC),
			start:    time.Date(2023, 5, 10, 8, 0, 0, 0, time.UTC),
			end:      time.Date(2023, 5, 20, 18, 0, 0, 0, time.UTC),
			expected: true,
		},
		{
			name:     "date is equal to end",
			date:     time.Date(2023, 5, 20, 14, 30, 0, 0, time.UTC),
			start:    time.Date(2023, 5, 10, 8, 0, 0, 0, time.UTC),
			end:      time.Date(2023, 5, 20, 18, 0, 0, 0, time.UTC),
			expected: true,
		},
		{
			name:     "date is before start",
			date:     time.Date(2023, 5, 5, 14, 30, 0, 0, time.UTC),
			start:    time.Date(2023, 5, 10, 8, 0, 0, 0, time.UTC),
			end:      time.Date(2023, 5, 20, 18, 0, 0, 0, time.UTC),
			expected: false,
		},
		{
			name:     "date is after end",
			date:     time.Date(2023, 5, 25, 14, 30, 0, 0, time.UTC),
			start:    time.Date(2023, 5, 10, 8, 0, 0, 0, time.UTC),
			end:      time.Date(2023, 5, 20, 18, 0, 0, 0, time.UTC),
			expected: false,
		},
		{
			name:     "different time parts should be ignored",
			date:     time.Date(2023, 5, 15, 23, 59, 59, 999, time.UTC),
			start:    time.Date(2023, 5, 10, 0, 0, 0, 0, time.UTC),
			end:      time.Date(2023, 5, 20, 0, 0, 0, 0, time.UTC),
			expected: true,
		},
		{
			name:     "different locations should not affect result",
			date:     time.Date(2023, 5, 15, 14, 30, 0, 0, time.FixedZone("UTC+2", 2*60*60)),
			start:    time.Date(2023, 5, 10, 8, 0, 0, 0, time.UTC),
			end:      time.Date(2023, 5, 20, 18, 0, 0, 0, time.FixedZone("UTC-5", -5*60*60)),
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsBetweenIgnoreTime(tc.date, tc.start, tc.end)
			if result != tc.expected {
				t.Errorf("IsBetweenIgnoreTime(%v, %v, %v) = %v; want %v",
					tc.date, tc.start, tc.end, result, tc.expected)
			}
		})
	}
}

func TestIsBeforeIgnoreTime(t *testing.T) {
	type testCase struct {
		name     string
		date     time.Time
		start    time.Time
		expected bool
	}

	testCases := []testCase{
		{
			name:     "date is after start",
			date:     time.Date(2023, 5, 15, 14, 30, 0, 0, time.UTC),
			start:    time.Date(2023, 5, 10, 8, 0, 0, 0, time.UTC),
			expected: false,
		},
		{
			name:     "date is before start",
			date:     time.Date(2023, 5, 8, 14, 30, 0, 0, time.UTC),
			start:    time.Date(2023, 5, 10, 8, 0, 0, 0, time.UTC),
			expected: true,
		},
		{
			name:     "date is equal to start",
			date:     time.Date(2023, 5, 10, 14, 30, 0, 0, time.UTC),
			start:    time.Date(2023, 5, 10, 8, 0, 0, 0, time.UTC),
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsBeforeIgnoreTime(tc.date, tc.start)
			if result != tc.expected {
				t.Errorf("IsBeforeIgnoreTime(%v, %v) = %v; want %v",
					tc.date, tc.start, result, tc.expected)
			}
		})
	}
}

func TestIsAfterIgnoreTime(t *testing.T) {
	type testCase struct {
		name     string
		date     time.Time
		start    time.Time
		expected bool
	}

	testCases := []testCase{
		{
			name:     "date is after start",
			date:     time.Date(2023, 5, 15, 14, 30, 0, 0, time.UTC),
			start:    time.Date(2023, 5, 10, 8, 0, 0, 0, time.UTC),
			expected: true,
		},
		{
			name:     "date is before start",
			date:     time.Date(2023, 5, 8, 14, 30, 0, 0, time.UTC),
			start:    time.Date(2023, 5, 10, 8, 0, 0, 0, time.UTC),
			expected: false,
		},
		{
			name:     "date is equal to start",
			date:     time.Date(2023, 5, 10, 14, 30, 0, 0, time.UTC),
			start:    time.Date(2023, 5, 10, 8, 0, 0, 0, time.UTC),
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsAfterIgnoreTime(tc.date, tc.start)
			if result != tc.expected {
				t.Errorf("IsAfterIgnoreTime(%v, %v) = %v; want %v",
					tc.date, tc.start, result, tc.expected)
			}
		})
	}
}

func TestIsBetweenIgnoreDate(t *testing.T) {
	type testCase struct {
		name     string
		time     time.Time
		start    time.Time
		end      time.Time
		expected bool
	}

	testCases := []testCase{
		{
			name:     "time is between start and end",
			time:     time.Date(2023, 5, 15, 14, 30, 0, 0, time.UTC),
			start:    time.Date(2023, 5, 20, 10, 0, 0, 0, time.UTC),
			end:      time.Date(2023, 5, 20, 18, 0, 0, 0, time.UTC),
			expected: true,
		},
		{
			name:     "time is equal to start",
			time:     time.Date(2023, 5, 15, 10, 0, 0, 0, time.UTC),
			start:    time.Date(2023, 5, 20, 10, 0, 0, 0, time.UTC),
			end:      time.Date(2023, 5, 20, 18, 0, 0, 0, time.UTC),
			expected: true,
		},
		{
			name:     "time is equal to end",
			time:     time.Date(2023, 5, 15, 18, 0, 0, 0, time.UTC),
			start:    time.Date(2023, 5, 20, 10, 0, 0, 0, time.UTC),
			end:      time.Date(2023, 5, 20, 18, 0, 0, 0, time.UTC),
			expected: true,
		},
		{
			name:     "time is before start",
			time:     time.Date(2023, 5, 15, 9, 30, 0, 0, time.UTC),
			start:    time.Date(2023, 5, 20, 10, 0, 0, 0, time.UTC),
			end:      time.Date(2023, 5, 20, 18, 0, 0, 0, time.UTC),
			expected: false,
		},
		{
			name:     "time is after end",
			time:     time.Date(2023, 5, 15, 19, 30, 0, 0, time.UTC),
			start:    time.Date(2023, 5, 20, 10, 0, 0, 0, time.UTC),
			end:      time.Date(2023, 5, 20, 18, 0, 0, 0, time.UTC),
			expected: false,
		},
		{
			name:     "different date parts should be ignored",
			time:     time.Date(2022, 1, 1, 14, 30, 0, 0, time.UTC),
			start:    time.Date(2023, 5, 20, 10, 0, 0, 0, time.UTC),
			end:      time.Date(2023, 5, 20, 18, 0, 0, 0, time.UTC),
			expected: true,
		},
		{
			name:     "nanoseconds should be considered",
			time:     time.Date(2023, 5, 15, 14, 30, 0, 1, time.UTC),
			start:    time.Date(2023, 5, 20, 14, 30, 0, 0, time.UTC),
			end:      time.Date(2023, 5, 20, 18, 0, 0, 0, time.UTC),
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsBetweenIgnoreDate(tc.time, tc.start, tc.end)
			if result != tc.expected {
				t.Errorf("IsBetweenIgnoreDate(%v, %v, %v) = %v; want %v",
					tc.time, tc.start, tc.end, result, tc.expected)
			}
		})
	}
}

func TestIsBeforeIgnoreDate(t *testing.T) {
	type testCase struct {
		name     string
		time     time.Time
		compare  time.Time
		expected bool
	}

	testCases := []testCase{
		{
			name:     "time is before compare",
			time:     time.Date(2023, 5, 15, 10, 0, 0, 0, time.UTC),
			compare:  time.Date(2023, 5, 20, 14, 30, 0, 0, time.UTC),
			expected: true,
		},
		{
			name:     "time is equal to compare",
			time:     time.Date(2023, 5, 15, 14, 30, 0, 0, time.UTC),
			compare:  time.Date(2023, 5, 20, 14, 30, 0, 0, time.UTC),
			expected: false,
		},
		{
			name:     "time is after compare",
			time:     time.Date(2023, 5, 15, 18, 0, 0, 0, time.UTC),
			compare:  time.Date(2023, 5, 20, 14, 30, 0, 0, time.UTC),
			expected: false,
		},
		{
			name:     "different date parts should be ignored",
			time:     time.Date(2022, 1, 1, 10, 0, 0, 0, time.UTC),
			compare:  time.Date(2023, 5, 20, 14, 30, 0, 0, time.UTC),
			expected: true,
		},
		{
			name:     "nanoseconds should be considered",
			time:     time.Date(2023, 5, 15, 14, 30, 0, 0, time.UTC),
			compare:  time.Date(2023, 5, 20, 14, 30, 0, 1, time.UTC),
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsBeforeIgnoreDate(tc.time, tc.compare)
			if result != tc.expected {
				t.Errorf("IsBeforeIgnoreDate(%v, %v) = %v; want %v",
					tc.time, tc.compare, result, tc.expected)
			}
		})
	}
}

func TestIsAfterIgnoreDate(t *testing.T) {
	type testCase struct {
		name     string
		time     time.Time
		compare  time.Time
		expected bool
	}

	testCases := []testCase{
		{
			name:     "time is after compare",
			time:     time.Date(2023, 5, 15, 18, 0, 0, 0, time.UTC),
			compare:  time.Date(2023, 5, 20, 14, 30, 0, 0, time.UTC),
			expected: true,
		},
		{
			name:     "time is equal to compare",
			time:     time.Date(2023, 5, 15, 14, 30, 0, 0, time.UTC),
			compare:  time.Date(2023, 5, 20, 14, 30, 0, 0, time.UTC),
			expected: false,
		},
		{
			name:     "time is before compare",
			time:     time.Date(2023, 5, 15, 10, 0, 0, 0, time.UTC),
			compare:  time.Date(2023, 5, 20, 14, 30, 0, 0, time.UTC),
			expected: false,
		},
		{
			name:     "different date parts should be ignored",
			time:     time.Date(2024, 12, 31, 18, 0, 0, 0, time.UTC),
			compare:  time.Date(2023, 5, 20, 14, 30, 0, 0, time.UTC),
			expected: true,
		},
		{
			name:     "nanoseconds should be considered",
			time:     time.Date(2023, 5, 15, 14, 30, 0, 1, time.UTC),
			compare:  time.Date(2023, 5, 20, 14, 30, 0, 0, time.UTC),
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsAfterIgnoreDate(tc.time, tc.compare)
			if result != tc.expected {
				t.Errorf("IsAfterIgnoreDate(%v, %v) = %v; want %v",
					tc.time, tc.compare, result, tc.expected)
			}
		})
	}
}
