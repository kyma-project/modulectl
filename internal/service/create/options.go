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
	ModuleConfigFile     string
	Credentials          string
	GitRemote            string
	Insecure             bool
	TemplateOutput       string
	RegistryURL          string
	RegistryCredSelector string
}

func (opts Options) Validate() error {
	if opts.Out == nil {
		return fmt.Errorf("%w: opts.Out must not be nil", commonerrors.ErrInvalidOption)
	}

	if opts.ModuleConfigFile == "" {
		return fmt.Errorf("%w:  opts.ModuleConfigFile must not be empty", commonerrors.ErrInvalidOption)
	}

	if opts.Credentials != "" {
		matched, err := regexp.MatchString("(.+):(.+)", opts.Credentials)
		if err != nil {
			return fmt.Errorf("%w: opts.Credentials could not be parsed: %w", commonerrors.ErrInvalidOption, err)
		} else if !matched {
			return fmt.Errorf("%w: opts.Credentials is in invalid format", commonerrors.ErrInvalidOption)
		}
	}

	if opts.TemplateOutput == "" {
		return fmt.Errorf("%w:  opts.TemplateOutput must not be empty", commonerrors.ErrInvalidOption)
	}

	if opts.RegistryURL != "" && !strings.HasPrefix(opts.RegistryURL, "http") {
		return fmt.Errorf("%w:  opts.RegistryURL does not start with http(s)", commonerrors.ErrInvalidOption)
	}

	return nil
}
