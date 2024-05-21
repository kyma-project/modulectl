package filegenerator

import "errors"

var (
	ErrGettingDefaultContent = errors.New("error getting default content")
	ErrWritingFile           = errors.New("error writing file")
)
