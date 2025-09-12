package schedules

import (
	"fmt"
	"testing"
	"time"
)

func TestSchedules(t *testing.T) {
	// 1. Creating a schedule using the local time for a user (e.g., America/New_York)
	userSchedule, err := NewSchedule(2025, 3, 28, 11, 19, "Europe/London")
	if err != nil {
		fmt.Println("Error creating schedule:", err)
		return
	}

	userSchedule2, _ := NewScheduleFromTime(userSchedule.ToUTC(), "Europe/London")

	fmt.Println("Sent on API:", userSchedule.GetScheduleData())
	fmt.Println("Sent on API:", userSchedule2.GetScheduleData())

	// Original time in the schedule's timezone
	fmt.Println("Received and Parsed on server - original:", userSchedule.ToOriginal().Format(time.RFC3339))
	fmt.Println("Received and Parsed on server - original:", userSchedule2.ToOriginal().Format(time.RFC3339))

	// UTC time (for storage or canonical representation)
	utc := userSchedule.ToUTC()
	fmt.Println("UTC time:", utc.Format(time.RFC3339))

	// Converting to another timezone
	berlinTime, err := userSchedule.ToLocal("Europe/Berlin")
	if err != nil {
		fmt.Println("Error converting to Berlin time:", err)
		return
	}
	fmt.Println("Time in Berlin:", berlinTime.Format(time.RFC3339))

	// Converting to another timezone
	lisbonTime, err := userSchedule.ToLocal("Europe/Lisbon")
	if err != nil {
		fmt.Println("Error converting to Lisbon time:", err)
		return
	}
	fmt.Println("Time in Lisbon:", lisbonTime.Format(time.RFC3339))
}
