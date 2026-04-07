package create_test

import (
	"io"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kyma-project/modulectl/internal/service/create"
	iotools "github.com/kyma-project/modulectl/tools/io"
)

func Test_Validate_Options(t *testing.T) {
	tests := []struct {
		name    string
		options create.Options
		wantErr bool
		errMsg  string
	}{
		{
			name: "Out is nil",
			options: create.Options{
				Out: nil,
			},
			wantErr: true,
			errMsg:  "opts.Out must not be nil",
		},
		{
			name: "ConfigFile is empty",
			options: create.Options{
				Out:        iotools.NewDefaultOut(io.Discard),
				ConfigFile: "",
			},
			wantErr: true,
			errMsg:  "opts.ConfigFile must not be empty",
		},
		{
			name: "TemplateOutput is empty",
			options: create.Options{
				Out:            iotools.NewDefaultOut(io.Discard),
				ConfigFile:     "config.yaml",
				TemplateOutput: "",
			},
			wantErr: true,
			errMsg:  "opts.TemplateOutput must not be empty",
		},
		{
			name: "OutputConstructorFile is empty",
			options: create.Options{
				Out:                   iotools.NewDefaultOut(io.Discard),
				ConfigFile:            "config.yaml",
				TemplateOutput:        "output",
				OutputConstructorFile: "",
			},
			wantErr: true,
			errMsg:  "opts.OutputConstructorFile must not be empty",
		},
		{
			name: "ModuleSourcesGitDirectory is empty",
			options: create.Options{
				Out:                       iotools.NewDefaultOut(io.Discard),
				ConfigFile:                "config.yaml",
				TemplateOutput:            "output",
				OutputConstructorFile:     "constructor.yaml",
				ModuleSourcesGitDirectory: "",
			},
			wantErr: true,
			errMsg:  "opts.ModuleSourcesGitDirectory must not be empty",
		},
		{
			name: "ModuleSourcesGitDirectory is not a git directory",
			options: create.Options{
				Out:                       iotools.NewDefaultOut(io.Discard),
				ConfigFile:                "config.yaml",
				TemplateOutput:            "output",
				OutputConstructorFile:     "constructor.yaml",
				ModuleSourcesGitDirectory: ".",
			},
			wantErr: true,
			errMsg:  "currently configured module-sources-git-directory \".\" must point to a valid git repository:",
		},
		{
			name: "All fields valid",
			options: create.Options{
				Out:                       iotools.NewDefaultOut(io.Discard),
				ConfigFile:                "config.yaml",
				TemplateOutput:            "output",
				OutputConstructorFile:     "constructor.yaml",
				ModuleSourcesGitDirectory: "../../../",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.options.Validate()
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
