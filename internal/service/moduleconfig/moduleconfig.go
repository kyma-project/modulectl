package moduleconfig

import (
	"fmt"
	"path"

	"github.com/kyma-project/modulectl/internal/common/errors"
	"github.com/kyma-project/modulectl/internal/common/types"
	"github.com/kyma-project/modulectl/tools/io"
)

type FileSystem interface {
	FileExists(path string) (bool, error)
}

type FileGenerator interface {
	GenerateFile(out io.Out, path string, args types.KeyValueArgs) error
}

type Service struct {
	fileSystem    FileSystem
	fileGenerator FileGenerator
}

func NewService(fileSystem FileSystem, fileGenerator FileGenerator) (*Service, error) {
	if fileSystem == nil {
		return nil, fmt.Errorf("%w: fileSystem must not be nil", errors.ErrInvalidArg)

	}

	if fileGenerator == nil {
		return nil, fmt.Errorf("%w: fileGenerator must not be nil", errors.ErrInvalidArg)
	}

	return &Service{
		fileSystem:    fileSystem,
		fileGenerator: fileGenerator,
	}, nil
}

func (s *Service) ForceExplicitOverwrite(directory, fileName string, overwrite bool) error {
	exists, err := s.fileSystem.FileExists(path.Join(directory, fileName))
	if err != nil {
		return err
	}

	if exists && !overwrite {
		return ErrFileExists
	}

	return nil
}

func (s *Service) GenerateFile(out io.Out, path string, args types.KeyValueArgs) error {
	return s.fileGenerator.GenerateFile(out, path, args)
}
