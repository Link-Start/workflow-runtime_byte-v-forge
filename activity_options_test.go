package workflowruntime

import (
	"testing"
	"time"
)

func TestDefaultActivityOptions(t *testing.T) {
	options := DefaultActivityOptions()
	if options.StartToCloseTimeout != 2*time.Minute {
		t.Fatalf("StartToCloseTimeout = %s, want 2m", options.StartToCloseTimeout)
	}
	if options.ScheduleToCloseTimeout != 10*time.Minute {
		t.Fatalf("ScheduleToCloseTimeout = %s, want 10m", options.ScheduleToCloseTimeout)
	}
	if options.HeartbeatTimeout != 30*time.Second {
		t.Fatalf("HeartbeatTimeout = %s, want 30s", options.HeartbeatTimeout)
	}
	if options.RetryPolicy == nil {
		t.Fatal("RetryPolicy should be set")
	}
}
