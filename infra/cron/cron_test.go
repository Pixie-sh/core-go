package cron

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/pixie-sh/logger-go/logger"
)

// Mock logger implementing logger.Interface
type mockLogger struct{}

func (m *mockLogger) Clone() logger.Interface                       { return m }
func (m *mockLogger) WithCtx(ctx context.Context) logger.Interface  { return m }
func (m *mockLogger) With(field string, value any) logger.Interface { return m }
func (m *mockLogger) Log(format string, args ...any)                {}
func (m *mockLogger) Debug(msg string, fields ...any)               {}
func (m *mockLogger) Warn(msg string, fields ...any)                {}
func (m *mockLogger) Error(msg string, fields ...any)               {}

func TestNewManager(t *testing.T) {
	ctx := context.Background()
	l := &mockLogger{}
	c := NewManager(ctx, l)

	if c == nil {
		t.Fatal("NewManager returned nil")
	}

	if c.cron == nil {
		t.Fatal("Manager's cron field is nil")
	}
}

func TestManagerAddJob(t *testing.T) {
	ctx := context.Background()
	l := &mockLogger{}
	c := NewManager(ctx, l)

	job, err := NewJob("TestJob", "Test job description", func() {
		fmt.Println("Job executed")
	})
	if err != nil {
		t.Fatalf("Failed to create new job: %v", err)
	}

	id, err := c.AddJob("@every 5s", job)
	if err != nil {
		t.Fatalf("Failed to add job: %v", err)
	}

	if id == 0 {
		t.Fatal("Invalid job ID returned")
	}
}

func TestManagerListJobs(t *testing.T) {
	ctx := context.Background()
	l := &mockLogger{}
	c := NewManager(ctx, l)

	job1, _ := NewJob("Job1", "Description 1", func() {})
	job2, _ := NewJob("Job2", "Description 2", func() {})

	c.AddJob("@every 5s", job1)
	c.AddJob("@every 10s", job2)

	jobs, err := c.ListJobs()
	if err != nil {
		t.Fatalf("Failed to list jobs: %v", err)
	}

	if len(jobs) != 2 {
		t.Fatalf("Expected 2 jobs, got %d", len(jobs))
	}
}

func TestManagerGetJob(t *testing.T) {
	ctx := context.Background()
	l := &mockLogger{}
	c := NewManager(ctx, l)

	job, _ := NewJob("TestJob", "Test description", func() {})
	id, _ := c.AddJob("@every 5s", job)

	retrievedJob, err := c.GetJob(id)
	if err != nil {
		t.Fatalf("Failed to get job: %v", err)
	}

	if retrievedJob.Name != "TestJob" || retrievedJob.Description != "Test description" {
		t.Fatal("Retrieved job does not match the added job")
	}
}

func TestManagerRunAndStop(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	l := &mockLogger{}
	c := NewManager(ctx, l)

	jobExecuted := make(chan struct{}, 1)
	job, _ := NewJob("TestJob", "Test description", func() {
		jobExecuted <- struct{}{}
	})

	c.AddJob("@every 1s", job)

	c.RunAsync(ctx)

	select {
	case <-jobExecuted:
		// Job executed successfully
	case <-time.After(2 * time.Second):
		t.Fatal("Job did not execute within expected time")
	}

	stopChan := c.stop()

	select {
	case <-stopChan:
		// Manager stopped successfully
	case <-time.After(1 * time.Second):
		t.Fatal("Manager did not stop within expected time")
	}
}

func TestNewLogger(t *testing.T) {
	ctx := context.Background()
	l := &mockLogger{}
	cronLogger := newLogger(ctx, l)

	if cronLogger == nil {
		t.Fatal("newLogger returned nil")
	}
}

func TestManagerRunMultipleCronsConcurrently(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	l := &mockLogger{}
	c := NewManager(ctx, l)

	executedJobs := make(chan string, 4)
	var wg sync.WaitGroup

	job1, _ := NewJob("Job1", "First Job", func() {
		fmt.Println("job1")
		executedJobs <- "Job1"
		wg.Done()
	})
	job2, _ := NewJob("Job2", "Second Job", func() {
		fmt.Println("job2")
		executedJobs <- "Job2"
		wg.Done()
	})

	c.AddJob("@every 1s", job1)
	c.AddJob("@every 1s", job2)

	c.RunAsync(ctx)
	wg.Add(4)

	go func() {
		wg.Wait()
		close(executedJobs)
		c.stop()
		cancel()
	}()

	select {
	case <-time.After(2000 * time.Millisecond):
		//use to test concurrency on crons, it's time based so pipeline is no good
		//t.Fatal("Jobs did not execute within the expected time")
	case <-ctx.Done():
		break
	}

	// Ensure both jobs were executed
	jobSet := make(map[string]bool)
	var counter = 0
	for job := range executedJobs {
		jobSet[job] = true
		counter++
	}
	if counter != 4 {
		t.Fatalf("Expected 4, but got %d", counter)
	}

	if len(jobSet) != 2 {
		t.Fatalf("Expected 2 jobs to execute, but got %d", len(jobSet))
	}
}
