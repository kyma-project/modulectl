package scaffold_test

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kyma-project/modulectl/internal/service/scaffold"
	iotools "github.com/kyma-project/modulectl/tools/io"
)

func Test_Validate_Options(t *testing.T) {
	tests := []struct {
		name    string
		options scaffold.Options
		wantErr bool
		errMsg  string
	}{
		{
			name:    "Out is nil",
			options: scaffold.Options{Out: nil},
			wantErr: true,
			errMsg:  "opts.Out must not be nil",
		},
		{
			name: "ModuleName is empty",
			options: scaffold.Options{
				Out:        iotools.NewDefaultOut(io.Discard),
				ModuleName: "",
			},
			wantErr: true,
			errMsg:  "opts.ModuleName must not be empty",
		},
		{
			name: "ModuleName exceeds length",
			options: scaffold.Options{
				Out:        iotools.NewDefaultOut(io.Discard),
				ModuleName: strings.Repeat("a", 256),
			},
			wantErr: true,
			errMsg:  "opts.ModuleName length must not exceed",
		},
		{
			name: "ModuleName invalid pattern",
			options: scaffold.Options{
				Out:        iotools.NewDefaultOut(io.Discard),
				ModuleName: "invalid_name",
			},
			wantErr: true,
			errMsg:  "opts.ModuleName must match the required pattern",
		},
		{
			name: "Directory is empty",
			options: scaffold.Options{
				Out:        iotools.NewDefaultOut(io.Discard),
				ModuleName: "github.com/kyma-project/test",
				Directory:  "",
			},
			wantErr: true,
			errMsg:  "opts.Directory must not be empty",
		},
		{
			name: "ModuleVersion is empty",
			options: scaffold.Options{
				Out:           iotools.NewDefaultOut(io.Discard),
				ModuleName:    "github.com/kyma-project/test",
				Directory:     "./",
				ModuleVersion: "",
			},
			wantErr: true,
			errMsg:  "opts.ModuleVersion must not be empty",
		},
		{
			name: "ModuleVersion invalid",
			options: scaffold.Options{
				Out:           iotools.NewDefaultOut(io.Discard),
				ModuleName:    "github.com/kyma-project/test",
				Directory:     "./",
				ModuleVersion: "invalid",
			},
			wantErr: true,
			errMsg:  "opts.ModuleVersion failed to parse as semantic version",
		},
		{
			name: "ModuleChannel is empty",
			options: scaffold.Options{
				Out:           iotools.NewDefaultOut(io.Discard),
				ModuleName:    "github.com/kyma-project/test",
				Directory:     "./",
				ModuleVersion: "0.0.1",
				ModuleChannel: "",
			},
			wantErr: true,
			errMsg:  "opts.ModuleChannel must not be empty",
		},
		{
			name: "ModuleChannel exceeds length",
			options: scaffold.Options{
				Out:           iotools.NewDefaultOut(io.Discard),
				ModuleName:    "github.com/kyma-project/test",
				Directory:     "./",
				ModuleVersion: "0.0.1",
				ModuleChannel: strings.Repeat("a", 33),
			},
			wantErr: true,
			errMsg:  "opts.ModuleChannel length must not exceed",
		},
		{
			name: "ModuleChannel below minimum length",
			options: scaffold.Options{
				Out:           iotools.NewDefaultOut(io.Discard),
				ModuleName:    "github.com/kyma-project/test",
				Directory:     "./",
				ModuleVersion: "0.0.1",
				ModuleChannel: "aa",
			},
			wantErr: true,
			errMsg:  "opts.ModuleChannel length must be at least",
		},
		{
			name: "ModuleChannel invalid pattern",
			options: scaffold.Options{
				Out:           iotools.NewDefaultOut(io.Discard),
				ModuleName:    "github.com/kyma-project/test",
				Directory:     "./",
				ModuleVersion: "0.0.1",
				ModuleChannel: "invalid_channel",
			},
			wantErr: true,
			errMsg:  "opts.ModuleChannel must match the required pattern",
		},
		{
			name: "ModuleConfigFileName is empty",
			options: scaffold.Options{
				Out:                  iotools.NewDefaultOut(io.Discard),
				ModuleName:           "github.com/kyma-project/test",
				Directory:            "./",
				ModuleVersion:        "0.0.1",
				ModuleChannel:        "stable",
				ModuleConfigFileName: "",
			},
			wantErr: true,
			errMsg:  "opts.ModuleConfigFileName must not be empty",
		},
		{
			name: "ManifestFileName is empty",
			options: scaffold.Options{
				Out:                  iotools.NewDefaultOut(io.Discard),
				ModuleName:           "github.com/kyma-project/test",
				Directory:            "./",
				ModuleVersion:        "0.0.1",
				ModuleChannel:        "stable",
				ModuleConfigFileName: "config.yaml",
				ManifestFileName:     "",
			},
			wantErr: true,
			errMsg:  "opts.ManifestFileName must not be empty",
		},
		{
			name: "All fields valid",
			options: scaffold.Options{
				Out:                    iotools.NewDefaultOut(io.Discard),
				ModuleName:             "github.com/kyma-project/test",
				Directory:              "./",
				ModuleVersion:          "0.0.1",
				ModuleChannel:          "stable",
				ModuleConfigFileName:   "config.yaml",
				ManifestFileName:       "manifest.yaml",
				SecurityConfigFileName: "",
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
