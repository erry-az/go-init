package domain

import (
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// DomainError represents different types of domain errors
type DomainError struct {
	Type    ErrorType
	Message string
	Cause   error
}

type ErrorType int

const (
	ErrorTypeValidation ErrorType = iota
	ErrorTypeNotFound
	ErrorTypeConflict
	ErrorTypeInternal
	ErrorTypeUnauthorized
	ErrorTypeForbidden
)

func (e *DomainError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

func (e *DomainError) Unwrap() error {
	return e.Cause
}

// ToGRPCError converts domain error to gRPC status error
func (e *DomainError) ToGRPCError() error {
	switch e.Type {
	case ErrorTypeValidation:
		return status.Error(codes.InvalidArgument, e.Message)
	case ErrorTypeNotFound:
		return status.Error(codes.NotFound, e.Message)
	case ErrorTypeConflict:
		return status.Error(codes.AlreadyExists, e.Message)
	case ErrorTypeUnauthorized:
		return status.Error(codes.Unauthenticated, e.Message)
	case ErrorTypeForbidden:
		return status.Error(codes.PermissionDenied, e.Message)
	case ErrorTypeInternal:
		return status.Error(codes.Internal, e.Message)
	default:
		return status.Error(codes.Internal, e.Message)
	}
}

// Error constructors
func NewValidationError(message string) *DomainError {
	return &DomainError{
		Type:    ErrorTypeValidation,
		Message: message,
	}
}

func NewValidationErrorWithCause(message string, cause error) *DomainError {
	return &DomainError{
		Type:    ErrorTypeValidation,
		Message: message,
		Cause:   cause,
	}
}

func NewNotFoundError(message string) *DomainError {
	return &DomainError{
		Type:    ErrorTypeNotFound,
		Message: message,
	}
}

func NewConflictError(message string) *DomainError {
	return &DomainError{
		Type:    ErrorTypeConflict,
		Message: message,
	}
}

func NewInternalError(message string) *DomainError {
	return &DomainError{
		Type:    ErrorTypeInternal,
		Message: message,
	}
}

func NewInternalErrorWithCause(message string, cause error) *DomainError {
	return &DomainError{
		Type:    ErrorTypeInternal,
		Message: message,
		Cause:   cause,
	}
}

func NewUnauthorizedError(message string) *DomainError {
	return &DomainError{
		Type:    ErrorTypeUnauthorized,
		Message: message,
	}
}

func NewForbiddenError(message string) *DomainError {
	return &DomainError{
		Type:    ErrorTypeForbidden,
		Message: message,
	}
}