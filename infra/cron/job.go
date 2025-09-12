package cron

import (
	"github.com/pixie-sh/errors-go"

	"github.com/pixie-sh/core-go/pkg/types"
)

type Job interface {
	Name() string
	Description() string
	Run()
}

type job struct {
	name        string
	description string
	function    func()
}

func (j *job) Name() string {
	return j.name
}

func (j *job) Description() string {
	return j.description
}

func (j *job) Run() {
	j.function()
}

func NewJob(name, description string, f func()) (Job, error) {
	if types.Nil(f) {
		return nil, errors.New("function is nil")
	}

	if types.IsEmpty(name) {
		return nil, errors.New("name is empty")
	}

	if types.IsEmpty(description) {
		return nil, errors.New("description is empty")
	}

	return &job{
		name:        name,
		description: description,
		function:    f,
	}, nil
}
