package apperror

import (
	"fmt"
	"net/http"
)

type AppError struct {
	HTTPCode int    // for user
	Message  string // for user
	Err      error  // for internal logging
}

func (e *AppError) Error() string {
	return fmt.Sprintf("HTTP %d: %s - %v", e.HTTPCode, e.Message, e.Err)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// NewBadReq used to create errors with
// statusCode = 400.
func NewBadReq(message string, err error) *AppError {
	return &AppError{HTTPCode: http.StatusBadRequest, Message: message, Err: fmt.Errorf("%s: %w", message, err)}
}

// NewUnauthorized used to create errors with
// statusCode = 401.
func NewUnauthorized(message string, err error) *AppError {
	return &AppError{HTTPCode: http.StatusUnauthorized, Message: message, Err: fmt.Errorf("%s: %w", message, err)}
}

// NewNotFound used to create errors with
// statusCode = 404.
func NewNotFound(message string, err error) *AppError {
	return &AppError{HTTPCode: http.StatusNotFound, Message: message, Err: fmt.Errorf("%s: %w", message, err)}
}

// NewInternal used to craete errors with
// statusCode = 500.
func NewInternal(message string, err error) *AppError {
	return &AppError{HTTPCode: http.StatusInternalServerError, Message: message, Err: fmt.Errorf("%s: %w", message, err)}
}
