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
			name:    "Out is nil",
			options: create.Options{Out: nil},
			wantErr: true,
			errMsg:  "opts.Out must not be nil",
		},
		{
			name: "ModuleConfigFile is empty",
			options: create.Options{
				Out:              iotools.NewDefaultOut(io.Discard),
				ModuleConfigFile: "",
			},
			wantErr: true,
			errMsg:  "opts.ModuleConfigFile must not be empty",
		},
		{
			name: "Credentials invalid format",
			options: create.Options{
				Out:              iotools.NewDefaultOut(io.Discard),
				ModuleConfigFile: "config.yaml",
				Credentials:      "missingsemicolon",
			},
			wantErr: true,
			errMsg:  "opts.Credentials is in invalid format",
		},
		{
			name: "GitRemote is empty",
			options: create.Options{
				Out:              iotools.NewDefaultOut(io.Discard),
				ModuleConfigFile: "config.yaml",
				Credentials:      "username:password",
				GitRemote:        "",
			},
			wantErr: true,
			errMsg:  "opts.GitRemote must not be empty",
		},
		{
			name: "TemplateOutput is empty",
			options: create.Options{
				Out:              iotools.NewDefaultOut(io.Discard),
				ModuleConfigFile: "config.yaml",
				Credentials:      "username:password",
				GitRemote:        "origin",
				TemplateOutput:   "",
			},
			wantErr: true,
			errMsg:  "opts.TemplateOutput must not be empty",
		},
		{
			name: "RegistryURL does not start with http",
			options: create.Options{
				Out:              iotools.NewDefaultOut(io.Discard),
				ModuleConfigFile: "config.yaml",
				Credentials:      "username:password",
				GitRemote:        "origin",
				TemplateOutput:   "output",
				RegistryURL:      "ftp://registry.example.com",
			},
			wantErr: true,
			errMsg:  "opts.RegistryURL does not start with http(s)",
		},
		{
			name: "All fields valid",
			options: create.Options{
				Out:              iotools.NewDefaultOut(io.Discard),
				ModuleConfigFile: "config.yaml",
				Credentials:      "username:password",
				GitRemote:        "origin",
				TemplateOutput:   "output",
				RegistryURL:      "http://registry.example.com",
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
