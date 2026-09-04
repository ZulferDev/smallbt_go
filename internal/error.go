package internal

import "fmt"

// ErrorType categorizes errors for better error handling and logging.
type ErrorType string

const (
	ErrorTypeValidation  ErrorType = "validation"
	ErrorTypeNotFound    ErrorType = "not_found"
	ErrorTypeInvalid     ErrorType = "invalid"
	ErrorTypeDuplicate   ErrorType = "duplicate"
	ErrorTypeCircular    ErrorType = "circular"
	ErrorTypeExecution   ErrorType = "execution"
	ErrorTypeDataFeed    ErrorType = "data_feed"
	ErrorTypeComputation ErrorType = "computation"
	ErrorTypeConfig      ErrorType = "config"
	ErrorTypeInternal    ErrorType = "internal"
)

// AppError represents an application-level error with context.
type AppError struct {
	Type          ErrorType
	Message       string
	Path          string // Configuration path or context
	Cause         error  // Wrapped error
	IsRecoverable bool
}

func (e *AppError) Error() string {
	if e.Path != "" {
		return fmt.Sprintf("%s: %s (at %s)", e.Type, e.Message, e.Path)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Cause
}

// NewValidationError creates a validation error.
func NewValidationError(message, path string) *AppError {
	return &AppError{
		Type:          ErrorTypeValidation,
		Message:       message,
		Path:          path,
		IsRecoverable: true,
	}
}

// NewNotFoundError creates a not found error.
func NewNotFoundError(message, path string) *AppError {
	return &AppError{
		Type:          ErrorTypeNotFound,
		Message:       message,
		Path:          path,
		IsRecoverable: true,
	}
}

// NewInvalidError creates an invalid error.
func NewInvalidError(message, path string, cause error) *AppError {
	return &AppError{
		Type:          ErrorTypeInvalid,
		Message:       message,
		Path:          path,
		Cause:         cause,
		IsRecoverable: true,
	}
}

// NewCircularError creates a circular dependency error.
func NewCircularError(message, path string) *AppError {
	return &AppError{
		Type:          ErrorTypeCircular,
		Message:       message,
		Path:          path,
		IsRecoverable: false,
	}
}

// NewExecutionError creates an execution error.
func NewExecutionError(message string, cause error) *AppError {
	return &AppError{
		Type:          ErrorTypeExecution,
		Message:       message,
		Cause:         cause,
		IsRecoverable: false,
	}
}
