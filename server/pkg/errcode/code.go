package errcode

// Code is the canonical application error code.
type Code int

const (
	// OK indicates success.
	OK Code = 0
	// Internal indicates an internal server failure.
	Internal Code = 1000
	// InvalidArgument indicates invalid request parameters.
	InvalidArgument Code = 1001
	// Unauthorized indicates authentication failure.
	Unauthorized Code = 1002
	// Forbidden indicates authorization failure.
	Forbidden Code = 1003
	// NotFound indicates requested resource not found.
	NotFound Code = 1004
	// RateLimited indicates the client exceeded a request rate limit.
	RateLimited Code = 1005
	// InstallTokenInvalid indicates install token validation failed.
	InstallTokenInvalid Code = 2001
	// AgentUnauthorized indicates agent authentication failure.
	AgentUnauthorized Code = 2002
	// EnrollmentTokenInvalid indicates node enrollment token validation failed.
	EnrollmentTokenInvalid Code = 2003
)

// AppError is the structured application error.
type AppError struct {
	Code    Code
	Message string
	Err     error
}

// Error implements error.
func (e *AppError) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

// Unwrap returns the wrapped cause.
func (e *AppError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

// New creates a new AppError.
func New(code Code, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

// Wrap creates a new AppError with a wrapped cause.
func Wrap(code Code, message string, err error) *AppError {
	return &AppError{Code: code, Message: message, Err: err}
}
