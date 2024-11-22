package create

import (
	"fmt"
	"regexp"
	"strings"

	commonerrors "github.com/kyma-project/modulectl/internal/common/errors"
	iotools "github.com/kyma-project/modulectl/tools/io"
)

type Options struct {
	Out                  iotools.Out
	ConfigFile           string
	Credentials          string
	Insecure             bool
	TemplateOutput       string
	RegistryURL          string
	RegistryCredSelector string
}

func (opts Options) Validate() error {
	if opts.Out == nil {
		return fmt.Errorf("opts.Out must not be nil: %w", commonerrors.ErrInvalidOption)
	}

	if opts.ConfigFile == "" {
		return fmt.Errorf("opts.ConfigFile must not be empty: %w", commonerrors.ErrInvalidOption)
	}

	if opts.Credentials != "" {
		matched, err := regexp.MatchString("(.+):(.+)", opts.Credentials)
		if err != nil {
			return fmt.Errorf("opts.Credentials could not be parsed: %w", commonerrors.ErrInvalidOption, err)
		} else if !matched {
			return fmt.Errorf("opts.Credentials is in invalid format: %w", commonerrors.ErrInvalidOption)
		}
	}

	if opts.TemplateOutput == "" {
		return fmt.Errorf("opts.TemplateOutput must not be empty: %w", commonerrors.ErrInvalidOption)
	}

	if opts.RegistryURL != "" && !strings.HasPrefix(opts.RegistryURL, "http") {
		return fmt.Errorf("opts.RegistryURL does not start with http(s): %w", commonerrors.ErrInvalidOption)
	}

	return nil
}
