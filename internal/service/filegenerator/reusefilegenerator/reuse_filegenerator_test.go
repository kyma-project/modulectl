package reusefilegenerator_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	commonerrors "github.com/kyma-project/modulectl/internal/common/errors"
	"github.com/kyma-project/modulectl/internal/common/types"
	"github.com/kyma-project/modulectl/internal/service/filegenerator/reusefilegenerator"
	iotools "github.com/kyma-project/modulectl/tools/io"
)

func Test_NewService_ReturnsError_WhenKindIsEmpty(t *testing.T) {
	_, err := reusefilegenerator.NewService("", &fileExistsErrorStub{}, &fileGeneratorErrorStub{})

	require.ErrorIs(t, err, commonerrors.ErrInvalidArg)
	assert.Contains(t, err.Error(), "kind")
}

func Test_NewService_ReturnsError_WhenFileSystemIsNil(t *testing.T) {
	_, err := reusefilegenerator.NewService("test-kind", nil, &fileGeneratorErrorStub{})

	require.ErrorIs(t, err, commonerrors.ErrInvalidArg)
	assert.Contains(t, err.Error(), "fileSystem")
}

func Test_NewService_ReturnsError_WhenFileGeneratorIsNil(t *testing.T) {
	_, err := reusefilegenerator.NewService("test-kind", &fileExistsErrorStub{}, nil)

	require.ErrorIs(t, err, commonerrors.ErrInvalidArg)
	assert.Contains(t, err.Error(), "fileGenerator")
}

func Test_Service_GenerateFile_ReturnsError_WhenFileExistenceCheckReturnsError(t *testing.T) {
	svc, _ := reusefilegenerator.NewService("test-kind", &fileExistsErrorStub{}, &fileGeneratorErrorStub{})

	result := svc.GenerateFile(&rfgTestOut{}, "some-path", nil)

	require.ErrorIs(t, result, reusefilegenerator.ErrCheckingFileExistence)
	require.ErrorIs(t, result, errSomeFileExistsOSError)
}

func Test_GenerateFile_Succeeds_WhenFileDoesAlreadyExist(t *testing.T) {
	out := &rfgTestOut{}
	svc, _ := reusefilegenerator.NewService("test-kind", &fileExistsStub{}, &fileGeneratorErrorStub{})

	result := svc.GenerateFile(out, "some-path", nil)

	require.NoError(t, result)
	require.Len(t, out.sink, 1)
	assert.Contains(t, out.sink[0], "The test-kind file already exists, reusing:")
}

func Test_GenerateFile_ReturnsError_WhenFileGenerationReturnsError(t *testing.T) {
	svc, _ := reusefilegenerator.NewService("test-kind", &fileDoesNotExistStub{}, &fileGeneratorErrorStub{})

	result := svc.GenerateFile(&rfgTestOut{}, "some-path", nil)

	require.ErrorIs(t, result, errSomeFileGeneratorError)
}

func Test_GenerateFile_Succeeds_WhenFileIsGenerated(t *testing.T) {
	svc, _ := reusefilegenerator.NewService("test-kind", &fileExistsStub{}, &fileGeneratorStub{})

	result := svc.GenerateFile(&rfgTestOut{}, "some-path", nil)

	require.NoError(t, result)
}

// ***************
// Test Stubs
// ***************

type rfgTestOut struct {
	sink []string
}

func (o *rfgTestOut) Write(msg string) {
	o.sink = append(o.sink, msg)
}

type fileGeneratorStub struct{}

func (*fileGeneratorStub) GenerateFile(_ iotools.Out, _ string, _ types.KeyValueArgs) error {
	return nil
}

type fileGeneratorErrorStub struct{}

var errSomeFileGeneratorError = errors.New("some file generator error")

func (*fileGeneratorErrorStub) GenerateFile(_ iotools.Out, _ string, _ types.KeyValueArgs) error {
	return errSomeFileGeneratorError
}

type fileExistsErrorStub struct{}

var errSomeFileExistsOSError = errors.New("some os error")

func (*fileExistsErrorStub) FileExists(_ string) (bool, error) {
	return false, errSomeFileExistsOSError
}

type fileExistsStub struct{}

func (*fileExistsStub) FileExists(_ string) (bool, error) {
	return true, nil
}

type fileDoesNotExistStub struct{}

func (*fileDoesNotExistStub) FileExists(_ string) (bool, error) {
	return false, nil
}
