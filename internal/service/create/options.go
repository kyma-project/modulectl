package create

import (
	"fmt"
	"os"
	"path/filepath"

	commonerrors "github.com/kyma-project/modulectl/internal/common/errors"
	iotools "github.com/kyma-project/modulectl/tools/io"
)

type Options struct {
	Out                       iotools.Out
	ConfigFile                string
	TemplateOutput            string
	ModuleSourcesGitDirectory string
	SkipVersionValidation     bool
	OutputConstructorFile     string
}

func (opts Options) Validate() error {
	if opts.Out == nil {
		return fmt.Errorf("opts.Out must not be nil: %w", commonerrors.ErrInvalidOption)
	}

	if opts.ConfigFile == "" {
		return fmt.Errorf("opts.ConfigFile must not be empty: %w", commonerrors.ErrInvalidOption)
	}

	if opts.TemplateOutput == "" {
		return fmt.Errorf("opts.TemplateOutput must not be empty: %w", commonerrors.ErrInvalidOption)
	}

	if opts.OutputConstructorFile == "" {
		return fmt.Errorf("opts.OutputConstructorFile must not be empty: %w", commonerrors.ErrInvalidOption)
	}

	if opts.ModuleSourcesGitDirectory == "" {
		return fmt.Errorf("opts.ModuleSourcesGitDirectory must not be empty: %w", commonerrors.ErrInvalidOption)
	} else {
		if isGitDir := isGitDirectory(opts.ModuleSourcesGitDirectory); !isGitDir {
			return fmt.Errorf("currently configured module-sources-git-directory \"%s\" must point to "+
				"a valid git repository: %w",
				opts.ModuleSourcesGitDirectory, commonerrors.ErrInvalidOption)
		}
	}

	return nil
}

func isGitDirectory(path string) bool {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	gitPath := filepath.Join(absPath, ".git")
	info, err := os.Stat(gitPath)
	if err != nil {
		return false
	}
	return info.IsDir()
}
