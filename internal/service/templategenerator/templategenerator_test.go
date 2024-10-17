package templategenerator_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kyma-project/modulectl/internal/service/contentprovider"
	"github.com/kyma-project/modulectl/internal/service/templategenerator"
	"github.com/kyma-project/modulectl/internal/testutils"

	_ "ocm.software/ocm/api/ocm/compdesc/versions/v2"
)

func TestNew_WhenCalledWithNilDependencies_ReturnsError(t *testing.T) {
	_, err := templategenerator.NewService(nil)

	require.Error(t, err)
}

func TestGenerateModuleTemplate_WhenCalledWithNilConfig_ReturnsError(t *testing.T) {
	svc, _ := templategenerator.NewService(&mockFileSystem{})

	err := svc.GenerateModuleTemplate(nil, nil, nil, false, "")

	require.Error(t, err)
	require.ErrorIs(t, err, templategenerator.ErrEmptyModuleConfig)
}

func TestGenerateModuleTemplate_WhenCalledWithNilDescriptor_ReturnsError(t *testing.T) {
	svc, _ := templategenerator.NewService(&mockFileSystem{})

	err := svc.GenerateModuleTemplate(&contentprovider.ModuleConfig{}, nil, nil, false, "")

	require.Error(t, err)
	require.ErrorIs(t, err, templategenerator.ErrEmptyDescriptor)
}

func TestGenerateModuleTemplate_Success(t *testing.T) {
	mockFS := &mockFileSystem{}
	svc, _ := templategenerator.NewService(mockFS)

	moduleConfig := &contentprovider.ModuleConfig{
		ResourceName: "test-resource",
		Namespace:    "default",
		Channel:      "stable",
		Labels:       map[string]string{"key": "value"},
		Annotations:  map[string]string{"annotation": "value"},
		Mandatory:    true,
		Manifest:     "https://github.com/kyma-project/template-operator/releases/download/1.0.1/template-operator.yaml",
		Resources:    contentprovider.ResourcesMap{"someResource": "https://some.other/location/template-operator.yaml"},
	}
	descriptor := testutils.CreateComponentDescriptor("example.com/component", "1.0.0")
	data := []byte("test-data")

	err := svc.GenerateModuleTemplate(moduleConfig, descriptor, data, true, "output.yaml")

	require.NoError(t, err)
	require.Equal(t, "output.yaml", mockFS.path)
	require.Contains(t, mockFS.writtenTemplate, "test-resource")
	require.Contains(t, mockFS.writtenTemplate, "default")
	require.Contains(t, mockFS.writtenTemplate, "stable")
	require.Contains(t, mockFS.writtenTemplate, "test-data")
	require.Contains(t, mockFS.writtenTemplate, "example.com/component")
	require.Contains(t, mockFS.writtenTemplate, "someResource")
	require.Contains(t, mockFS.writtenTemplate, "https://some.other/location/template-operator.yaml")
	require.Contains(t, mockFS.writtenTemplate, "rawManifest")
	require.Contains(t, mockFS.writtenTemplate, "https://github.com/kyma-project/template-operator/releases/download/1.0.1/template-operator.yaml")
}

func TestGenerateModuleTemplate_Success_With_Overwritten_RawManifest(t *testing.T) {
	mockFS := &mockFileSystem{}
	svc, _ := templategenerator.NewService(mockFS)

	moduleConfig := &contentprovider.ModuleConfig{
		Manifest:  "https://github.com/kyma-project/template-operator/releases/download/1.0.1/template-operator.yaml",
		Resources: contentprovider.ResourcesMap{"rawManifest": "https://some.other/location/template-operator.yaml"},
	}
	descriptor := testutils.CreateComponentDescriptor("example.com/component", "1.0.0")
	data := []byte("test-data")

	err := svc.GenerateModuleTemplate(moduleConfig, descriptor, data, true, "output.yaml")

	require.NoError(t, err)
	require.Equal(t, "output.yaml", mockFS.path)
	require.Contains(t, mockFS.writtenTemplate, "rawManifest")
	require.Contains(t, mockFS.writtenTemplate, "https://some.other/location/template-operator.yaml")
	require.NotContains(t, mockFS.writtenTemplate, "https://github.com/kyma-project/template-operator/releases/download/1.0.1/template-operator.yaml")
}

type mockFileSystem struct {
	path, writtenTemplate string
}

func (m *mockFileSystem) WriteFile(path, content string) error {
	m.path = path
	m.writtenTemplate = content
	return nil
}
