package workflowruntime

import (
	"os"
	"strings"

	"go.temporal.io/sdk/client"
	sdklog "go.temporal.io/sdk/log"
)

const (
	EnvTemporalAddress   = "TEMPORAL_ADDRESS"
	EnvTemporalNamespace = "TEMPORAL_NAMESPACE"
	EnvTemporalTaskQueue = "TEMPORAL_TASK_QUEUE"
	EnvTemporalIdentity  = "TEMPORAL_IDENTITY"
)

type Config struct {
	Address   string
	Namespace string
	TaskQueue string
	Identity  string
}

type ClientOption func(*client.Options)

func LoadConfigFromEnv(getenv func(string) string) (Config, error) {
	if getenv == nil {
		getenv = os.Getenv
	}
	config := Config{
		Address:   strings.TrimSpace(getenv(EnvTemporalAddress)),
		Namespace: strings.TrimSpace(getenv(EnvTemporalNamespace)),
		TaskQueue: strings.TrimSpace(getenv(EnvTemporalTaskQueue)),
		Identity:  strings.TrimSpace(getenv(EnvTemporalIdentity)),
	}
	if err := config.Validate(); err != nil {
		return Config{}, err
	}
	return config, nil
}

func (c Config) Validate() error {
	if c.Address == "" {
		return validationError(EnvTemporalAddress, "is required")
	}
	if c.Namespace == "" {
		return validationError(EnvTemporalNamespace, "is required")
	}
	return ValidateTaskQueueName(c.TaskQueue)
}

func NewClientOptions(config Config, options ...ClientOption) (client.Options, error) {
	if err := config.Validate(); err != nil {
		return client.Options{}, err
	}
	clientOptions := client.Options{
		HostPort:  config.Address,
		Namespace: config.Namespace,
		Identity:  config.Identity,
	}
	for _, option := range options {
		if option != nil {
			option(&clientOptions)
		}
	}
	return clientOptions, nil
}

func WithLogger(logger sdklog.Logger) ClientOption {
	return func(options *client.Options) {
		options.Logger = logger
	}
}

func Dial(config Config, options ...ClientOption) (client.Client, error) {
	clientOptions, err := NewClientOptions(config, options...)
	if err != nil {
		return nil, err
	}
	return client.Dial(clientOptions)
}
