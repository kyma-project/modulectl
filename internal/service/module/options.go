package module

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/kyma-project/modulectl/tools/io"
)

// Options defines available options for the create module command
type Options struct {
	Out                  io.Out // TODO: Can be extracted one level above
	ModuleConfigFile     string
	Credentials          string
	GitRemote            string
	Insecure             bool
	TemplateOutput       string
	RegistryURL          string
	RegistryCredSelector string
	SecScannerConfig     string
}

func (opts Options) validate() error {
	if opts.Out == nil {
		return fmt.Errorf("%w: opts.Out must not be nil", ErrInvalidOption)
	}

	if opts.ModuleConfigFile == "" {
		return fmt.Errorf("%w:  opts.ModuleConfigFile must not be empty", ErrInvalidOption)
	}

	matched, err := regexp.MatchString("/(.+):(.+)/g", opts.Credentials)
	if err != nil {
		return fmt.Errorf("%w: opts.Credentials could not be parsed: %w", ErrInvalidOption, err)
	} else if !matched {
		return fmt.Errorf("%w: opts.Credentials is in invalid format", ErrInvalidOption)
	}

	if opts.GitRemote == "" {
		return fmt.Errorf("%w:  opts.GitRemote must not be empty", ErrInvalidOption)
	}

	if opts.TemplateOutput == "" {
		return fmt.Errorf("%w:  opts.TemplateOutput must not be empty", ErrInvalidOption)
	}

	if !strings.HasPrefix(opts.RegistryURL, "http") {
		return fmt.Errorf("%w:  opts.RegistryURL does not start with http(s)", ErrInvalidOption)
	}

	if opts.SecScannerConfig == "" {
		return fmt.Errorf("%w:  opts.SecurityScanConfig must not be empty", ErrInvalidOption)
	}

	return nil
}
