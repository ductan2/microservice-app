package repository

import (
	"errors"
	"fmt"
)

// Common repository errors
var (
	// ErrEntityNotFound is returned when an entity is not found
	ErrEntityNotFound = errors.New("entity not found")
	// ErrDuplicateEntity is returned when trying to create a duplicate entity
	ErrDuplicateEntity = errors.New("duplicate entity")
	// ErrInvalidID is returned when an invalid ID is provided
	ErrInvalidID = errors.New("invalid ID")
	// ErrInvalidFilter is returned when an invalid filter is provided
	ErrInvalidFilter = errors.New("invalid filter")
	// ErrInvalidSort is returned when an invalid sort option is provided
	ErrInvalidSort = errors.New("invalid sort option")
	// ErrOperationNotAllowed is returned when an operation is not allowed
	ErrOperationNotAllowed = errors.New("operation not allowed")
	// ErrDatabaseConnection is returned when there's a database connection issue
	ErrDatabaseConnection = errors.New("database connection error")
	// ErrTimeout is returned when an operation times out
	ErrTimeout = errors.New("operation timeout")
	// ErrAlreadyDeleted is returned when trying to delete an already deleted entity
	ErrAlreadyDeleted = errors.New("entity already deleted")
)

// Lesson-specific repository errors
var (
	ErrFlashcardSetNotFound = errors.New("flashcard set not found")
	ErrFlashcardNotFound    = errors.New("flashcard not found")
)

// RepositoryError wraps errors with additional context
type RepositoryError struct {
	Operation string
	Entity    string
	Err       error
}

func (e *RepositoryError) Error() string {
	return fmt.Sprintf("repository error during %s on %s: %v", e.Operation, e.Entity, e.Err)
}

func (e *RepositoryError) Unwrap() error {
	return e.Err
}

// NewRepositoryError creates a new repository error
func NewRepositoryError(operation, entity string, err error) *RepositoryError {
	return &RepositoryError{
		Operation: operation,
		Entity:    entity,
		Err:       err,
	}
}

// IsNotFoundError checks if the error is a not found error
func IsNotFoundError(err error) bool {
	return errors.Is(err, ErrEntityNotFound) ||
		errors.Is(err, ErrFlashcardSetNotFound) ||
		errors.Is(err, ErrFlashcardNotFound)
}

// IsDuplicateError checks if the error is a duplicate error
func IsDuplicateError(err error) bool {
	return errors.Is(err, ErrDuplicateEntity)
}

// IsValidationError checks if the error is a validation error
func IsValidationError(err error) bool {
	return errors.Is(err, ErrInvalidID) ||
		errors.Is(err, ErrInvalidFilter) ||
		errors.Is(err, ErrInvalidSort)
}

// IsConnectionError checks if the error is a connection error
func IsConnectionError(err error) bool {
	return errors.Is(err, ErrDatabaseConnection)
}