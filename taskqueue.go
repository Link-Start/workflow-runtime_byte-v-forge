package workflowruntime

import (
	"regexp"
	"strings"
)

var taskQueueNamePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,62}[a-z0-9]$`)

func NormalizeTaskQueueName(service string) (string, error) {
	name := strings.TrimSpace(strings.ToLower(service))
	name = strings.ReplaceAll(name, "_", "-")
	if err := ValidateTaskQueueName(name); err != nil {
		return "", err
	}
	return name, nil
}

func ValidateTaskQueueName(name string) error {
	if name == "" {
		return validationError("task_queue", "is required")
	}
	if !taskQueueNamePattern.MatchString(name) {
		return validationError("task_queue", "must be 3 to 64 lowercase letters, digits, or hyphens, and must start and end with a letter or digit")
	}
	return nil
}
