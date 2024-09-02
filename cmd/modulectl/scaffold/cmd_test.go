package scaffold_test

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	scaffoldcmd "github.com/kyma-project/modulectl/cmd/modulectl/scaffold"
	"github.com/kyma-project/modulectl/internal/service/scaffold"
	"github.com/kyma-project/modulectl/internal/testutils"
)

func Test_NewCmd_ReturnsError_WhenScaffoldServiceIsNil(t *testing.T) {
	_, err := scaffoldcmd.NewCmd(nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "scaffoldService")
}

func Test_NewCmd_Succceeds(t *testing.T) {
	_, err := scaffoldcmd.NewCmd(&scaffoldServiceStub{})

	require.NoError(t, err)
}

func Test_Execute_CallsScaffoldService(t *testing.T) {
	svc := &scaffoldServiceStub{}
	cmd, _ := scaffoldcmd.NewCmd(svc)

	err := cmd.Execute()

	require.NoError(t, err)
	require.True(t, svc.called)
}

func Test_Execute_ReturnsError_WhenServiceReturnsError(t *testing.T) {
	cmd, _ := scaffoldcmd.NewCmd(&scaffoldServiceErrorStub{})

	err := cmd.Execute()

	require.ErrorIs(t, err, errSomeTestError)
}

func Test_Execute_ParsesOptions(t *testing.T) {
	directory := testutils.RandomName(10)
	moduleConfigFile := testutils.RandomName(10)
	manifestFile := testutils.RandomName(10)
	defaultCRFile := testutils.RandomName(10)
	securityConfigFile := testutils.RandomName(10)
	moduleName := testutils.RandomName(10)
	moduleVersion := "1.1.1"
	moduleChannel := testutils.RandomName(10)
	os.Args = []string{
		"scaffold",
		"--directory", directory,
		"--module-config", moduleConfigFile,
		"--overwrite",
		"--gen-manifest", manifestFile,
		"--gen-default-cr=" + defaultCRFile,
		"--gen-security-config=" + securityConfigFile,
		"--module-name", moduleName,
		"--module-version", moduleVersion,
		"--module-channel", moduleChannel,
	}
	svc := &scaffoldServiceStub{}
	cmd, _ := scaffoldcmd.NewCmd(svc)

	err := cmd.Execute()
	require.NoError(t, err)

	assert.Equal(t, moduleName, svc.opts.ModuleName)
	assert.Equal(t, directory, svc.opts.Directory)
	assert.Equal(t, moduleConfigFile, svc.opts.ModuleConfigFileName)
	assert.True(t, svc.opts.ModuleConfigFileOverwrite)
	assert.Equal(t, manifestFile, svc.opts.ManifestFileName)
	assert.Equal(t, defaultCRFile, svc.opts.DefaultCRFileName)
	assert.Equal(t, securityConfigFile, svc.opts.SecurityConfigFileName)
	assert.Equal(t, moduleVersion, svc.opts.ModuleVersion)
	assert.Equal(t, moduleChannel, svc.opts.ModuleChannel)
}

func Test_Execute_ParsesShortOptions(t *testing.T) {
	directory := testutils.RandomName(10)
	os.Args = []string{
		"scaffold",
		"-d", directory,
		"-o",
	}
	svc := &scaffoldServiceStub{}
	cmd, _ := scaffoldcmd.NewCmd(svc)

	err := cmd.Execute()
	require.NoError(t, err)

	assert.Equal(t, directory, svc.opts.Directory)
	assert.True(t, svc.opts.ModuleConfigFileOverwrite)
}

func Test_Execute_ParsesDefaults(t *testing.T) {
	os.Args = []string{
		"scaffold",
	}
	svc := &scaffoldServiceStub{}
	cmd, _ := scaffoldcmd.NewCmd(svc)

	err := cmd.Execute()
	require.NoError(t, err)

	assert.Equal(t, scaffoldcmd.ModuleNameFlagDefault, svc.opts.ModuleName)
	assert.Equal(t, scaffoldcmd.DirectoryFlagDefault, svc.opts.Directory)
	assert.Equal(t, scaffoldcmd.ModuleConfigFileFlagDefault, svc.opts.ModuleConfigFileName)
	assert.Equal(t, scaffoldcmd.ModuleConfigFileOverwriteFlagDefault, svc.opts.ModuleConfigFileOverwrite)
	assert.Equal(t, scaffoldcmd.ManifestFileFlagDefault, svc.opts.ManifestFileName)
	assert.Equal(t, scaffoldcmd.DefaultCRFlagDefault, svc.opts.DefaultCRFileName)
	assert.Equal(t, scaffoldcmd.SecurityConfigFileFlagDefault, svc.opts.SecurityConfigFileName)
	assert.Equal(t, scaffoldcmd.ModuleVersionFlagDefault, svc.opts.ModuleVersion)
	assert.Equal(t, scaffoldcmd.ModuleChannelFlagDefault, svc.opts.ModuleChannel)
}

func Test_Execute_ParsesNoOptDefaults(t *testing.T) {
	os.Args = []string{
		"scaffold",
		"--gen-default-cr",
		"--gen-security-config",
	}
	svc := &scaffoldServiceStub{}
	cmd, _ := scaffoldcmd.NewCmd(svc)

	err := cmd.Execute()
	require.NoError(t, err)

	assert.Equal(t, scaffoldcmd.DefaultCRFlagNoOptDefault, svc.opts.DefaultCRFileName)
	assert.Equal(t, scaffoldcmd.SecurityConfigFileFlagNoOptDefault, svc.opts.SecurityConfigFileName)
}

// Test Stubs

type scaffoldServiceStub struct {
	called bool
	opts   scaffold.Options
}

func (s *scaffoldServiceStub) CreateScaffold(opts scaffold.Options) error {
	s.called = true
	s.opts = opts
	return nil
}

type scaffoldServiceErrorStub struct{}

var errSomeTestError = errors.New("some test error")

func (s *scaffoldServiceErrorStub) CreateScaffold(_ scaffold.Options) error {
	return errSomeTestError
}
