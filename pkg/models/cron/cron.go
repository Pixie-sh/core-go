package cron

import (
	"time"
)

// jobInstance avoid circular imports
type jobInstance interface {
	Name() string
	Description() string
	Run()
}

type CronJob struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	NextRun     time.Time `json:"next_run"`
	PreviousRun time.Time `json:"previous_run"`

	JobInstance jobInstance `json:"-"`
}
