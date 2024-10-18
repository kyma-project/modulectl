package fileresolver_test

import (
	"fmt"
	"net/url"
	"testing"

	commonerrors "github.com/kyma-project/modulectl/internal/common/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kyma-project/modulectl/internal/service/fileresolver"
)

const filePattern = "kyma-module-manifest-*.yaml"

func TestNew_CalledWithEmptyFilePattern_ReturnsErr(t *testing.T) {
	_, err := fileresolver.NewFileResolver("", &tmpfileSystemStub{})
	require.ErrorIs(t, err, commonerrors.ErrInvalidArg)
	assert.Contains(t, err.Error(), "filePattern must not be empty")
}

func TestNew_CalledWithNilDependencies_ReturnsErr(t *testing.T) {
	_, err := fileresolver.NewFileResolver(filePattern, nil)
	require.ErrorIs(t, err, commonerrors.ErrInvalidArg)
	assert.Contains(t, err.Error(), "tempFileSystem must not be nil")
}

func TestCleanupTempFiles_CalledWithNoTempFiles_ReturnsNoErrors(t *testing.T) {
	resolver, _ := fileresolver.NewFileResolver(filePattern, &tmpfileSystemStub{})

	errs := resolver.CleanupTempFiles()
	assert.Empty(t, errs)
}

func Test_Resolve_Returns_CorrectPath(t *testing.T) {
	resolver, _ := fileresolver.NewFileResolver(filePattern, &tmpfileSystemStub{})
	result, err := resolver.Resolve("https://example.com/path")

	require.NoError(t, err)
	require.Equal(t, "file.yaml", result)
}

func Test_Resolve_Returns_Error_WhenFailingToDownload(t *testing.T) {
	resolver, _ := fileresolver.NewFileResolver(filePattern, &tempfileSystemErrorStub{})
	result, err := resolver.Resolve("https://example.com/path")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to download file")
	assert.Empty(t, result)
}

func Test_Resolve_Returns_CorrectPath_When_AbsolutePath(t *testing.T) {
	resolver, _ := fileresolver.NewFileResolver(filePattern, &tmpfileSystemStub{})
	result, err := resolver.Resolve("/path/to/manifest.yaml")

	require.NoError(t, err)
	assert.Equal(t, "/path/to/manifest.yaml", result)
}

func Test_Resolve_Returns_CorrectPath_When_Relative(t *testing.T) {
	resolver, _ := fileresolver.NewFileResolver(filePattern, &tmpfileSystemStub{})
	result, err := resolver.Resolve("./path/to/manifest.yaml")

	require.NoError(t, err)
	assert.Contains(t, result, "/path/to/manifest.yaml")
	assert.Equal(t, rune(result[0]), '/')
}

func TestService_ParseURL(t *testing.T) {
	tests := []struct {
		name          string
		urlString     string
		want          *url.URL
		expectedError error
	}{
		{
			name:      "valid URL",
			urlString: "https://example.com/path",
			want: &url.URL{
				Scheme: "https",
				Host:   "example.com",
				Path:   "/path",
			},
			expectedError: nil,
		},
		{
			name:          "invalid URL",
			urlString:     "invalid-url",
			want:          nil,
			expectedError: fmt.Errorf("failed to parse url invalid-url: %w", commonerrors.ErrInvalidArg),
		},
		{
			name:          "URL without Scheme",
			urlString:     "example.com/path",
			want:          nil,
			expectedError: fmt.Errorf("failed to parse url example.com/path: %w", commonerrors.ErrInvalidArg),
		},
		{
			name:          "URL without Host",
			urlString:     "https://",
			want:          nil,
			expectedError: fmt.Errorf("failed to parse url https://: %w", commonerrors.ErrInvalidArg),
		},
		{
			name:          "Empty URL",
			urlString:     "",
			want:          nil,
			expectedError: fmt.Errorf("failed to parse url : %w", commonerrors.ErrInvalidArg),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resolver, _ := fileresolver.NewFileResolver(filePattern, &tmpfileSystemStub{})
			got, err := resolver.ParseURL(test.urlString)

			if test.expectedError != nil {
				require.EqualError(t, err, test.expectedError.Error())
				return
			}
			require.Equalf(t, test.want, got, "ParseURL(%v)", test.urlString)
		})
	}
}

type tmpfileSystemStub struct{}

func (*tmpfileSystemStub) DownloadTempFile(_ string, _ string, _ *url.URL) (string, error) {
	return "file.yaml", nil
}

func (s *tmpfileSystemStub) RemoveTempFiles() []error {
	return nil
}

type tempfileSystemErrorStub struct{}

func (*tempfileSystemErrorStub) DownloadTempFile(_ string, _ string, _ *url.URL) (string, error) {
	return "", fmt.Errorf("error downloading file")
}

func (s *tempfileSystemErrorStub) RemoveTempFiles() []error {
	return nil
}
