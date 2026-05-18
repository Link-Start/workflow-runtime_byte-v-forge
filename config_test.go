package workflowruntime

import "testing"

func TestLoadConfigFromEnv(t *testing.T) {
	values := map[string]string{
		EnvTemporalAddress:   "temporal:7233",
		EnvTemporalNamespace: "register",
		EnvTemporalTaskQueue: "example-registration",
		EnvTemporalIdentity:  "worker-1",
	}
	config, err := LoadConfigFromEnv(func(key string) string {
		return values[key]
	})
	if err != nil {
		t.Fatalf("LoadConfigFromEnv() error = %v", err)
	}
	if config.Address != "temporal:7233" {
		t.Fatalf("Address = %q, want temporal:7233", config.Address)
	}
	if config.Namespace != "register" {
		t.Fatalf("Namespace = %q, want register", config.Namespace)
	}
	if config.TaskQueue != "example-registration" {
		t.Fatalf("TaskQueue = %q, want example-registration", config.TaskQueue)
	}
	if config.Identity != "worker-1" {
		t.Fatalf("Identity = %q, want worker-1", config.Identity)
	}
}

func TestLoadConfigFromEnvRejectsMissingRequiredValues(t *testing.T) {
	_, err := LoadConfigFromEnv(func(string) string {
		return ""
	})
	if err == nil {
		t.Fatal("LoadConfigFromEnv() expected error")
	}
}

func TestNewClientOptions(t *testing.T) {
	options, err := NewClientOptions(Config{
		Address:   "temporal:7233",
		Namespace: "register",
		TaskQueue: "example-worker",
		Identity:  "worker-1",
	})
	if err != nil {
		t.Fatalf("NewClientOptions() error = %v", err)
	}
	if options.HostPort != "temporal:7233" {
		t.Fatalf("HostPort = %q, want temporal:7233", options.HostPort)
	}
	if options.Namespace != "register" {
		t.Fatalf("Namespace = %q, want register", options.Namespace)
	}
	if options.Identity != "worker-1" {
		t.Fatalf("Identity = %q, want worker-1", options.Identity)
	}
}
