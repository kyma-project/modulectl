package scaffold

import "errors"

var (
	ErrGeneratingFile  = errors.New("error generating file")
	ErrOverwritingFile = errors.New("error overwriting file")
)
