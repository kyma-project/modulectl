package scaffold

import "errors"

var (
	ErrInvalidOption  = errors.New("invalid option")
	ErrGeneratingFile = errors.New("error generating file")
)
