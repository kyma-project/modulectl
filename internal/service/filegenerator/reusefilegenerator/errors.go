package reusefilegenerator

import "errors"

var (
	ErrCheckingFileExistence = errors.New("error checking file existence")
	ErrGeneratingFile        = errors.New("error generating file")
)
