package moduleconfigreader_test

import (
	"errors"
	"fmt"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	commonerrors "github.com/kyma-project/modulectl/internal/common/errors"
	"github.com/kyma-project/modulectl/internal/service/contentprovider"
	moduleconfigreader "github.com/kyma-project/modulectl/internal/service/moduleconfig/reader"
)

const (
	moduleConfigFile = "config.yaml"
)

func Test_ParseModuleConfig_ReturnsError_WhenFileReaderReturnsError(t *testing.T) {
	result, err := moduleconfigreader.ParseModuleConfig(moduleConfigFile, &fileDoesNotExistStub{})

	require.ErrorIs(t, err, errReadingFile)
	assert.Nil(t, result)
}

func Test_ParseModuleConfig_Returns_CorrectModuleConfig(t *testing.T) {
	result, err := moduleconfigreader.ParseModuleConfig(moduleConfigFile, &fileExistsStub{})

	require.NoError(t, err)
	assert.Equal(t, "github.com/module-name", result.Name)
	assert.Equal(t, "0.0.1", result.Version)
	assert.Equal(t, "regular", result.Channel)
	assert.Equal(t, "path/to/manifests", result.ManifestPath)
	assert.Equal(t, "path/to/defaultCR", result.DefaultCRPath)
	assert.Equal(t, "module-name-0.0.1", result.ResourceName)
	assert.False(t, result.Mandatory)
	assert.Equal(t, "kcp-system", result.Namespace)
	assert.Equal(t, "path/to/securityConfig", result.Security)
	assert.False(t, result.Internal)
	assert.False(t, result.Beta)
	assert.Equal(t, map[string]string{"label1": "value1"}, result.Labels)
	assert.Equal(t, map[string]string{"annotation1": "value1"}, result.Annotations)
}

func Test_GetDefaultCRData_Returns_CorrectData(t *testing.T) {
	moduleConfigService, err := moduleconfigreader.NewService(
		&fileExistsStub{},
		&tmpfileSystemStub{})
	require.NoError(t, err)

	result, err := moduleConfigService.GetDefaultCRData("/path/to/defaultcr")
	require.NoError(t, err)

	expected, err := yaml.Marshal(expectedReturnedModuleConfig)
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func Test_GetDefaultCRPath_Returns_CorrectPath(t *testing.T) {
	result, err := moduleconfigreader.GetDefaultCRPath("https://example.com/path", &tmpfileSystemStub{})

	require.NoError(t, err)
	assert.Equal(t, "file.yaml", result)
}

func Test_GetDefaultCRPath_Returns_CorrectPath_When_NotUrl(t *testing.T) {
	result, err := moduleconfigreader.GetDefaultCRPath("/path/to/defaultcr.yaml", &tmpfileSystemStub{})

	require.NoError(t, err)
	assert.Equal(t, "/path/to/defaultcr.yaml", result)
}

func Test_GetManifestPath_Returns_CorrectPath(t *testing.T) {
	result, err := moduleconfigreader.GetDefaultCRPath("https://example.com/path", &tmpfileSystemStub{})

	require.NoError(t, err)
	assert.Equal(t, "file.yaml", result)
}

func Test_GetManifestPath_Returns_CorrectPath_When_NotUrl(t *testing.T) {
	result, err := moduleconfigreader.GetDefaultCRPath("/path/to/manifest.yaml", &tmpfileSystemStub{})

	require.NoError(t, err)
	assert.Equal(t, "/path/to/manifest.yaml", result)
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
			expectedError: fmt.Errorf("%w: parsing url failed for invalid-url", commonerrors.ErrInvalidArg),
		},
		{
			name:          "URL without Scheme",
			urlString:     "example.com/path",
			want:          nil,
			expectedError: fmt.Errorf("%w: parsing url failed for example.com/path", commonerrors.ErrInvalidArg),
		},
		{
			name:          "URL without Host",
			urlString:     "https://",
			want:          nil,
			expectedError: fmt.Errorf("%w: parsing url failed for https://", commonerrors.ErrInvalidArg),
		},
		{
			name:          "Empty URL",
			urlString:     "",
			want:          nil,
			expectedError: fmt.Errorf("%w: parsing url failed for ", commonerrors.ErrInvalidArg),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := moduleconfigreader.ParseURL(test.urlString)

			if test.expectedError != nil {
				require.EqualError(t, err, test.expectedError.Error())
				return
			}
			assert.Equalf(t, test.want, got, "ParseURL(%v)", test.urlString)
		})
	}
}

func Test_ValidateModuleConfig(t *testing.T) {
	tests := []struct {
		name          string
		moduleConfig  *contentprovider.ModuleConfig
		expectedError error
	}{
		{
			name:          "valid module config",
			moduleConfig:  &expectedReturnedModuleConfig,
			expectedError: nil,
		},
		{
			name: "invalid module name",
			moduleConfig: &contentprovider.ModuleConfig{
				Name:         "invalid name",
				Version:      "0.0.1",
				Channel:      "regular",
				Namespace:    "kcp-system",
				ManifestPath: "test",
			},
			expectedError: fmt.Errorf("failed to validate module name: %w", commonerrors.ErrInvalidOption),
		},
		{
			name: "invalid module version",
			moduleConfig: &contentprovider.ModuleConfig{
				Name:         "github.com/module-name",
				Version:      "invalid version",
				Channel:      "regular",
				Namespace:    "kcp-system",
				ManifestPath: "test",
			},
			expectedError: fmt.Errorf("failed to validate module version: %w", commonerrors.ErrInvalidOption),
		},
		{
			name: "invalid module channel",
			moduleConfig: &contentprovider.ModuleConfig{
				Name:         "github.com/module-name",
				Version:      "0.0.1",
				Channel:      "invalid channel",
				Namespace:    "kcp-system",
				ManifestPath: "test",
			},
			expectedError: fmt.Errorf("failed to validate module channel: %w", commonerrors.ErrInvalidOption),
		},
		{
			name: "invalid module namespace",
			moduleConfig: &contentprovider.ModuleConfig{
				Name:         "github.com/module-name",
				Version:      "0.0.1",
				Channel:      "regular",
				Namespace:    "invalid namespace",
				ManifestPath: "test",
			},
			expectedError: fmt.Errorf("failed to validate module namespace: %w", commonerrors.ErrInvalidOption),
		},
		{
			name: "empty manifest path",
			moduleConfig: &contentprovider.ModuleConfig{
				Name:         "github.com/module-name",
				Version:      "0.0.1",
				Channel:      "regular",
				Namespace:    "kcp-system",
				ManifestPath: "",
			},
			expectedError: fmt.Errorf("%w: manifest path must not be empty", commonerrors.ErrInvalidOption),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := moduleconfigreader.ValidateModuleConfig(test.moduleConfig)
			if test.expectedError != nil {
				require.ErrorContains(t, err, test.expectedError.Error())
				return
			}
			require.NoError(t, err)
		})
	}
}

// Test Stubs

type fileExistsStub struct{}

func (*fileExistsStub) FileExists(_ string) (bool, error) {
	return true, nil
}

var expectedReturnedModuleConfig = contentprovider.ModuleConfig{
	Name:          "github.com/module-name",
	Version:       "0.0.1",
	Channel:       "regular",
	ManifestPath:  "path/to/manifests",
	Mandatory:     false,
	DefaultCRPath: "path/to/defaultCR",
	ResourceName:  "module-name-0.0.1",
	Namespace:     "kcp-system",
	Security:      "path/to/securityConfig",
	Internal:      false,
	Beta:          false,
	Labels:        map[string]string{"label1": "value1"},
	Annotations:   map[string]string{"annotation1": "value1"},
}

func (*fileExistsStub) ReadFile(_ string) ([]byte, error) {
	return yaml.Marshal(expectedReturnedModuleConfig)
}

type tmpfileSystemStub struct{}

func (*tmpfileSystemStub) DownloadTempFile(_ string, _ string, _ *url.URL) (string, error) {
	return "file.yaml", nil
}

func (*tmpfileSystemStub) RemoveTempFiles() []error {
	return nil
}

type fileDoesNotExistStub struct{}

func (*fileDoesNotExistStub) FileExists(_ string) (bool, error) {
	return false, nil
}

var errReadingFile = errors.New("some error reading file")

func (*fileDoesNotExistStub) ReadFile(_ string) ([]byte, error) {
	return nil, errReadingFile
}
