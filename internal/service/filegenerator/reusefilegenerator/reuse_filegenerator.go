package reusefilegenerator

import (
	"fmt"

	commonerrors "github.com/kyma-project/modulectl/internal/common/errors"
	"github.com/kyma-project/modulectl/internal/common/types"
	iotools "github.com/kyma-project/modulectl/tools/io"
)

type FileReader interface {
	FileExists(path string) (bool, error)
}

type FileGenerator interface {
	GenerateFile(out iotools.Out, path string, args types.KeyValueArgs) error
}

type Service struct {
	kind          string
	fileReader    FileReader
	fileGenerator FileGenerator
}

func NewService(
	kind string,
	fileSystem FileReader,
	fileGenerator FileGenerator,
) (*Service, error) {
	if kind == "" {
		return nil, fmt.Errorf("%w: kind must not be empty", commonerrors.ErrInvalidArg)
	}

	if fileSystem == nil {
		return nil, fmt.Errorf("%w: fileSystem must not be nil", commonerrors.ErrInvalidArg)
	}

	if fileGenerator == nil {
		return nil, fmt.Errorf("%w: fileGenerator must not be nil", commonerrors.ErrInvalidArg)
	}

	return &Service{
		kind:          kind,
		fileReader:    fileSystem,
		fileGenerator: fileGenerator,
	}, nil
}

func (s *Service) GenerateFile(out iotools.Out, path string, args types.KeyValueArgs) error {
	fileExists, err := s.fileReader.FileExists(path)
	if err != nil {
		return fmt.Errorf("%w %s: %w", ErrCheckingFileExistence, path, err)
	}

	if fileExists {
		out.Write(fmt.Sprintf("The %s file already exists, reusing: %s\n", s.kind, path))
		return nil
	}

	err = s.fileGenerator.GenerateFile(out, path, args)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrGeneratingFile, err)
	}

	return nil
}
