package domain

import "errors"

var (
	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("conflict")
	ErrInvalid  = errors.New("invalid input")
)

type ValidationError struct {
	Message string
}

func (err ValidationError) Error() string {
	return err.Message
}

func (err ValidationError) Unwrap() error {
	return ErrInvalid
}
