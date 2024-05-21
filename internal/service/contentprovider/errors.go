package contentprovider

import "errors"

var (
	ErrMissingArg = errors.New("missing required argument")
	ErrInvalidArg = errors.New("invalid argument")
)
