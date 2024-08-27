package module_test

import (
	"errors"
	"os"
	"testing"

	modulecmd "github.com/kyma-project/modulectl/cmd/modulectl/create/module"
	modulesvc "github.com/kyma-project/modulectl/internal/service/module"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	moduleConfigFile := "some random file path"

	os.Args = []string{
		"module",
		"--module-config-file", moduleConfigFile,
	}

	svc := &moduleServiceStub{}
	cmd, _ := modulecmd.NewCmd(svc)

	cmd.Execute()

	assert.Equal(t, moduleConfigFile, svc.opts.ModuleConfigFile)
}

// ***************
// Test Stubs
// ***************

type moduleServiceStub struct {
	called bool
	opts   modulesvc.Options
}

func (m *moduleServiceStub) CreateModule(opts modulesvc.Options) error {
	m.called = true
	m.opts = opts
	return nil
}

type moduleServiceErrorStub struct{}

var errSomeTestError = errors.New("some test error")

func (s *moduleServiceErrorStub) CreateModule(_ modulesvc.Options) error {
	return errSomeTestError
}
