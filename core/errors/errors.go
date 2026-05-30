package errors

import "net/http"

type AppError struct {
	httpStatus int
	code       int
	message    string
	cause      error
}

func New(httpStatus int, code int, message string, cause error) *AppError {
	if httpStatus == 0 {
		httpStatus = http.StatusInternalServerError
	}
	if code == 0 {
		code = httpStatus
	}
	if message == "" {
		message = http.StatusText(httpStatus)
	}
	return &AppError{
		httpStatus: httpStatus,
		code:       code,
		message:    message,
		cause:      cause,
	}
}

func (e *AppError) Error() string {
	if e.cause == nil {
		return e.message
	}
	return e.message + ": " + e.cause.Error()
}

func (e *AppError) Unwrap() error {
	return e.cause
}

func (e *AppError) HTTPStatus() int {
	return e.httpStatus
}

func (e *AppError) Code() int {
	return e.code
}

func (e *AppError) PublicMessage() string {
	return e.message
}
