package cron

import (
	"context"

	"github.com/pixie-sh/errors-go"
	"github.com/pixie-sh/logger-go/logger"
	"github.com/robfig/cron/v3"

	cronmodels "github.com/pixie-sh/core-go/pkg/models/cron"
)

type Manager struct {
	cron *cron.Cron
}

func NewManager(ctx context.Context, log logger.Interface) *Manager {
	return &Manager{
		cron.New(
			cron.WithLogger(newLogger(ctx, log)),
			func(c *cron.Cron) {
				c.Stop()
			}),
	}
}

// Run runs the scheduler; blocking call
func (c *Manager) Run(_ context.Context) error {
	c.cron.Run()
	return nil
}

// RunAsync runs the scheduler with cancel context; Run in a go routine
// Stop function must be called whenever the ctx is canceled
func (c *Manager) RunAsync(ctx context.Context) {
	go func() {
		_ = c.Run(ctx)
	loop:
		for {
			select {
			case <-ctx.Done():
				<-c.stop()
				break loop
			}
		}
	}()
}

// stop the internal cron jobs; cronCtx.Done() channel is returned
func (c *Manager) stop() <-chan struct{} {
	return c.cron.Stop().Done()
}

// ListJobs be aware this locks internal mutex
func (c *Manager) ListJobs() ([]cronmodels.CronJob, error) {
	var jobs []cronmodels.CronJob

	for _, entry := range c.cron.Entries() {
		if !entry.Valid() {
			continue
		}

		casted, ok := entry.WrappedJob.(Job)
		if !ok {
			return nil, errors.New("invalid job(%d) type", entry.ID).WithErrorCode(errors.ErrorPerformingRequestErrorCode)
		}

		jobs = append(jobs, c.jobModel(entry, casted))
	}

	return jobs, nil
}

func (c *Manager) GetJob(id int) (cronmodels.CronJob, error) {
	entry := c.cron.Entry(cron.EntryID(id))
	if !entry.Valid() {
		return cronmodels.CronJob{}, errors.New("job(%d) not found", id).WithErrorCode(errors.NotFoundErrorCode)
	}

	casted, ok := entry.WrappedJob.(Job)
	if !ok {
		return cronmodels.CronJob{}, errors.New("invalid job(%d) type", id).WithErrorCode(errors.ErrorPerformingRequestErrorCode)
	}

	return c.jobModel(entry, casted), nil
}

func (c *Manager) jobModel(entry cron.Entry, casted Job) cronmodels.CronJob {
	return cronmodels.CronJob{
		ID:          int(entry.ID),
		Name:        casted.Name(),
		Description: casted.Description(),
		NextRun:     entry.Next,
		PreviousRun: entry.Prev,
		JobInstance: casted,
	}
}

func (c *Manager) AddJob(schedule string, job Job) (int, error) {
	id, err := c.cron.AddJob(schedule, job)
	return int(id), err
}
