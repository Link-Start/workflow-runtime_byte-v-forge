package workflowruntime

import "testing"

func TestWorkerSpecValidateRequiresTaskQueue(t *testing.T) {
	spec := WorkerSpec{
		Workflows: []WorkflowDefinition{{Name: "test.workflow.v1", Definition: func() {}}},
	}
	if err := spec.Validate(); err == nil {
		t.Fatal("Validate() expected error")
	}
}

func TestWorkerSpecValidateRequiresStableRegistrationName(t *testing.T) {
	spec := WorkerSpec{
		TaskQueue: "example-registration",
		Workflows: []WorkflowDefinition{{
			Definition: func() {},
		}},
	}
	if err := spec.Validate(); err == nil {
		t.Fatal("Validate() expected error")
	}
}

func TestWorkerSpecValidateAcceptsWorkflowOnlyWorker(t *testing.T) {
	spec := WorkerSpec{
		TaskQueue: "example-registration",
		Workflows: []WorkflowDefinition{{
			Name:       "example-registration.workflow.v1",
			Definition: func() {},
		}},
	}
	if err := spec.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}
