package scaffold

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/Masterminds/semver/v3"

	commonerrors "github.com/kyma-project/modulectl/internal/common/errors"
	iotools "github.com/kyma-project/modulectl/tools/io"
)

const (
	// // taken from "https://github.com/open-component-model/ocm/blob/4473dacca406e4c84c0ac5e6e14393c659384afc/resources/component-descriptor-v2-schema.yaml#L40"
	moduleNamePattern   = "^[a-z][-a-z0-9]*([.][a-z][-a-z0-9]*)*[.][a-z]{2,}(/[a-z][-a-z0-9_]*([.][a-z][-a-z0-9_]*)*)+$"
	moduleNameMaxLength = 255
	channelMinLength    = 3
	channelMaxLength    = 32
	channelPattern      = "^[a-z]+$"
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
	if opts.ModuleName == "" {
		return fmt.Errorf("%w: opts.ModuleName must not be empty", commonerrors.ErrInvalidOption)
	}

	if len(opts.ModuleName) > moduleNameMaxLength {
		return fmt.Errorf("%w: opts.ModuleName length must not exceed %q characters", commonerrors.ErrInvalidOption, moduleNameMaxLength)
	}

	if matched, err := regexp.MatchString(moduleNamePattern, opts.ModuleName); err != nil {
		return fmt.Errorf("%w: failed to evaluate regex pattern for opts.ModuleName", commonerrors.ErrInvalidOption)
	} else if !matched {
		return fmt.Errorf("%w: opts.ModuleName must match the required pattern, e.g: 'github.com/path-to/your-repo'", commonerrors.ErrInvalidOption)
	}

	return nil
}

func (opts Options) validateVersion() error {
	if opts.ModuleVersion == "" {
		return fmt.Errorf("%w: opts.ModuleVersion must not be empty", commonerrors.ErrInvalidOption)
	}

	if err := opts.validateSemanticVersion(); err != nil {
		return err
	}

	return nil
}

func (opts Options) validateSemanticVersion() error {
	val := strings.TrimSpace(opts.ModuleVersion)

	// strip the leading "v", if any, because it's not part of a proper semver
	val = strings.TrimPrefix(val, "v")

	_, err := semver.StrictNewVersion(val)
	if err != nil {
		return fmt.Errorf("%w: opts.ModuleVersion failed to parse as semantic version: %w", commonerrors.ErrInvalidOption, err)
	}

	return nil
}

func (opts Options) validateChannel() error {
	if opts.ModuleChannel == "" {
		return fmt.Errorf("%w: opts.ModuleChannel must not be empty", commonerrors.ErrInvalidOption)
	}

	if len(opts.ModuleChannel) > channelMaxLength {
		return fmt.Errorf("%w: opts.ModuleChannel length must not exceed %q characters", commonerrors.ErrInvalidOption, channelMaxLength)
	}

	if len(opts.ModuleChannel) < channelMinLength {
		return fmt.Errorf("%w: opts.ModuleChannel length must be at least %q characters", commonerrors.ErrInvalidOption, channelMinLength)
	}

	if matched, err := regexp.MatchString(channelPattern, opts.ModuleChannel); err != nil {
		return fmt.Errorf("%w: failed to evaluate regex pattern for opts.ModuleChannel", commonerrors.ErrInvalidOption)
	} else if !matched {
		return fmt.Errorf("%w: opts.ModuleChannel must match the required pattern, only characters from a-z are allowed", commonerrors.ErrInvalidOption)
	}

	return nil
}

func (opts Options) defaultCRFileNameConfigured() bool {
	return opts.DefaultCRFileName != ""
}

func (opts Options) securityConfigFileNameConfigured() bool {
	return opts.SecurityConfigFileName != ""
}
