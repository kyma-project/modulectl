package templategenerator_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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
		Namespace:   "default",
		Version:     "1.0.0",
		Labels:      map[string]string{"key": "value"},
		Annotations: map[string]string{"annotation": "value"},
		Mandatory:   true,
		Manifest:    "https://github.com/kyma-project/template-operator/releases/download/1.0.1/template-operator.yaml",
		Resources:   contentprovider.Resources{"someResource": "https://some.other/location/template-operator.yaml"},
	}
	descriptor := testutils.CreateComponentDescriptor("example.com/component", "1.0.0")
	data := []byte("test-data")

	err := svc.GenerateModuleTemplate(moduleConfig, descriptor, data, true, "output.yaml")

	require.NoError(t, err)
	require.Equal(t, "output.yaml", mockFS.path)
	require.Contains(t, mockFS.writtenTemplate, "version: 1.0.0")
	require.Contains(t, mockFS.writtenTemplate, "moduleName: component")
	require.Contains(t, mockFS.writtenTemplate, "component-1.0.0")
	require.Contains(t, mockFS.writtenTemplate, "default")
	require.Contains(t, mockFS.writtenTemplate, "test-data")
	require.Contains(t, mockFS.writtenTemplate, "example.com/component")
	require.Contains(t, mockFS.writtenTemplate, "someResource")
	require.Contains(t, mockFS.writtenTemplate, "https://some.other/location/template-operator.yaml")
	require.Contains(t, mockFS.writtenTemplate, "rawManifest")
	require.Contains(t, mockFS.writtenTemplate,
		"https://github.com/kyma-project/template-operator/releases/download/1.0.1/template-operator.yaml")
}

func TestGenerateModuleTemplate_Success_With_Overwritten_RawManifest(t *testing.T) {
	mockFS := &mockFileSystem{}
	svc, _ := templategenerator.NewService(mockFS)

	moduleConfig := &contentprovider.ModuleConfig{
		Manifest:  "https://github.com/kyma-project/template-operator/releases/download/1.0.1/template-operator.yaml",
		Resources: contentprovider.Resources{"rawManifest": "https://some.other/location/template-operator.yaml"},
	}
	descriptor := testutils.CreateComponentDescriptor("example.com/component", "1.0.0")
	data := []byte("test-data")

	err := svc.GenerateModuleTemplate(moduleConfig, descriptor, data, true, "output.yaml")

	require.NoError(t, err)
	require.Equal(t, "output.yaml", mockFS.path)
	require.Contains(t, mockFS.writtenTemplate, "rawManifest")
	require.Contains(t, mockFS.writtenTemplate, "https://some.other/location/template-operator.yaml")
	require.NotContains(t, mockFS.writtenTemplate,
		"https://github.com/kyma-project/template-operator/releases/download/1.0.1/template-operator.yaml")
}

func TestGenerateModuleTemplateWithAssociatedResources_Success(t *testing.T) {
	mockFS := &mockFileSystem{}
	svc, _ := templategenerator.NewService(mockFS)

	moduleConfig := &contentprovider.ModuleConfig{
		Namespace:   "default",
		Labels:      map[string]string{"key": "value"},
		Annotations: map[string]string{"annotation": "value"},
		Mandatory:   true,
		AssociatedResources: []*metav1.GroupVersionKind{
			{
				Group:   "networking.istio.io",
				Version: "v1alpha3",
				Kind:    "Gateway",
			},
		},
	}
	descriptor := testutils.CreateComponentDescriptor("example.com/component", "1.0.0")
	data := []byte("test-data")

	err := svc.GenerateModuleTemplate(moduleConfig, descriptor, data, true, "output.yaml")

	require.NoError(t, err)
	require.Equal(t, "output.yaml", mockFS.path)
	require.Contains(t, mockFS.writtenTemplate, "default")
	require.Contains(t, mockFS.writtenTemplate, "test-data")
	require.Contains(t, mockFS.writtenTemplate, "example.com/component")
	require.Contains(t, mockFS.writtenTemplate, "associatedResources")
	require.Contains(t, mockFS.writtenTemplate, "networking.istio.io")
	require.Contains(t, mockFS.writtenTemplate, "v1alpha3")
	require.Contains(t, mockFS.writtenTemplate, "Gateway")
}

func TestGenerateModuleTemplateWithManager_Success(t *testing.T) {
	mockFS := &mockFileSystem{}
	svc, _ := templategenerator.NewService(mockFS)

	moduleConfig := &contentprovider.ModuleConfig{
		Namespace:   "default",
		Version:     "1.0.0",
		Labels:      map[string]string{"key": "value"},
		Annotations: map[string]string{"annotation": "value"},
		Mandatory:   true,
		Manager: &contentprovider.Manager{
			Name:      "manager-name",
			Namespace: "manager-ns",
			GroupVersionKind: metav1.GroupVersionKind{
				Group:   "apps",
				Version: "v1",
				Kind:    "Deployment",
			},
		},
	}
	descriptor := testutils.CreateComponentDescriptor("example.com/component", "1.0.0")
	data := []byte("test-data")

	err := svc.GenerateModuleTemplate(moduleConfig, descriptor, data, true, "output.yaml")

	require.NoError(t, err)
	require.Equal(t, "output.yaml", mockFS.path)
	require.Contains(t, mockFS.writtenTemplate, "component-1.0.0")
	require.Contains(t, mockFS.writtenTemplate, "default")
	require.Contains(t, mockFS.writtenTemplate, "test-data")
	require.Contains(t, mockFS.writtenTemplate, "example.com/component")
	require.Contains(t, mockFS.writtenTemplate, "manager-name")
	require.Contains(t, mockFS.writtenTemplate, "manager-ns")
	require.Contains(t, mockFS.writtenTemplate, "apps")
	require.Contains(t, mockFS.writtenTemplate, "v1")
	require.Contains(t, mockFS.writtenTemplate, "Deployment")
	require.Equal(t, 2, strings.Count(mockFS.writtenTemplate, "namespace"))
}

func TestGenerateModuleTemplateWithManagerWithoutNamespace_Success(t *testing.T) {
	mockFS := &mockFileSystem{}
	svc, _ := templategenerator.NewService(mockFS)

	moduleConfig := &contentprovider.ModuleConfig{
		Namespace:   "default",
		Version:     "1.0.0",
		Labels:      map[string]string{"key": "value"},
		Annotations: map[string]string{"annotation": "value"},
		Mandatory:   true,
		Manager: &contentprovider.Manager{
			Name: "manager-name",
			GroupVersionKind: metav1.GroupVersionKind{
				Group:   "apps",
				Version: "v1",
				Kind:    "Deployment",
			},
		},
	}
	descriptor := testutils.CreateComponentDescriptor("example.com/component", "1.0.0")
	data := []byte("test-data")

	err := svc.GenerateModuleTemplate(moduleConfig, descriptor, data, true, "output.yaml")

	require.NoError(t, err)
	require.Equal(t, "output.yaml", mockFS.path)
	require.Contains(t, mockFS.writtenTemplate, "component-1.0.0")
	require.Contains(t, mockFS.writtenTemplate, "default")
	require.Contains(t, mockFS.writtenTemplate, "test-data")
	require.Contains(t, mockFS.writtenTemplate, "example.com/component")
	require.Contains(t, mockFS.writtenTemplate, "manager-name")
	require.Contains(t, mockFS.writtenTemplate, "apps")
	require.Contains(t, mockFS.writtenTemplate, "v1")
	require.Contains(t, mockFS.writtenTemplate, "Deployment")
	require.Equal(t, 1, strings.Count(mockFS.writtenTemplate, "namespace"))
}

func TestGenerateModuleTemplateWithMandatoryTrue_Success(t *testing.T) {
	mockFS := &mockFileSystem{}
	svc, _ := templategenerator.NewService(mockFS)

	moduleConfig := &contentprovider.ModuleConfig{
		Namespace:   "default",
		Version:     "1.0.0",
		Labels:      map[string]string{"key": "value"},
		Annotations: map[string]string{"annotation": "value"},
		Mandatory:   true,
		Manifest:    "https://github.com/kyma-project/template-operator/releases/download/1.0.1/template-operator.yaml",
		Resources:   contentprovider.Resources{"someResource": "https://some.other/location/template-operator.yaml"},
	}
	descriptor := testutils.CreateComponentDescriptor("example.com/component", "1.0.0")
	data := []byte("test-data")

	err := svc.GenerateModuleTemplate(moduleConfig, descriptor, data, true, "output.yaml")

	require.NoError(t, err)
	require.Equal(t, "output.yaml", mockFS.path)
	require.Contains(t, mockFS.writtenTemplate, "version: 1.0.0")
	require.Contains(t, mockFS.writtenTemplate, "moduleName: component")
	require.Contains(t, mockFS.writtenTemplate, "component-1.0.0")
	require.Contains(t, mockFS.writtenTemplate, "default")
	require.Contains(t, mockFS.writtenTemplate, "test-data")
	require.Contains(t, mockFS.writtenTemplate, "example.com/component")
	require.Contains(t, mockFS.writtenTemplate, "someResource")
	require.Contains(t, mockFS.writtenTemplate, "https://some.other/location/template-operator.yaml")
	require.Contains(t, mockFS.writtenTemplate, "rawManifest")
	require.Contains(t, mockFS.writtenTemplate,
		"https://github.com/kyma-project/template-operator/releases/download/1.0.1/template-operator.yaml")
	require.Contains(t, mockFS.writtenTemplate, "mandatory: true")
	require.Contains(t, mockFS.writtenTemplate,
		"\"operator.kyma-project.io/mandatory-module\": \"true\"")
}

func TestGenerateModuleTemplateWithMandatoryFalse_Success(t *testing.T) {
	mockFS := &mockFileSystem{}
	svc, _ := templategenerator.NewService(mockFS)

	moduleConfig := &contentprovider.ModuleConfig{
		Namespace:   "default",
		Version:     "1.0.0",
		Labels:      map[string]string{"key": "value"},
		Annotations: map[string]string{"annotation": "value"},
		Mandatory:   false,
		Manifest:    "https://github.com/kyma-project/template-operator/releases/download/1.0.1/template-operator.yaml",
		Resources:   contentprovider.Resources{"someResource": "https://some.other/location/template-operator.yaml"},
	}
	descriptor := testutils.CreateComponentDescriptor("example.com/component", "1.0.0")
	data := []byte("test-data")

	err := svc.GenerateModuleTemplate(moduleConfig, descriptor, data, true, "output.yaml")

	require.NoError(t, err)
	require.Equal(t, "output.yaml", mockFS.path)
	require.Contains(t, mockFS.writtenTemplate, "version: 1.0.0")
	require.Contains(t, mockFS.writtenTemplate, "moduleName: component")
	require.Contains(t, mockFS.writtenTemplate, "component-1.0.0")
	require.Contains(t, mockFS.writtenTemplate, "default")
	require.Contains(t, mockFS.writtenTemplate, "test-data")
	require.Contains(t, mockFS.writtenTemplate, "example.com/component")
	require.Contains(t, mockFS.writtenTemplate, "someResource")
	require.Contains(t, mockFS.writtenTemplate, "https://some.other/location/template-operator.yaml")
	require.Contains(t, mockFS.writtenTemplate, "rawManifest")
	require.Contains(t, mockFS.writtenTemplate,
		"https://github.com/kyma-project/template-operator/releases/download/1.0.1/template-operator.yaml")
	require.Contains(t, mockFS.writtenTemplate, "mandatory: false")
	require.NotContains(t, mockFS.writtenTemplate,
		"\"operator.kyma-project.io/mandatory-module\"")
}

type mockFileSystem struct {
	path, writtenTemplate string
}

func (m *mockFileSystem) WriteFile(path, content string) error {
	m.path = path
	m.writtenTemplate = content
	return nil
}
