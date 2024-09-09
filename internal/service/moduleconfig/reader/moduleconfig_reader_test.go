package moduleconfigreader_test

import (
	"errors"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/kyma-project/modulectl/internal/service/contentprovider"
	moduleconfigreader "github.com/kyma-project/modulectl/internal/service/moduleconfig/reader"
)

const (
	moduleConfigFile = "config.yaml"
)

func Test_ParseModuleConfig_ReturnsError_WhenFileReaderReturnsError(t *testing.T) {
	svc, _ := moduleconfigreader.NewService(
		&fileDoesNotExistStub{},
		&tmpfileSystemStub{},
	)

	result, err := svc.ParseModuleConfig(moduleConfigFile)

	require.ErrorIs(t, err, errReadingFile)
	assert.Nil(t, result)
}

func Test_ParseModuleConfig_ReturnsCorrect_ModuleConfig(t *testing.T) {
	svc, _ := moduleconfigreader.NewService(
		&fileExistsStub{},
		&tmpfileSystemStub{},
	)

	result, err := svc.ParseModuleConfig(moduleConfigFile)

	require.NoError(t, err)
	assert.Equal(t, "module-name", result.Name)
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

// Test Stubs

type fileExistsStub struct{}

func (*fileExistsStub) FileExists(_ string) (bool, error) {
	return true, nil
}

func (*fileExistsStub) ReadFile(_ string) ([]byte, error) {
	moduleConfig := contentprovider.ModuleConfig{
		Name:          "module-name",
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

	return yaml.Marshal(moduleConfig)
}

type tmpfileSystemStub struct{}

func (*tmpfileSystemStub) DownloadTempFile(_ string, _ string, _ *url.URL) (string, error) {
	return "test", nil
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
