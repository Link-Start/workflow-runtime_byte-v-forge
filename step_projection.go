package workflowruntime

import (
	workflowruntimev1 "github.com/byte-v-forge/contracts-go/byte/v/forge/contracts/workflowruntime/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func NewRuntimeError(code, message string, retryable bool) *workflowruntimev1.WorkflowRuntimeError {
	return &workflowruntimev1.WorkflowRuntimeError{
		Code:      code,
		Message:   message,
		Retryable: retryable,
	}
}

func NewStepProjection(
	ref *workflowruntimev1.WorkflowStepRef,
	strategy workflowruntimev1.WorkflowStepFailureStrategy,
	policy *workflowruntimev1.ActivityPolicy,
) *workflowruntimev1.WorkflowStepProjection {
	return &workflowruntimev1.WorkflowStepProjection{
		Step:            cloneStepRef(ref),
		Status:          workflowruntimev1.WorkflowStepStatus_WORKFLOW_STEP_STATUS_PENDING,
		FailureStrategy: normalizeStepFailureStrategy(strategy),
		MaximumAttempts: maxAttempts(policy),
		ActivityPolicy:  cloneActivityPolicy(policy),
	}
}

func MarkStepRunning(
	step *workflowruntimev1.WorkflowStepProjection,
	attempt int32,
	now *timestamppb.Timestamp,
) *workflowruntimev1.WorkflowStepProjection {
	next := cloneStep(step)
	next.Status = workflowruntimev1.WorkflowStepStatus_WORKFLOW_STEP_STATUS_RUNNING
	next.CurrentAttempt = attempt
	next.LastError = nil
	if next.StartedAt == nil {
		next.StartedAt = now
	}
	next.UpdatedAt = now

	return next
}

func MarkStepSucceeded(
	step *workflowruntimev1.WorkflowStepProjection,
	now *timestamppb.Timestamp,
) *workflowruntimev1.WorkflowStepProjection {
	next := cloneStep(step)
	next.Status = workflowruntimev1.WorkflowStepStatus_WORKFLOW_STEP_STATUS_SUCCEEDED
	next.LastError = nil
	next.ClosedAt = now
	next.UpdatedAt = now
	next.Attempts = append(next.Attempts, &workflowruntimev1.WorkflowStepAttempt{
		Attempt:  next.GetCurrentAttempt(),
		Status:   workflowruntimev1.WorkflowStepStatus_WORKFLOW_STEP_STATUS_SUCCEEDED,
		ClosedAt: now,
	})

	return next
}

func MarkStepFailed(
	step *workflowruntimev1.WorkflowStepProjection,
	err *workflowruntimev1.WorkflowRuntimeError,
	now *timestamppb.Timestamp,
) *workflowruntimev1.WorkflowStepProjection {
	next := cloneStep(step)
	next.LastError = cloneRuntimeError(err)
	next.Status = failedStepStatus(next.GetFailureStrategy())
	next.ClosedAt = now
	next.UpdatedAt = now
	next.Attempts = append(next.Attempts, &workflowruntimev1.WorkflowStepAttempt{
		Attempt:  next.GetCurrentAttempt(),
		Status:   workflowruntimev1.WorkflowStepStatus_WORKFLOW_STEP_STATUS_FAILED,
		Error:    cloneRuntimeError(err),
		ClosedAt: now,
	})

	return next
}

func StepContinuesAfterFailure(step *workflowruntimev1.WorkflowStepProjection) bool {
	return step.GetFailureStrategy() == workflowruntimev1.WorkflowStepFailureStrategy_WORKFLOW_STEP_FAILURE_STRATEGY_CONTINUE_WORKFLOW
}

func StepWaitsForRetrySignal(step *workflowruntimev1.WorkflowStepProjection) bool {
	return step.GetFailureStrategy() == workflowruntimev1.WorkflowStepFailureStrategy_WORKFLOW_STEP_FAILURE_STRATEGY_WAIT_RETRY_SIGNAL
}

func normalizeStepFailureStrategy(
	strategy workflowruntimev1.WorkflowStepFailureStrategy,
) workflowruntimev1.WorkflowStepFailureStrategy {
	if strategy == workflowruntimev1.WorkflowStepFailureStrategy_WORKFLOW_STEP_FAILURE_STRATEGY_UNSPECIFIED {
		return workflowruntimev1.WorkflowStepFailureStrategy_WORKFLOW_STEP_FAILURE_STRATEGY_FAIL_WORKFLOW
	}

	return strategy
}

func failedStepStatus(
	strategy workflowruntimev1.WorkflowStepFailureStrategy,
) workflowruntimev1.WorkflowStepStatus {
	if strategy == workflowruntimev1.WorkflowStepFailureStrategy_WORKFLOW_STEP_FAILURE_STRATEGY_WAIT_RETRY_SIGNAL {
		return workflowruntimev1.WorkflowStepStatus_WORKFLOW_STEP_STATUS_WAITING_RETRY
	}

	return workflowruntimev1.WorkflowStepStatus_WORKFLOW_STEP_STATUS_FAILED
}

func maxAttempts(policy *workflowruntimev1.ActivityPolicy) int32 {
	if policy == nil || policy.GetRetryPolicy() == nil {
		return 0
	}

	return policy.GetRetryPolicy().GetMaximumAttempts()
}

func cloneStep(step *workflowruntimev1.WorkflowStepProjection) *workflowruntimev1.WorkflowStepProjection {
	if step == nil {
		return &workflowruntimev1.WorkflowStepProjection{}
	}

	return proto.Clone(step).(*workflowruntimev1.WorkflowStepProjection)
}

func cloneStepRef(ref *workflowruntimev1.WorkflowStepRef) *workflowruntimev1.WorkflowStepRef {
	if ref == nil {
		return nil
	}

	return proto.Clone(ref).(*workflowruntimev1.WorkflowStepRef)
}

func cloneActivityPolicy(policy *workflowruntimev1.ActivityPolicy) *workflowruntimev1.ActivityPolicy {
	if policy == nil {
		return nil
	}

	return proto.Clone(policy).(*workflowruntimev1.ActivityPolicy)
}

func cloneRuntimeError(err *workflowruntimev1.WorkflowRuntimeError) *workflowruntimev1.WorkflowRuntimeError {
	if err == nil {
		return nil
	}

	return proto.Clone(err).(*workflowruntimev1.WorkflowRuntimeError)
}
