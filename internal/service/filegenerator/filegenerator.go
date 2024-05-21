package filegenerator

import (
	"errors"
	"fmt"

	commonerrors "github.com/kyma-project/modulectl/internal/common/errors"
	"github.com/kyma-project/modulectl/internal/common/types"
	"github.com/kyma-project/modulectl/tools/io"
)

type FileWriter interface {
	WriteFile(path, content string) error
}

type DefaultContentProvider interface {
	GetDefaultContent(args types.KeyValueArgs) (string, error)
}

type Service struct {
	kind                   string
	fileWriter             FileWriter
	defaultContentProvider DefaultContentProvider
}

func NewService(kind string, fileSystem FileWriter, defaultContentProvider DefaultContentProvider) (*Service, error) {
	if kind == "" {
		return nil, fmt.Errorf("%w: kind must not be empty", commonerrors.ErrInvalidArg)
	}

	if fileSystem == nil {
		return nil, fmt.Errorf("%w: fileSystem must not be nil", commonerrors.ErrInvalidArg)

	}

	if defaultContentProvider == nil {
		return nil, fmt.Errorf("%w: defaultContentProvider must not be nil", commonerrors.ErrInvalidArg)
	}

	return &Service{
		kind:                   kind,
		fileWriter:             fileSystem,
		defaultContentProvider: defaultContentProvider,
	}, nil
}

func (s *Service) GenerateFile(out io.Out, path string, args types.KeyValueArgs) error {
	defaultContent, err := s.defaultContentProvider.GetDefaultContent(args)
	if err != nil {
		return errors.Join(ErrGettingDefaultContent, err)
	}

	if err := s.writeFile(defaultContent, path); err != nil {
		return err
	}

	out.Write(fmt.Sprintf("Generated a blank %s file: %s\n", s.kind, path))

	return nil
}

func (s *Service) writeFile(content, path string) error {
	if err := s.fileWriter.WriteFile(path, content); err != nil {
		return fmt.Errorf("%w %s: %w", ErrWritingFile, path, err)
	}

	return nil
}
