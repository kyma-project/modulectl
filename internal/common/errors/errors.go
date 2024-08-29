package errors

import "errors"

var (
	ErrInvalidArg    = errors.New("invalid argument")
	ErrInvalidOption = errors.New("invalid Option")
)
