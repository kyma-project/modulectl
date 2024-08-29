package scaffold

import "errors"

var (
	ErrGeneratingFile  = errors.New("error generating file")
	ErrOverwritingFile = errors.New("%w: error overwriting file")
)
