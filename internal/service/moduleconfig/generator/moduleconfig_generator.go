package moduleconfiggenerator

import (
	"fmt"
	"path"

	commonerrors "github.com/kyma-project/modulectl/internal/common/errors"
	"github.com/kyma-project/modulectl/internal/common/types"
	"github.com/kyma-project/modulectl/internal/service/moduleconfig"
	iotools "github.com/kyma-project/modulectl/tools/io"
)

type FileSystem interface {
	FileExists(path string) (bool, error)
}

type FileGenerator interface {
	GenerateFile(out iotools.Out, path string, args types.KeyValueArgs) error
}

type Service struct {
	fileSystem    FileSystem
	fileGenerator FileGenerator
}

func NewService(fileSystem FileSystem, fileGenerator FileGenerator) (*Service, error) {
	if fileSystem == nil {
		return nil, fmt.Errorf("fileSystem must not be nil: %w", commonerrors.ErrInvalidArg)
	}

	if fileGenerator == nil {
		return nil, fmt.Errorf("fileGenerator must not be nil: %w", commonerrors.ErrInvalidArg)
	}

	return &Service{
		fileSystem:    fileSystem,
		fileGenerator: fileGenerator,
	}, nil
}

func (s *Service) ForceExplicitOverwrite(directory, fileName string, overwrite bool) error {
	exists, err := s.fileSystem.FileExists(path.Join(directory, fileName))
	if err != nil {
		return fmt.Errorf("failed to check if file exists: %w", err)
	}

	if exists && !overwrite {
		return moduleconfig.ErrFileExists
	}

	return nil
}

func (s *Service) GenerateFile(out iotools.Out, path string, args types.KeyValueArgs) error {
	if err := s.fileGenerator.GenerateFile(out, path, args); err != nil {
		return fmt.Errorf("failed to generate file: %w", err)
	}
	return nil
}
