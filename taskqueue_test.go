package workflowruntime

import "testing"

func TestNormalizeTaskQueueName(t *testing.T) {
	name, err := NormalizeTaskQueueName("Example_Registration")
	if err != nil {
		t.Fatalf("NormalizeTaskQueueName() error = %v", err)
	}
	if name != "example-registration" {
		t.Fatalf("name = %q, want example-registration", name)
	}
}

func TestValidateTaskQueueName(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{name: "example-worker"},
		{name: "example-registration"},
		{name: "ab", wantErr: true},
		{name: "-example-registration", wantErr: true},
		{name: "example-registration-", wantErr: true},
		{name: "ExampleRegister", wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateTaskQueueName(test.name)
			if test.wantErr && err == nil {
				t.Fatal("ValidateTaskQueueName() expected error")
			}
			if !test.wantErr && err != nil {
				t.Fatalf("ValidateTaskQueueName() error = %v", err)
			}
		})
	}
}
