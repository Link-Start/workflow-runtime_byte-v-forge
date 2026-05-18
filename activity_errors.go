package workflowruntime

import "go.temporal.io/sdk/temporal"

func NewRetryableActivityError(code, message string, cause error, details ...any) error {
	return temporal.NewApplicationErrorWithCause(message, code, cause, details...)
}

func NewNonRetryableActivityError(code, message string, cause error, details ...any) error {
	return temporal.NewNonRetryableApplicationError(message, code, cause, details...)
}
