package apperrors

import (
	"errors"
	"fmt"
)

// Common application-level errors
var (
	ErrNotFound            = errors.New("resource not found")
	ErrAlreadyExists       = errors.New("resource already exists")
	ErrInvalidInput        = errors.New("invalid input provided")
	ErrInternal            = errors.New("internal server error")
	ErrUnauthorized        = errors.New("unauthorized access")
	ErrForbidden           = errors.New("access forbidden")
	ErrForeignKeyViolation = errors.New("foreign key not found")
)

type ValidationField string

const (
	INVALID_EMAIL    ValidationField = "INVALID_EMAIL"
	INVALID_NAME     ValidationField = "INVALID_NAME"
	INVALID_PASSWORD ValidationField = "INVALID_PASSWORD"
	INVALID_ID       ValidationField = "INVALID_ID"
	INVALID_DATE     ValidationField = "INVALID_DATE"
	INVALID_SETTING  ValidationField = "INVALID_SETTING"
	INVALID_INPUT    ValidationField = "INVALID_INPUT"
)

type ValidationError struct {
	Field   ValidationField
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error on field '%s': %s", e.Field, e.Message)
}

func NewValidationError(field ValidationField, message string) *ValidationError {
	return &ValidationError{Field: field, Message: message}
}

type ErrorCode string

const (
	NOT_FOUND      ErrorCode = "NOT_FOUND"
	ALREADY_EXISTS ErrorCode = "ALREADY_EXISTS"
	UNAUTHORIZED   ErrorCode = "UNAUTHORIZED"
	FORBIDDEN      ErrorCode = "FORBIDDEN"
	INTERNAL_ERROR ErrorCode = "INTERNAL_ERROR"
	BAD_REQUEST    ErrorCode = "BAD_REQUEST"
)
