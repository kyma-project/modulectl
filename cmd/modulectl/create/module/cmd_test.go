package module_test

import (
	"errors"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	modulecmd "github.com/kyma-project/modulectl/cmd/modulectl/create/module"
	testutils "github.com/kyma-project/modulectl/internal/common/utils/test"
	"github.com/kyma-project/modulectl/internal/service/module"
)

func Test_NewCmd_ReturnsError_WhenModuleServiceIsNil(t *testing.T) {
	_, err := modulecmd.NewCmd(nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "moduleService")
}

func Test_NewCmd_Succeeds(t *testing.T) {
	_, err := modulecmd.NewCmd(&moduleServiceStub{})

	require.NoError(t, err)
}

func Test_Execute_CallsModuleService(t *testing.T) {
	svc := &moduleServiceStub{}
	cmd, _ := modulecmd.NewCmd(svc)

	err := cmd.Execute()

	require.NoError(t, err)
	require.True(t, svc.called)
}

func Test_Execute_ReturnsError_WhenModuleServiceReturnsError(t *testing.T) {
	cmd, _ := modulecmd.NewCmd(&moduleServiceErrorStub{})

	err := cmd.Execute()

	require.ErrorIs(t, err, errSomeTestError)
}

func Test_Execute_ParsesAllModuleOptions(t *testing.T) {
	moduleConfigFile := testutils.GetRandomName(10)
	credentials := testutils.GetRandomName(10)
	gitRemote := testutils.GetRandomName(10)
	insecure := "true"
	templateOutput := testutils.GetRandomName(10)
	registryURL := testutils.GetRandomName(10)
	registryCredSelector := testutils.GetRandomName(10)
	secScannerConfig := testutils.GetRandomName(10)

	os.Args = []string{
		"module",
		"--module-config-file", moduleConfigFile,
		"--credentials", credentials,
		"--git-remote", gitRemote,
		"--insecure", insecure,
		"--output", templateOutput,
		"--registry", registryURL,
		"--registry-cred-selector", registryCredSelector,
		"--sec-scanners-config", secScannerConfig,
	}

	svc := &moduleServiceStub{}
	cmd, _ := modulecmd.NewCmd(svc)

	err := cmd.Execute()
	require.NoError(t, err)

	insecureFlagSet, err := strconv.ParseBool(insecure)
	require.NoError(t, err)

	assert.Equal(t, moduleConfigFile, svc.opts.ModuleConfigFile)
	assert.Equal(t, credentials, svc.opts.Credentials)
	assert.Equal(t, gitRemote, svc.opts.GitRemote)
	assert.Equal(t, insecureFlagSet, svc.opts.Insecure)
	assert.Equal(t, templateOutput, svc.opts.TemplateOutput)
	assert.Equal(t, registryURL, svc.opts.RegistryURL)
	assert.Equal(t, registryCredSelector, svc.opts.RegistryCredSelector)
	assert.Equal(t, secScannerConfig, svc.opts.SecScannerConfig)
}

func Test_Execute_ParsesModuleShortOptions(t *testing.T) {
	credentials := testutils.GetRandomName(10)
	templateOutput := testutils.GetRandomName(10)

	os.Args = []string{
		"module",
		"-c", credentials,
		"-o", templateOutput,
	}

	svc := &moduleServiceStub{}
	cmd, _ := modulecmd.NewCmd(svc)

	err := cmd.Execute()
	require.NoError(t, err)

	assert.Equal(t, credentials, svc.opts.Credentials)
	assert.Equal(t, templateOutput, svc.opts.TemplateOutput)
}

func Test_Execute_ModuleParsesDefaults(t *testing.T) {
	os.Args = []string{
		"module",
	}

	svc := &moduleServiceStub{}
	cmd, _ := modulecmd.NewCmd(svc)

	err := cmd.Execute()
	require.NoError(t, err)

	assert.Equal(t, modulecmd.ModuleConfigFileFlagDefault, svc.opts.ModuleConfigFile)
	assert.Equal(t, modulecmd.CredentialsFlagDefault, svc.opts.Credentials)
	assert.Equal(t, modulecmd.GitRemoteFlagDefault, svc.opts.GitRemote)
	assert.Equal(t, modulecmd.InsecureFlagDefault, svc.opts.Insecure)
	assert.Equal(t, modulecmd.TemplateOutputFlagDefault, svc.opts.TemplateOutput)
	assert.Equal(t, modulecmd.RegistryURLFlagDefault, svc.opts.RegistryURL)
	assert.Equal(t, modulecmd.RegistryCredSelectorFlagDefault, svc.opts.RegistryCredSelector)
	assert.Equal(t, modulecmd.SecScannersConfigFlagDefault, svc.opts.SecScannerConfig)
}

// Test Stubs

type moduleServiceStub struct {
	called bool
	opts   module.Options
}

func (m *moduleServiceStub) CreateModule(opts module.Options) error {
	m.called = true
	m.opts = opts
	return nil
}

type moduleServiceErrorStub struct{}

var errSomeTestError = errors.New("some test error")

func (s *moduleServiceErrorStub) CreateModule(_ module.Options) error {
	return errSomeTestError
}
