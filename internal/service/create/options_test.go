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
			name: "Credentials invalid format",
			options: create.Options{
				Out:         iotools.NewDefaultOut(io.Discard),
				ConfigFile:  "config.yaml",
				Credentials: "missingsemicolon",
			},
			wantErr: true,
			errMsg:  "opts.Credentials is in invalid format",
		},
		{
			name: "TemplateOutput is empty",
			options: create.Options{
				Out:            iotools.NewDefaultOut(io.Discard),
				ConfigFile:     "config.yaml",
				Credentials:    "username:password",
				TemplateOutput: "",
			},
			wantErr: true,
			errMsg:  "opts.TemplateOutput must not be empty",
		},
		{
			name: "All fields valid",
			options: create.Options{
				Out:            iotools.NewDefaultOut(io.Discard),
				ConfigFile:     "config.yaml",
				Credentials:    "username:password",
				TemplateOutput: "output",
				RegistryURL:    "http://registry.example.com",
			},
			wantErr: false,
		},
		{
			name: "RegistryURL does not start with http",
			options: create.Options{
				Out:            iotools.NewDefaultOut(io.Discard),
				ConfigFile:     "config.yaml",
				Credentials:    "username:password",
				TemplateOutput: "output",
				RegistryURL:    "ftp://registry.example.com",
			},
			wantErr: true,
			errMsg:  "opts.RegistryURL does not start with http(s)",
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
