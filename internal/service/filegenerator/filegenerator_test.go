package filegenerator_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	commonerrors "github.com/kyma-project/modulectl/internal/common/errors"
	"github.com/kyma-project/modulectl/internal/common/types"
	"github.com/kyma-project/modulectl/internal/service/filegenerator"
)

func Test_NewService_ReturnsError_WhenKindIsEmpty(t *testing.T) {
	_, err := filegenerator.NewService("", &fileWriterErrorStub{}, &contentProviderErrorStub{})

	require.ErrorIs(t, err, commonerrors.ErrInvalidArg)
	assert.Contains(t, err.Error(), "kind")
}

func Test_NewService_ReturnsError_WhenFileSystemIsNil(t *testing.T) {
	_, err := filegenerator.NewService("test-kind", nil, &contentProviderErrorStub{})

	require.ErrorIs(t, err, commonerrors.ErrInvalidArg)
	assert.Contains(t, err.Error(), "fileSystem")
}

func Test_NewService_ReturnsError_WhenDefaultContentProviderIsNil(t *testing.T) {
	_, err := filegenerator.NewService("test-kind", &fileWriterErrorStub{}, nil)

	require.ErrorIs(t, err, commonerrors.ErrInvalidArg)
	assert.Contains(t, err.Error(), "defaultContentProvider")
}

func Test_GenerateFile_ReturnsError_WhenErrorGettingDefaultContent(t *testing.T) {
	out := &fgTestOut{}
	svc, _ := filegenerator.NewService("test-kind", &fileWriterErrorStub{}, &contentProviderErrorStub{})

	result := svc.GenerateFile(out, "some-path", nil)

	require.ErrorIs(t, result, filegenerator.ErrGettingDefaultContent)
	require.ErrorIs(t, result, errSomeContentProviderError)
	require.Empty(t, out.sink)
}

func Test_GenerateFile_ReturnsError_WhenErrorWritingFile(t *testing.T) {
	out := &fgTestOut{}
	svc, _ := filegenerator.NewService("test-kind", &fileWriterErrorStub{}, &contentProviderStub{})

	result := svc.GenerateFile(out, "some-path", nil)

	require.ErrorIs(t, result, filegenerator.ErrWritingFile)
	require.Empty(t, out.sink)
}

func Test_GenerateFile_Succeeds_WhenFileIsGenerated(t *testing.T) {
	out := &fgTestOut{}
	svc, _ := filegenerator.NewService("test-kind", &fileWriterStub{}, &contentProviderStub{})

	result := svc.GenerateFile(out, "some-path", nil)

	require.NoError(t, result)
	require.Len(t, out.sink, 1)
	assert.Contains(t, out.sink[0], "Generated a blank test-kind file:")
}

// ***************
// Test Stubs
// ***************

type fgTestOut struct {
	sink []string
}

func (o *fgTestOut) Write(msg string) {
	o.sink = append(o.sink, msg)
}

type contentProviderStub struct{}

func (*contentProviderStub) GetDefaultContent(args types.KeyValueArgs) (string, error) {
	return "test-content", nil
}

type contentProviderErrorStub struct{}

var errSomeContentProviderError = errors.New("some error")

func (*contentProviderErrorStub) GetDefaultContent(args types.KeyValueArgs) (string, error) {
	return "", errSomeContentProviderError
}

type fileWriterErrorStub struct{}

var errSomeOSError = errors.New("some error")

func (*fileWriterErrorStub) WriteFile(_, _ string) error {
	return errSomeOSError
}

type fileWriterStub struct{}

func (*fileWriterStub) WriteFile(_, _ string) error {
	return nil
}
