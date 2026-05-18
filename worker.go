package workflowruntime

import (
	"context"
	"errors"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	temporalworkflow "go.temporal.io/sdk/workflow"
)

type WorkflowDefinition struct {
	Name       string
	Definition any
}

type ActivityDefinition struct {
	Name       string
	Definition any
}

type WorkerSpec struct {
	TaskQueue  string
	Workflows  []WorkflowDefinition
	Activities []ActivityDefinition
	Options    worker.Options
}

func (s WorkerSpec) Validate() error {
	if err := ValidateTaskQueueName(s.TaskQueue); err != nil {
		return err
	}
	if len(s.Workflows) == 0 && len(s.Activities) == 0 {
		return validationError("worker", "must register at least one workflow or activity")
	}
	for _, definition := range s.Workflows {
		if definition.Name == "" {
			return validationError("workflow.name", "is required")
		}
		if definition.Definition == nil {
			return validationError("workflow.definition", "is required")
		}
	}
	for _, definition := range s.Activities {
		if definition.Name == "" {
			return validationError("activity.name", "is required")
		}
		if definition.Definition == nil {
			return validationError("activity.definition", "is required")
		}
	}
	return nil
}

func NewWorker(c client.Client, spec WorkerSpec) (worker.Worker, error) {
	if c == nil {
		return nil, validationError("client", "is required")
	}
	if err := spec.Validate(); err != nil {
		return nil, err
	}
	w := worker.New(c, spec.TaskQueue, spec.Options)
	for _, definition := range spec.Workflows {
		w.RegisterWorkflowWithOptions(definition.Definition, temporalworkflow.RegisterOptions{Name: definition.Name})
	}
	for _, definition := range spec.Activities {
		w.RegisterActivityWithOptions(definition.Definition, activity.RegisterOptions{Name: definition.Name})
	}
	return w, nil
}

func RunWorker(ctx context.Context, c client.Client, spec WorkerSpec) error {
	if ctx == nil {
		return validationError("context", "is required")
	}
	w, err := NewWorker(c, spec)
	if err != nil {
		return err
	}
	if err := w.Start(); err != nil {
		return err
	}
	<-ctx.Done()
	w.Stop()
	if errors.Is(ctx.Err(), context.Canceled) {
		return nil
	}
	return ctx.Err()
}
