package moduleconfig_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	commonerrors "github.com/kyma-project/modulectl/internal/common/errors"
	"github.com/kyma-project/modulectl/internal/common/types"
	"github.com/kyma-project/modulectl/internal/service/moduleconfig"
	iotools "github.com/kyma-project/modulectl/tools/io"
)

const (
	directory        = "./bin/dir"
	moduleConfigFile = "config.yaml"
)

func Test_NewService_ReturnsError_WhenFileSystemIsNil(t *testing.T) {
	_, err := moduleconfig.NewService(
		nil,
		&fileGeneratorErrorStub{},
	)

	require.ErrorIs(t, err, commonerrors.ErrInvalidArg)
	assert.Contains(t, err.Error(), "fileSystem must not be nil")
}

func Test_NewService_ReturnsError_WhenFileGeneratorIsNil(t *testing.T) {
	_, err := moduleconfig.NewService(
		&errorStub{},
		nil,
	)

	require.ErrorIs(t, err, commonerrors.ErrInvalidArg)
	assert.Contains(t, err.Error(), "fileGenerator must not be nil")
}

func Test_ForceExplicitOverwrite_ReturnsError_WhenFilesystemReturnsError(t *testing.T) {
	svc, _ := moduleconfig.NewService(
		&errorStub{},
		&fileGeneratorErrorStub{},
	)

	result := svc.ForceExplicitOverwrite(directory, moduleConfigFile, true)

	require.ErrorIs(t, result, errSomeOSError)
}

func Test_ForceExplicitOverwrite_ReturnsError_WhenFileExistsAndNoOverwrite(t *testing.T) {
	svc, _ := moduleconfig.NewService(
		&fileExistsStub{},
		&fileGeneratorErrorStub{},
	)

	result := svc.ForceExplicitOverwrite(directory, moduleConfigFile, false)

	require.ErrorIs(t, result, moduleconfig.ErrFileExists)
}

func Test_ForceExplicitOverwrite_ReturnsNil_WhenFileExistsAndOverwrite(t *testing.T) {
	svc, _ := moduleconfig.NewService(
		&fileExistsStub{},
		&fileGeneratorErrorStub{},
	)

	result := svc.ForceExplicitOverwrite(directory, moduleConfigFile, true)

	require.NoError(t, result)
}

func Test_ForceExplicitOverwrite_ReturnsNil_WhenFileDoesNotExist(t *testing.T) {
	svc, _ := moduleconfig.NewService(
		&fileDoesNotExistStub{},
		&fileGeneratorErrorStub{},
	)

	result := svc.ForceExplicitOverwrite(directory, moduleConfigFile, true)

	require.NoError(t, result)
}

func Test_GenerateFile_ReturnsError_WhenFileGeneratorReturnsError(t *testing.T) {
	svc, _ := moduleconfig.NewService(
		&errorStub{},
		&fileGeneratorErrorStub{},
	)

	result := svc.GenerateFile(nil, moduleConfigFile, types.KeyValueArgs{})

	require.ErrorIs(t, result, errSomeFileGeneratorError)
}

func Test_GenerateFile_Succeeds(t *testing.T) {
	svc, _ := moduleconfig.NewService(
		&errorStub{},
		&fileGeneratorStub{},
	)

	result := svc.GenerateFile(nil, moduleConfigFile, types.KeyValueArgs{})

	require.NoError(t, result)
}

// ***************
// Test Stubs
// ***************

type fileExistsStub struct{}

func (*fileExistsStub) FileExists(_ string) (bool, error) {
	return true, nil
}

type fileDoesNotExistStub struct{}

func (*fileDoesNotExistStub) FileExists(_ string) (bool, error) {
	return false, nil
}

var errSomeOSError = errors.New("some OS error")

type errorStub struct{}

func (*errorStub) FileExists(_ string) (bool, error) {
	return false, errSomeOSError
}

type fileGeneratorStub struct{}

func (*fileGeneratorStub) GenerateFile(_ iotools.Out, _ string, _ types.KeyValueArgs) error {
	return nil
}

type fileGeneratorErrorStub struct{}

var errSomeFileGeneratorError = errors.New("some file generator error")

func (*fileGeneratorErrorStub) GenerateFile(out iotools.Out, path string, args types.KeyValueArgs) error {
	return errSomeFileGeneratorError
}
