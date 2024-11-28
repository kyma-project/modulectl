package scaffold

import (
	"errors"
	"fmt"
	"os"

	commonerrors "github.com/kyma-project/modulectl/internal/common/errors"
	"github.com/kyma-project/modulectl/internal/common/validation"
	iotools "github.com/kyma-project/modulectl/tools/io"
)

type Options struct {
	Out                       iotools.Out
	Directory                 string
	ModuleConfigFileName      string
	ModuleConfigFileOverwrite bool
	ManifestFileName          string
	DefaultCRFileName         string
	SecurityConfigFileName    string
	ModuleName                string
	ModuleVersion             string
}

func (opts Options) Validate() error {
	if opts.Out == nil {
		return fmt.Errorf("opts.Out must not be nil: %w", commonerrors.ErrInvalidOption)
	}

	if err := opts.validateModuleName(); err != nil {
		return err
	}

	if err := opts.validateDirectory(); err != nil {
		return err
	}

	if err := opts.validateVersion(); err != nil {
		return err
	}

	if opts.ModuleConfigFileName == "" {
		return fmt.Errorf("opts.ModuleConfigFileName must not be empty: %w", commonerrors.ErrInvalidOption)
	}

	if opts.ManifestFileName == "" {
		return fmt.Errorf("opts.ManifestFileName must not be empty: %w", commonerrors.ErrInvalidOption)
	}

	return nil
}

func (opts Options) validateDirectory() error {
	if opts.Directory == "" {
		return fmt.Errorf("opts.Directory must not be empty: %w", commonerrors.ErrInvalidOption)
	}

	fileInfo, err := os.Stat(opts.Directory)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("directory %s does not exist: %w", opts.Directory, commonerrors.ErrInvalidOption)
		}
		return fmt.Errorf("failed to get directory info %s: %w: %w", opts.Directory, commonerrors.ErrInvalidOption, err)
	}

	if !fileInfo.IsDir() {
		return fmt.Errorf("%s is not a directory: %w", opts.Directory, commonerrors.ErrInvalidOption)
	}

	return nil
}

func (opts Options) validateModuleName() error {
	if err := validation.ValidateModuleName(opts.ModuleName); err != nil {
		return fmt.Errorf("opts.ModuleName: %w", err)
	}

	return nil
}

func (opts Options) validateVersion() error {
	if err := validation.ValidateModuleVersion(opts.ModuleVersion); err != nil {
		return fmt.Errorf("opts.ModuleVersion: %w", err)
	}

	return nil
}

func (opts Options) defaultCRFileNameConfigured() bool {
	return opts.DefaultCRFileName != ""
}

func (opts Options) securityConfigFileNameConfigured() bool {
	return opts.SecurityConfigFileName != ""
}
