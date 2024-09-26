package create_test

import (
	"errors"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	createcmd "github.com/kyma-project/modulectl/cmd/modulectl/create"
	"github.com/kyma-project/modulectl/internal/service/create"
	"github.com/kyma-project/modulectl/internal/testutils"
)

func Test_NewCmd_ReturnsError_WhenModuleServiceIsNil(t *testing.T) {
	_, err := createcmd.NewCmd(nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "service must not be nil")
}

func Test_NewCmd_Succeeds(t *testing.T) {
	_, err := createcmd.NewCmd(&moduleServiceStub{})

	require.NoError(t, err)
}

func Test_Execute_CallsModuleService(t *testing.T) {
	svc := &moduleServiceStub{}
	cmd, _ := createcmd.NewCmd(svc)

	err := cmd.Execute()

	require.NoError(t, err)
	require.True(t, svc.called)
}

func Test_Execute_ReturnsError_WhenModuleServiceReturnsError(t *testing.T) {
	cmd, _ := createcmd.NewCmd(&moduleServiceErrorStub{})

	err := cmd.Execute()

	require.ErrorIs(t, err, errSomeTestError)
}

func Test_Execute_ParsesAllModuleOptions(t *testing.T) {
	moduleConfigFile := testutils.RandomName(10)
	credentials := testutils.RandomName(10)
	gitRemote := testutils.RandomName(10)
	insecure := "true"
	templateOutput := testutils.RandomName(10)
	registryURL := testutils.RandomName(10)
	credSelector := testutils.RandomName(10)

	os.Args = []string{
		"create",
		"--module-config-file", moduleConfigFile,
		"--git-remote", gitRemote,
		"--insecure", insecure,
		"--output", templateOutput,
		"--registry", registryURL,
		"--registry-credentials", credentials,
		"--registry-cred-selector", credSelector,
	}

	svc := &moduleServiceStub{}
	cmd, _ := createcmd.NewCmd(svc)

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
	assert.Equal(t, credSelector, svc.opts.RegistryCredSelector)
}

func Test_Execute_ParsesModuleShortOptions(t *testing.T) {
	credentials := testutils.RandomName(10)
	templateOutput := testutils.RandomName(10)

	os.Args = []string{
		"create",
		"-c", credentials,
		"-o", templateOutput,
	}

	svc := &moduleServiceStub{}
	cmd, _ := createcmd.NewCmd(svc)

	err := cmd.Execute()
	require.NoError(t, err)

	assert.Equal(t, credentials, svc.opts.Credentials)
	assert.Equal(t, templateOutput, svc.opts.TemplateOutput)
}

func Test_Execute_ModuleParsesDefaults(t *testing.T) {
	os.Args = []string{
		"create",
	}

	svc := &moduleServiceStub{}
	cmd, _ := createcmd.NewCmd(svc)

	err := cmd.Execute()
	require.NoError(t, err)

	assert.Equal(t, createcmd.ModuleConfigFileFlagDefault, svc.opts.ModuleConfigFile)
	assert.Equal(t, createcmd.CredentialsFlagDefault, svc.opts.Credentials)
	assert.Equal(t, createcmd.GitRemoteFlagDefault, svc.opts.GitRemote)
	assert.Equal(t, createcmd.InsecureFlagDefault, svc.opts.Insecure)
	assert.Equal(t, createcmd.TemplateOutputFlagDefault, svc.opts.TemplateOutput)
	assert.Equal(t, createcmd.RegistryURLFlagDefault, svc.opts.RegistryURL)
	assert.Equal(t, createcmd.RegistryCredSelectorFlagDefault, svc.opts.RegistryCredSelector)
}

// Test Stubs

type moduleServiceStub struct {
	called bool
	opts   create.Options
}

func (m *moduleServiceStub) Run(opts create.Options) error {
	m.called = true
	m.opts = opts
	return nil
}

type moduleServiceErrorStub struct{}

var errSomeTestError = errors.New("some test error")

func (s *moduleServiceErrorStub) Run(_ create.Options) error {
	return errSomeTestError
}
