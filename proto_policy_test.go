package workflowruntime

import (
	"testing"
	"time"
)

func TestRetryPolicyProtoRoundTrip(t *testing.T) {
	policy := DefaultRetryPolicy()
	policy.NonRetryableErrorTypes = []string{"validation_failed"}

	roundTripped := FromProtoRetryPolicy(ToProtoRetryPolicy(policy))
	if roundTripped.InitialInterval != time.Second {
		t.Fatalf("InitialInterval = %s, want 1s", roundTripped.InitialInterval)
	}
	if roundTripped.MaximumInterval != 30*time.Second {
		t.Fatalf("MaximumInterval = %s, want 30s", roundTripped.MaximumInterval)
	}
	if roundTripped.MaximumAttempts != 5 {
		t.Fatalf("MaximumAttempts = %d, want 5", roundTripped.MaximumAttempts)
	}
	if len(roundTripped.NonRetryableErrorTypes) != 1 || roundTripped.NonRetryableErrorTypes[0] != "validation_failed" {
		t.Fatalf("NonRetryableErrorTypes = %#v, want validation_failed", roundTripped.NonRetryableErrorTypes)
	}
}

func TestActivityPolicyProtoRoundTrip(t *testing.T) {
	options := DefaultActivityOptions()

	roundTripped := FromProtoActivityPolicy(ToProtoActivityPolicy(options))
	if roundTripped.StartToCloseTimeout != options.StartToCloseTimeout {
		t.Fatalf("StartToCloseTimeout = %s, want %s", roundTripped.StartToCloseTimeout, options.StartToCloseTimeout)
	}
	if roundTripped.ScheduleToCloseTimeout != options.ScheduleToCloseTimeout {
		t.Fatalf("ScheduleToCloseTimeout = %s, want %s", roundTripped.ScheduleToCloseTimeout, options.ScheduleToCloseTimeout)
	}
	if roundTripped.HeartbeatTimeout != options.HeartbeatTimeout {
		t.Fatalf("HeartbeatTimeout = %s, want %s", roundTripped.HeartbeatTimeout, options.HeartbeatTimeout)
	}
	if roundTripped.RetryPolicy == nil {
		t.Fatal("RetryPolicy should be set")
	}
}

func TestToProtoTaskQueueRef(t *testing.T) {
	ref := ToProtoTaskQueueRef(Config{
		Namespace: "register",
		TaskQueue: "example-worker",
	}, "example-worker")
	if ref.GetNamespace() != "register" {
		t.Fatalf("Namespace = %q, want register", ref.GetNamespace())
	}
	if ref.GetName() != "example-worker" {
		t.Fatalf("Name = %q, want example-worker", ref.GetName())
	}
	if ref.GetOwningService() != "example-worker" {
		t.Fatalf("OwningService = %q, want example-worker", ref.GetOwningService())
	}
}
