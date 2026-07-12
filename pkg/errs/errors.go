package errs

import "errors"

type Kind string

const (
	KindInvalid      Kind = "invalid"
	KindNotFound     Kind = "not_found"
	KindConflict     Kind = "conflict"
	KindForbidden    Kind = "forbidden"
	KindUnauthorized Kind = "unauthorized"
	KindInternal     Kind = "internal"
)

type Error struct {
	Kind    Kind              `json:"-"`
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

func (e *Error) Error() string { return e.Message }

func Invalid(code, message string, fields map[string]string) *Error {
	return &Error{Kind: KindInvalid, Code: code, Message: message, Fields: fields}
}

func NotFound(code, message string) *Error {
	return &Error{Kind: KindNotFound, Code: code, Message: message}
}

func Conflict(code, message string) *Error {
	return &Error{Kind: KindConflict, Code: code, Message: message}
}

func Forbidden(code, message string) *Error {
	return &Error{Kind: KindForbidden, Code: code, Message: message}
}

func Unauthorized(code, message string) *Error {
	return &Error{Kind: KindUnauthorized, Code: code, Message: message}
}

func Internal(code, message string) *Error {
	if message == "" {
		message = "Something went wrong. Please try again."
	}
	if code == "" {
		code = "internal"
	}
	return &Error{Kind: KindInternal, Code: code, Message: message}
}

func As(err error) (*Error, bool) {
	var target *Error
	ok := errors.As(err, &target)
	return target, ok
}
