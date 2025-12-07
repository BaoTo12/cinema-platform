package apperrors

import (
	"errors"
	"fmt"
	"net/http"
	"runtime"
)

// ErrorCode represents application error codes
type ErrorCode string

const (
	// General errors
	CodeInternal       ErrorCode = "INTERNAL_ERROR"
	CodeValidation     ErrorCode = "VALIDATION_ERROR"
	CodeNotFound       ErrorCode = "NOT_FOUND"
	CodeConflict       ErrorCode = "CONFLICT"
	CodeBadRequest     ErrorCode = "BAD_REQUEST"
	CodeUnauthorized   ErrorCode = "UNAUTHORIZED"
	CodeForbidden      ErrorCode = "FORBIDDEN"
	CodeTooManyRequests ErrorCode = "TOO_MANY_REQUESTS"

	// Auth specific errors
	CodeInvalidCredentials ErrorCode = "INVALID_CREDENTIALS"
	CodeTokenExpired       ErrorCode = "TOKEN_EXPIRED"
	CodeTokenInvalid       ErrorCode = "TOKEN_INVALID"
	CodeEmailNotVerified   ErrorCode = "EMAIL_NOT_VERIFIED"
	CodeAccountDisabled    ErrorCode = "ACCOUNT_DISABLED"

	// Resource specific errors
	CodeUserNotFound      ErrorCode = "USER_NOT_FOUND"
	CodeMovieNotFound     ErrorCode = "MOVIE_NOT_FOUND"
	CodeBookingNotFound   ErrorCode = "BOOKING_NOT_FOUND"
	CodeShowtimeNotFound  ErrorCode = "SHOWTIME_NOT_FOUND"
	CodeCinemaNotFound    ErrorCode = "CINEMA_NOT_FOUND"
	CodeSeatNotAvailable  ErrorCode = "SEAT_NOT_AVAILABLE"
	CodeEmailAlreadyExists ErrorCode = "EMAIL_ALREADY_EXISTS"

	// Business logic errors
	CodeBookingExpired    ErrorCode = "BOOKING_EXPIRED"
	CodePaymentFailed     ErrorCode = "PAYMENT_FAILED"
	CodeInvalidPromoCode  ErrorCode = "INVALID_PROMO_CODE"
	CodeSeatsAlreadyBooked ErrorCode = "SEATS_ALREADY_BOOKED"
)

// AppError represents an application error with context
type AppError struct {
	Code       ErrorCode `json:"code"`
	Message    string    `json:"message"`
	Details    any       `json:"details,omitempty"`
	HTTPStatus int       `json:"-"`
	Err        error     `json:"-"`
	Stack      string    `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the wrapped error
func (e *AppError) Unwrap() error {
	return e.Err
}

// WithDetails adds details to the error
func (e *AppError) WithDetails(details any) *AppError {
	e.Details = details
	return e
}

// WithError wraps an underlying error
func (e *AppError) WithError(err error) *AppError {
	e.Err = err
	return e
}

// captureStack captures the current stack trace
func captureStack() string {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])

	stack := ""
	for {
		frame, more := frames.Next()
		stack += fmt.Sprintf("%s\n\t%s:%d\n", frame.Function, frame.File, frame.Line)
		if !more {
			break
		}
	}
	return stack
}

// New creates a new AppError
func New(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: codeToHTTPStatus(code),
		Stack:      captureStack(),
	}
}

// Wrap wraps an existing error with an AppError
func Wrap(err error, code ErrorCode, message string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: codeToHTTPStatus(code),
		Err:        err,
		Stack:      captureStack(),
	}
}

// codeToHTTPStatus maps error codes to HTTP status codes
func codeToHTTPStatus(code ErrorCode) int {
	switch code {
	case CodeValidation, CodeBadRequest:
		return http.StatusBadRequest
	case CodeUnauthorized, CodeInvalidCredentials, CodeTokenExpired, CodeTokenInvalid:
		return http.StatusUnauthorized
	case CodeForbidden, CodeEmailNotVerified, CodeAccountDisabled:
		return http.StatusForbidden
	case CodeNotFound, CodeUserNotFound, CodeMovieNotFound, CodeBookingNotFound,
		CodeShowtimeNotFound, CodeCinemaNotFound:
		return http.StatusNotFound
	case CodeConflict, CodeEmailAlreadyExists, CodeSeatsAlreadyBooked:
		return http.StatusConflict
	case CodeTooManyRequests:
		return http.StatusTooManyRequests
	case CodeSeatNotAvailable, CodeBookingExpired, CodePaymentFailed, CodeInvalidPromoCode:
		return http.StatusUnprocessableEntity
	default:
		return http.StatusInternalServerError
	}
}

// Common error constructors

// ErrInternal creates an internal server error
func ErrInternal(message string) *AppError {
	return New(CodeInternal, message)
}

// ErrValidation creates a validation error
func ErrValidation(message string) *AppError {
	return New(CodeValidation, message)
}

// ErrNotFound creates a not found error
func ErrNotFound(resource string) *AppError {
	return New(CodeNotFound, fmt.Sprintf("%s not found", resource))
}

// ErrUnauthorized creates an unauthorized error
func ErrUnauthorized(message string) *AppError {
	return New(CodeUnauthorized, message)
}

// ErrForbidden creates a forbidden error
func ErrForbidden(message string) *AppError {
	return New(CodeForbidden, message)
}

// ErrConflict creates a conflict error
func ErrConflict(message string) *AppError {
	return New(CodeConflict, message)
}

// ErrBadRequest creates a bad request error
func ErrBadRequest(message string) *AppError {
	return New(CodeBadRequest, message)
}

// ErrInvalidCredentials creates an invalid credentials error
func ErrInvalidCredentials() *AppError {
	return New(CodeInvalidCredentials, "Invalid email or password")
}

// ErrTokenExpired creates a token expired error
func ErrTokenExpired() *AppError {
	return New(CodeTokenExpired, "Token has expired")
}

// ErrTokenInvalid creates an invalid token error
func ErrTokenInvalid() *AppError {
	return New(CodeTokenInvalid, "Invalid token")
}

// ErrEmailExists creates an email already exists error
func ErrEmailExists() *AppError {
	return New(CodeEmailAlreadyExists, "Email address is already registered")
}

// ErrAccountDisabled creates an account disabled error
func ErrAccountDisabled() *AppError {
	return New(CodeAccountDisabled, "Account has been disabled")
}

// Is checks if the error matches a specific error code
func Is(err error, code ErrorCode) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code == code
	}
	return false
}

// GetHTTPStatus gets the HTTP status from an error
func GetHTTPStatus(err error) int {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.HTTPStatus
	}
	return http.StatusInternalServerError
}

// GetCode gets the error code from an error
func GetCode(err error) ErrorCode {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code
	}
	return CodeInternal
}
