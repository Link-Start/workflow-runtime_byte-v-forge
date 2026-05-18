package workflowruntime

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type ActivityOptionsMutator func(*workflow.ActivityOptions)

func DefaultRetryPolicy() *temporal.RetryPolicy {
	return &temporal.RetryPolicy{
		InitialInterval:    time.Second,
		BackoffCoefficient: 2,
		MaximumInterval:    30 * time.Second,
		MaximumAttempts:    5,
	}
}

func DefaultActivityOptions() workflow.ActivityOptions {
	return workflow.ActivityOptions{
		StartToCloseTimeout:    2 * time.Minute,
		ScheduleToCloseTimeout: 10 * time.Minute,
		HeartbeatTimeout:       30 * time.Second,
		RetryPolicy:            DefaultRetryPolicy(),
	}
}

func WithDefaultActivityOptions(ctx workflow.Context, mutators ...ActivityOptionsMutator) workflow.Context {
	options := DefaultActivityOptions()
	for _, mutator := range mutators {
		if mutator != nil {
			mutator(&options)
		}
	}
	return workflow.WithActivityOptions(ctx, options)
}
