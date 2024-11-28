package reusefilegenerator

import (
	"fmt"

	commonerrors "github.com/kyma-project/modulectl/internal/common/errors"
	"github.com/kyma-project/modulectl/internal/common/types"
	iotools "github.com/kyma-project/modulectl/tools/io"
)

type FileReader interface {
	FileExists(path string) (bool, error)
	ReadFile(path string) ([]byte, error)
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
		return nil, fmt.Errorf("kind must not be empty: %w", commonerrors.ErrInvalidArg)
	}

	if fileSystem == nil {
		return nil, fmt.Errorf("fileSystem must not be nil: %w", commonerrors.ErrInvalidArg)
	}

	if fileGenerator == nil {
		return nil, fmt.Errorf("fileGenerator must not be nil: %w", commonerrors.ErrInvalidArg)
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
		return fmt.Errorf("the '%s' file path: %w: %w", path, ErrCheckingFileExistence, err)
	}

	if fileExists {
		out.Write(fmt.Sprintf("the '%s' file path already exists, reusing: '%s'\n", s.kind, path))
		return nil
	}

	err = s.fileGenerator.GenerateFile(out, path, args)
	if err != nil {
		return fmt.Errorf("the '%s' file path: %w: %w", path, ErrGeneratingFile, err)
	}

	return nil
}
