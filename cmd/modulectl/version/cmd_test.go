package version_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kyma-project/modulectl/cmd/modulectl/version"
)

func TestNewCmd_WhenCalled_ReturnsNoErr(t *testing.T) {
	_, err := version.NewCmd()

	require.NoError(t, err)
}

func TestNewCmd_WhenCalled_CmdContainsUseDescription(t *testing.T) {
	cmd, _ := version.NewCmd()

	assert.Equal(t, "version", cmd.Use)
}

func TestNewCmd_WhenCalled_CmdContainsShortDescription(t *testing.T) {
	cmd, _ := version.NewCmd()

	assert.Equal(t, "Prints the current modulectl version.", cmd.Short)
}

func TestNewCmd_WhenCalled_CmdContainsLongDescription(t *testing.T) {
	cmd, _ := version.NewCmd()

	assert.Equal(
		t,
		"This command prints the current semantic version of the modulectl binary set at build time.",
		cmd.Long,
	)
}

func TestNewCmd_WhenCalled_CmdRunNotNil(t *testing.T) {
	cmd, _ := version.NewCmd()

	require.NotNil(t, cmd.Run)
}

func TestNewCmd_WhenCalled_CmdHasAlias(t *testing.T) {
	cmd, _ := version.NewCmd()

	require.Len(t, cmd.Aliases, 1)
	require.Contains(t, cmd.Aliases, "v")
}

func TestNewCmd_WhenCalled_CmdExecuteCanBeCalledWithNoVersionGlobalSet(t *testing.T) {
	cmd, _ := version.NewCmd()
	os.Args = []string{"version"}

	err := cmd.Execute()

	require.NoError(t, err)
}
