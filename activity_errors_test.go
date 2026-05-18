package workflowruntime

import (
	"errors"
	"testing"

	"go.temporal.io/sdk/temporal"
)

func TestActivityErrorHelpers(t *testing.T) {
	retryable := NewRetryableActivityError("dependency_unavailable", "dependency unavailable", errors.New("upstream"))
	if _, ok := retryable.(*temporal.ApplicationError); !ok {
		t.Fatalf("retryable error type = %T, want *temporal.ApplicationError", retryable)
	}

	nonRetryable := NewNonRetryableActivityError("validation_failed", "validation failed", nil)
	appErr, ok := nonRetryable.(*temporal.ApplicationError)
	if !ok {
		t.Fatalf("non retryable error type = %T, want *temporal.ApplicationError", nonRetryable)
	}
	if !appErr.NonRetryable() {
		t.Fatal("non retryable error should be marked non-retryable")
	}
}
