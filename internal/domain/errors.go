package domain

import (
	"errors"
	"strings"
)

var (
	ErrMissingRequiredField    = errors.New("missing required field")
	ErrInvalidTicketStatus     = errors.New("invalid ticket status")
	ErrInvalidTicketPriority   = errors.New("invalid ticket priority")
	ErrInvalidUserRole         = errors.New("invalid user role")
	ErrInvalidStatusTransition = errors.New("invalid ticket status transition")
)

type FieldError struct {
	Field string
	Err   error
}

func (e FieldError) Error() string {
	if e.Field == "" {
		return e.Err.Error()
	}

	return e.Field + ": " + e.Err.Error()
}

func (e FieldError) Unwrap() error {
	return e.Err
}

func required(field string, missing bool) error {
	if !missing {
		return nil
	}

	return FieldError{Field: field, Err: ErrMissingRequiredField}
}

func blank(value string) bool {
	return strings.TrimSpace(value) == ""
}
