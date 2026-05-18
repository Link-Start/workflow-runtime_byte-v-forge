package workflowruntime

import (
	"testing"

	workflowruntimev1 "github.com/byte-v-forge/contracts-go/byte/v/forge/contracts/workflowruntime/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestStepProjectionFailureCanContinue(t *testing.T) {
	policy := ToProtoActivityPolicy(DefaultActivityOptions())
	step := NewStepProjection(&workflowruntimev1.WorkflowStepRef{
		StepId:       "check-payment",
		DisplayName:  "Check payment",
		ActivityType: "gpt.activation.check-payment.v1",
		Sequence:     1,
	}, workflowruntimev1.WorkflowStepFailureStrategy_WORKFLOW_STEP_FAILURE_STRATEGY_CONTINUE_WORKFLOW, policy)

	running := MarkStepRunning(step, 1, timestamppb.Now())
	failed := MarkStepFailed(running, NewRuntimeError("payment_not_ready", "payment is not ready", true), timestamppb.Now())

	if got := failed.GetStatus(); got != workflowruntimev1.WorkflowStepStatus_WORKFLOW_STEP_STATUS_FAILED {
		t.Fatalf("status = %v, want failed", got)
	}
	if !StepContinuesAfterFailure(failed) {
		t.Fatal("StepContinuesAfterFailure() returned false")
	}
	if got := len(failed.GetAttempts()); got != 1 {
		t.Fatalf("attempt count = %d, want 1", got)
	}
}

func TestStepProjectionFailureCanWaitForRetrySignal(t *testing.T) {
	step := NewStepProjection(&workflowruntimev1.WorkflowStepRef{
		StepId: "manual-solve",
	}, workflowruntimev1.WorkflowStepFailureStrategy_WORKFLOW_STEP_FAILURE_STRATEGY_WAIT_RETRY_SIGNAL, nil)

	failed := MarkStepFailed(
		MarkStepRunning(step, 1, timestamppb.Now()),
		NewRuntimeError("captcha_failed", "captcha failed", true),
		timestamppb.Now(),
	)

	if got := failed.GetStatus(); got != workflowruntimev1.WorkflowStepStatus_WORKFLOW_STEP_STATUS_WAITING_RETRY {
		t.Fatalf("status = %v, want waiting retry", got)
	}
	if !StepWaitsForRetrySignal(failed) {
		t.Fatal("StepWaitsForRetrySignal() returned false")
	}
}

func TestStepProjectionDefaultsToFailWorkflow(t *testing.T) {
	step := NewStepProjection(&workflowruntimev1.WorkflowStepRef{StepId: "create-account"}, 0, nil)

	if got := step.GetFailureStrategy(); got != workflowruntimev1.WorkflowStepFailureStrategy_WORKFLOW_STEP_FAILURE_STRATEGY_FAIL_WORKFLOW {
		t.Fatalf("failure strategy = %v, want fail workflow", got)
	}
}
