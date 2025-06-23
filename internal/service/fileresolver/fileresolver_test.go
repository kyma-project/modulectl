package fileresolver_test

import (
	"errors"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	commonerrors "github.com/kyma-project/modulectl/internal/common/errors"
	"github.com/kyma-project/modulectl/internal/service/contentprovider"
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

func urlOrFileRef(urlString string) contentprovider.UrlOrLocalFile {
	return contentprovider.MustUrlOrLocalFile(urlString)
}

func Test_Resolve_WhenURL_Returns_CorrectPath(t *testing.T) {
	resolver, _ := fileresolver.NewFileResolver(filePattern, &tmpfileSystemStub{})
	fileRef := urlOrFileRef("https://example.com/path")
	require.True(t, fileRef.IsURL(), "Expected UrlOrLocalFile to be a URL")
	result, err := resolver.Resolve(fileRef, "")

	require.NoError(t, err)
	require.Equal(t, "file.yaml", result)
}

func Test_Resolve_WhenURL_Returns_Error_WhenFailingToDownload(t *testing.T) {
	resolver, _ := fileresolver.NewFileResolver(filePattern, &tempfileSystemErrorStub{})
	fileRef := urlOrFileRef("https://example.com/path")
	require.True(t, fileRef.IsURL(), "Expected UrlOrLocalFile to be a URL")
	result, err := resolver.Resolve(fileRef, "")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to download file")
	assert.Empty(t, result)
}

func Test_Resolve_WhenLocalFile_WithJustFilename(t *testing.T) {
	resolver, _ := fileresolver.NewFileResolver(filePattern, &tmpfileSystemStub{})
	fileRef := urlOrFileRef("manifest.yaml")
	require.False(t, fileRef.IsURL(), "Expected UrlOrLocalFile to be a file name")
	result, err := resolver.Resolve(fileRef, "")

	require.NoError(t, err)
	assert.Equal(t, "manifest.yaml", result)
}

func Test_Resolve_WhenLocalFile_WithAbsolutePath(t *testing.T) {
	resolver, _ := fileresolver.NewFileResolver(filePattern, &tmpfileSystemStub{})
	fileRef := urlOrFileRef("/path/to/manifest.yaml")
	require.False(t, fileRef.IsURL(), "Expected UrlOrLocalFile to be a local file path")
	result, err := resolver.Resolve(fileRef, "")

	require.NoError(t, err)
	assert.Equal(t, "/path/to/manifest.yaml", result)
}

func Test_Resolve_WhenLocalFile_WithRelativePath(t *testing.T) {
	resolver, _ := fileresolver.NewFileResolver(filePattern, &tmpfileSystemStub{})
	fileRef := urlOrFileRef("path/to/manifest.yaml")
	require.False(t, fileRef.IsURL(), "Expected UrlOrLocalFile to be a local file path")
	result, err := resolver.Resolve(fileRef, "")

	require.NoError(t, err)
	assert.Equal(t, "path/to/manifest.yaml", result)
}

func Test_Resolve_WhenLocalFile_RelativeToCurrentDir(t *testing.T) {
	resolver, _ := fileresolver.NewFileResolver(filePattern, &tmpfileSystemStub{})
	fileRef := urlOrFileRef("./path/to/manifest.yaml")
	require.False(t, fileRef.IsURL(), "Expected UrlOrLocalFile to be a local file path")
	result, err := resolver.Resolve(fileRef, "")

	require.NoError(t, err)
	assert.Equal(t, "path/to/manifest.yaml", result)
}

func Test_Resolve_WhenLocalFile_IsEmpty(t *testing.T) {
	resolver, _ := fileresolver.NewFileResolver(filePattern, &tmpfileSystemStub{})
	fileRef := urlOrFileRef("")
	require.False(t, fileRef.IsURL(), "Expected UrlOrLocalFile to be a local file path")
	_, err := resolver.Resolve(fileRef, "")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "file reference is empty: invalid argument")
}

func Test_Resolve_WhenLocalFile_WithRelativePath_WithRelativeBasePath(t *testing.T) {
	resolver, _ := fileresolver.NewFileResolver(filePattern, &tmpfileSystemStub{})
	fileRef := urlOrFileRef("path/to/manifest.yaml")
	require.False(t, fileRef.IsURL(), "Expected UrlOrLocalFile to be a local file path")

	result, err := resolver.Resolve(fileRef, "relative/base")
	require.NoError(t, err)
	assert.Equal(t, "relative/base/path/to/manifest.yaml", result)

	result, err = resolver.Resolve(fileRef, "relative/base/")
	require.NoError(t, err)
	assert.Equal(t, "relative/base/path/to/manifest.yaml", result)
}

func Test_Resolve_WhenLocalFile_WithRelativePath_WithAbsoluteBasePath(t *testing.T) {
	resolver, _ := fileresolver.NewFileResolver(filePattern, &tmpfileSystemStub{})
	fileRef := urlOrFileRef("path/to/manifest.yaml")
	require.False(t, fileRef.IsURL(), "Expected UrlOrLocalFile to be a local file path")

	result, err := resolver.Resolve(fileRef, "/absolute/base")
	require.NoError(t, err)
	assert.Equal(t, "/absolute/base/path/to/manifest.yaml", result)

	result, err = resolver.Resolve(fileRef, "/absolute/base/")
	require.NoError(t, err)
	assert.Equal(t, "/absolute/base/path/to/manifest.yaml", result)
}

func Test_Resolve_WhenLocalFile_WithJustFilename_WithSimpleBasePath(t *testing.T) {
	resolver, _ := fileresolver.NewFileResolver(filePattern, &tmpfileSystemStub{})
	fileRef := urlOrFileRef("manifest.yaml")
	require.False(t, fileRef.IsURL(), "Expected UrlOrLocalFile to be a local file path")

	result, err := resolver.Resolve(fileRef, "base")
	require.NoError(t, err)
	assert.Equal(t, "base/manifest.yaml", result)

	result, err = resolver.Resolve(fileRef, "base/")
	require.NoError(t, err)
	assert.Equal(t, "base/manifest.yaml", result)
}

func Test_Resolve_WhenLocalFile_Returns_Error_WhenFailingToCheckIfExists(t *testing.T) {
	resolver, _ := fileresolver.NewFileResolver(filePattern, &tempfileSystemErrorStub{})
	fileRef := urlOrFileRef("manifest.yaml")
	require.False(t, fileRef.IsURL(), "Expected UrlOrLocalFile to be a local file path")
	result, err := resolver.Resolve(fileRef, "")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "error checking if file exists")
	assert.Empty(t, result)
}

func Test_Resolve_WhenLocalFile_Returns_Error_WhenFileDoesNotExist(t *testing.T) {
	resolver, _ := fileresolver.NewFileResolver(filePattern, &tempfileSystemErrorStub{})
	fileRef := urlOrFileRef("notexists-manifest.yaml")
	require.False(t, fileRef.IsURL(), "Expected UrlOrLocalFile to be a local file path")
	result, err := resolver.Resolve(fileRef, "")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "file does not exist: notexists-manifest.yaml: invalid argument")
	assert.Empty(t, result)
}

type tmpfileSystemStub struct{}

func (*tmpfileSystemStub) DownloadTempFile(_ string, _ string, _ *url.URL) (string, error) {
	return "file.yaml", nil
}

func (s *tmpfileSystemStub) FileExists(filePath string) (bool, error) {
	if strings.Contains(filePath, "path/to/manifest.yaml") || strings.Contains(filePath, "manifest.yaml") {
		return true, nil
	}
	return false, nil
}

func (s *tmpfileSystemStub) RemoveTempFiles() []error {
	return nil
}

type tempfileSystemErrorStub struct{}

func (*tempfileSystemErrorStub) DownloadTempFile(_ string, _ string, _ *url.URL) (string, error) {
	return "", errors.New("error downloading file")
}

func (s *tempfileSystemErrorStub) FileExists(filePath string) (bool, error) {
	if strings.HasPrefix(filePath, "notexist") {
		return false, nil
	}
	return false, errors.New("error checking if file exists")
}

func (s *tempfileSystemErrorStub) RemoveTempFiles() []error {
	return nil
}
