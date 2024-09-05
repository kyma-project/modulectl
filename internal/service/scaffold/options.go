package scaffold

import (
	"errors"
	"fmt"
	"os"

	commonerrors "github.com/kyma-project/modulectl/internal/common/errors"
	"github.com/kyma-project/modulectl/internal/common/validations"
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
	ModuleChannel             string
}

func (opts Options) validate() error {
	if opts.Out == nil {
		return fmt.Errorf("%w: opts.Out must not be nil", commonerrors.ErrInvalidOption)
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

	if err := opts.validateChannel(); err != nil {
		return err
	}

	if opts.ModuleConfigFileName == "" {
		return fmt.Errorf("%w: opts.ModuleConfigFileName must not be empty", commonerrors.ErrInvalidOption)
	}

	if opts.ManifestFileName == "" {
		return fmt.Errorf("%w: opts.ManifestFileName must not be empty", commonerrors.ErrInvalidOption)
	}

	return nil
}

func (opts Options) validateDirectory() error {
	if opts.Directory == "" {
		return fmt.Errorf("%w: opts.Directory must not be empty", commonerrors.ErrInvalidOption)
	}

	fileInfo, err := os.Stat(opts.Directory)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("%w: directory %s does not exist", commonerrors.ErrInvalidOption, opts.Directory)
		}
		return fmt.Errorf("%w: failed to get directory info %s: %w", commonerrors.ErrInvalidOption, opts.Directory, err)
	}

	if !fileInfo.IsDir() {
		return fmt.Errorf("%w: %s is not a directory", commonerrors.ErrInvalidOption, opts.Directory)
	}

	return nil
}

func (opts Options) validateModuleName() error {
	if err := validations.ValidateModuleName(opts.ModuleName); err != nil {
		return fmt.Errorf("%w: %w", commonerrors.ErrInvalidOption, err)
	}

	return nil
}

func (opts Options) validateVersion() error {
	if err := validations.ValidateModuleVersion(opts.ModuleVersion); err != nil {
		return fmt.Errorf("%w: %w", commonerrors.ErrInvalidOption, err)
	}

	return nil
}

func (opts Options) validateChannel() error {
	if err := validations.ValidateModuleChannel(opts.ModuleChannel); err != nil {
		return fmt.Errorf("%w: %w", commonerrors.ErrInvalidOption, err)
	}

	return nil
}

func (opts Options) defaultCRFileNameConfigured() bool {
	return opts.DefaultCRFileName != ""
}

func (opts Options) securityConfigFileNameConfigured() bool {
	return opts.SecurityConfigFileName != ""
}
