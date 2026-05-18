package workflowruntime

import (
	"time"

	workflowruntimev1 "github.com/byte-v-forge/contracts-go/byte/v/forge/contracts/workflowruntime/v1"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"google.golang.org/protobuf/types/known/durationpb"
)

func ToProtoTaskQueueRef(config Config, owningService string) *workflowruntimev1.TaskQueueRef {
	return &workflowruntimev1.TaskQueueRef{
		Namespace:     config.Namespace,
		Name:          config.TaskQueue,
		OwningService: owningService,
	}
}

func ToProtoRetryPolicy(policy *temporal.RetryPolicy) *workflowruntimev1.RetryPolicy {
	if policy == nil {
		return nil
	}
	return &workflowruntimev1.RetryPolicy{
		InitialInterval:        duration(policy.InitialInterval),
		BackoffCoefficient:     policy.BackoffCoefficient,
		MaximumInterval:        duration(policy.MaximumInterval),
		MaximumAttempts:        policy.MaximumAttempts,
		NonRetryableErrorTypes: append([]string(nil), policy.NonRetryableErrorTypes...),
	}
}

func FromProtoRetryPolicy(policy *workflowruntimev1.RetryPolicy) *temporal.RetryPolicy {
	if policy == nil {
		return nil
	}
	return &temporal.RetryPolicy{
		InitialInterval:        protoDuration(policy.GetInitialInterval()),
		BackoffCoefficient:     policy.GetBackoffCoefficient(),
		MaximumInterval:        protoDuration(policy.GetMaximumInterval()),
		MaximumAttempts:        policy.GetMaximumAttempts(),
		NonRetryableErrorTypes: append([]string(nil), policy.GetNonRetryableErrorTypes()...),
	}
}

func ToProtoActivityPolicy(options workflow.ActivityOptions) *workflowruntimev1.ActivityPolicy {
	return &workflowruntimev1.ActivityPolicy{
		ScheduleToCloseTimeout: duration(options.ScheduleToCloseTimeout),
		StartToCloseTimeout:    duration(options.StartToCloseTimeout),
		ScheduleToStartTimeout: duration(options.ScheduleToStartTimeout),
		HeartbeatTimeout:       duration(options.HeartbeatTimeout),
		RetryPolicy:            ToProtoRetryPolicy(options.RetryPolicy),
	}
}

func FromProtoActivityPolicy(policy *workflowruntimev1.ActivityPolicy) workflow.ActivityOptions {
	if policy == nil {
		return workflow.ActivityOptions{}
	}
	return workflow.ActivityOptions{
		ScheduleToCloseTimeout: protoDuration(policy.GetScheduleToCloseTimeout()),
		StartToCloseTimeout:    protoDuration(policy.GetStartToCloseTimeout()),
		ScheduleToStartTimeout: protoDuration(policy.GetScheduleToStartTimeout()),
		HeartbeatTimeout:       protoDuration(policy.GetHeartbeatTimeout()),
		RetryPolicy:            FromProtoRetryPolicy(policy.GetRetryPolicy()),
	}
}

func duration(value time.Duration) *durationpb.Duration {
	if value == 0 {
		return nil
	}
	return durationpb.New(value)
}

func protoDuration(value *durationpb.Duration) time.Duration {
	if value == nil {
		return 0
	}
	return value.AsDuration()
}
