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
	defaultData := []byte(`apiVersion: operator.kyma-project.io/v1alpha1
kind: Sample
metadata:
  name: sample-yaml
spec:
  resourceFilePath: "./module-data.yaml"
`)

	tests := []struct {
		name         string
		data         []byte
		moduleConfig *contentprovider.ModuleConfig
		assertions   func(*testing.T, *mockFileSystem)
	}{
		{
			name: "With Resources",
			data: defaultData,
			moduleConfig: &contentprovider.ModuleConfig{
				Name:        "example.com/component",
				Namespace:   "default",
				Version:     "1.0.0",
				Labels:      map[string]string{"key": "value"},
				Annotations: map[string]string{"annotation": "value"},
				Mandatory:   true,
				Manifest:    "https://github.com/kyma-project/template-operator/releases/download/1.0.1/template-operator.yaml",
				Resources:   contentprovider.Resources{"someResource": "https://some.other/location/template-operator.yaml"},
			},
			assertions: func(t *testing.T, mockFS *mockFileSystem) {
				t.Helper()
				assertCommonTemplateProperties(t, mockFS)
				require.Equal(t, "output.yaml", mockFS.path)

				require.Contains(t, mockFS.writtenTemplate,
					"https://github.com/kyma-project/template-operator/releases/download/1.0.1/template-operator.yaml")
				require.Contains(t, mockFS.writtenTemplate, "someResource")
				require.Contains(t, mockFS.writtenTemplate, "https://some.other/location/template-operator.yaml")
			},
		},
		{
			name: "With Default CR Starting With Dashes",
			data: []byte(`---
apiVersion: operator.kyma-project.io/v1alpha1
kind: Sample
metadata:
  name: sample-yaml
spec:
   resourceFilePath: "./module-data.yaml"
`),
			moduleConfig: &contentprovider.ModuleConfig{
				Name:        "example.com/component",
				Namespace:   "default",
				Version:     "1.0.0",
				Labels:      map[string]string{"key": "value"},
				Annotations: map[string]string{"annotation": "value"},
				Mandatory:   true,
				Manifest:    "https://github.com/kyma-project/template-operator/releases/download/1.0.1/template-operator.yaml",
			},
			assertions: func(t *testing.T, mockFS *mockFileSystem) {
				t.Helper()
				assertCommonTemplateProperties(t, mockFS)
			},
		},
		{
			name: "With Overwritten Raw Manifest",
			data: defaultData,
			moduleConfig: &contentprovider.ModuleConfig{
				Name:      "example.com/component",
				Manifest:  "https://github.com/kyma-project/template-operator/releases/download/1.0.1/template-operator.yaml",
				Resources: contentprovider.Resources{"rawManifest": "https://some.other/location/template-operator.yaml"},
			},
			assertions: func(t *testing.T, mockFS *mockFileSystem) {
				t.Helper()
				require.Contains(t, mockFS.writtenTemplate, "https://some.other/location/template-operator.yaml")
				require.NotContains(t, mockFS.writtenTemplate,
					"https://github.com/kyma-project/template-operator/releases/download/1.0.1/template-operator.yaml")
			},
		},
		{
			name: "With Associated Resources",
			data: defaultData,
			moduleConfig: &contentprovider.ModuleConfig{
				Name:        "example.com/component",
				Namespace:   "default",
				Version:     "1.0.0",
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
			},
			assertions: func(t *testing.T, mockFS *mockFileSystem) {
				t.Helper()
				assertCommonTemplateProperties(t, mockFS)
				require.Contains(t, mockFS.writtenTemplate, "associatedResources")
				require.Contains(t, mockFS.writtenTemplate, "networking.istio.io")
				require.Contains(t, mockFS.writtenTemplate, "v1alpha3")
				require.Contains(t, mockFS.writtenTemplate, "Gateway")
			},
		},
		{
			name: "With Manager",
			data: defaultData,
			moduleConfig: &contentprovider.ModuleConfig{
				Name:        "example.com/component",
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
			},
			assertions: func(t *testing.T, mockFS *mockFileSystem) {
				t.Helper()
				assertCommonTemplateProperties(t, mockFS)
				require.Contains(t, mockFS.writtenTemplate, "manager-name")
				require.Contains(t, mockFS.writtenTemplate, "manager-ns")
				require.Contains(t, mockFS.writtenTemplate, "apps")
				require.Contains(t, mockFS.writtenTemplate, "v1")
				require.Contains(t, mockFS.writtenTemplate, "Deployment")
				require.Equal(t, 2, strings.Count(mockFS.writtenTemplate, "namespace"))
			},
		},
		{
			name: "With Manager Without Namespace",
			data: defaultData,
			moduleConfig: &contentprovider.ModuleConfig{
				Name:        "example.com/component",
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
			},
			assertions: func(t *testing.T, mockFS *mockFileSystem) {
				t.Helper()
				assertCommonTemplateProperties(t, mockFS)
				require.Contains(t, mockFS.writtenTemplate, "manager-name")
				require.Contains(t, mockFS.writtenTemplate, "apps")
				require.Contains(t, mockFS.writtenTemplate, "v1")
				require.Contains(t, mockFS.writtenTemplate, "Deployment")
				require.Equal(t, 1, strings.Count(mockFS.writtenTemplate, "namespace"))
			},
		},
		{
			name: "With Mandatory False",
			data: defaultData,
			moduleConfig: &contentprovider.ModuleConfig{
				Name:        "example.com/component",
				Namespace:   "default",
				Version:     "1.0.0",
				Labels:      map[string]string{"key": "value"},
				Annotations: map[string]string{"annotation": "value"},
				Mandatory:   false,
				Manifest:    "https://github.com/kyma-project/template-operator/releases/download/1.0.1/template-operator.yaml",
				Resources:   contentprovider.Resources{"someResource": "https://some.other/location/template-operator.yaml"},
			},
			assertions: func(t *testing.T, mockFS *mockFileSystem) {
				t.Helper()
				assertCommonTemplateProperties(t, mockFS)
				require.Contains(t, mockFS.writtenTemplate, "mandatory: false")
				require.NotContains(t, mockFS.writtenTemplate, "\"operator.kyma-project.io/mandatory-module\"")
			},
		},
		{
			name: "With Mandatory True",
			data: defaultData,
			moduleConfig: &contentprovider.ModuleConfig{
				Name:        "example.com/component",
				Namespace:   "default",
				Version:     "1.0.0",
				Labels:      map[string]string{"key": "value"},
				Annotations: map[string]string{"annotation": "value"},
				Mandatory:   true,
				Manifest:    "https://github.com/kyma-project/template-operator/releases/download/1.0.1/template-operator.yaml",
				Resources:   contentprovider.Resources{"someResource": "https://some.other/location/template-operator.yaml"},
			},
			assertions: func(t *testing.T, mockFS *mockFileSystem) {
				t.Helper()
				assertCommonTemplateProperties(t, mockFS)
				require.Contains(t, mockFS.writtenTemplate, "mandatory: true")
				require.Contains(t, mockFS.writtenTemplate,
					"\"operator.kyma-project.io/mandatory-module\": \"true\"")
			},
		},
		{
			name: "With Requires Downtime True",
			data: defaultData,
			moduleConfig: &contentprovider.ModuleConfig{
				Name:             "example.com/component",
				Namespace:        "default",
				Version:          "1.0.0",
				Labels:           map[string]string{"key": "value"},
				Annotations:      map[string]string{"annotation": "value"},
				RequiresDowntime: true,
				Manifest:         "https://github.com/kyma-project/template-operator/releases/download/1.0.1/template-operator.yaml",
				Resources:        contentprovider.Resources{"someResource": "https://some.other/location/template-operator.yaml"},
			},
			assertions: func(t *testing.T, mockFS *mockFileSystem) {
				t.Helper()
				assertCommonTemplateProperties(t, mockFS)
				require.Contains(t, mockFS.writtenTemplate, "requiresDowntime: true")
			},
		},
		{
			name: "With Requires Downtime False",
			data: defaultData,
			moduleConfig: &contentprovider.ModuleConfig{
				Name:             "example.com/component",
				Namespace:        "default",
				Version:          "1.0.0",
				Labels:           map[string]string{"key": "value"},
				Annotations:      map[string]string{"annotation": "value"},
				RequiresDowntime: false,
				Manifest:         "https://github.com/kyma-project/template-operator/releases/download/1.0.1/template-operator.yaml",
				Resources:        contentprovider.Resources{"someResource": "https://some.other/location/template-operator.yaml"},
			},
			assertions: func(t *testing.T, mockFS *mockFileSystem) {
				t.Helper()
				assertCommonTemplateProperties(t, mockFS)
				require.Contains(t, mockFS.writtenTemplate, "requiresDowntime: false")
			},
		},
		{
			name: "With No Default CR",
			data: nil,
			moduleConfig: &contentprovider.ModuleConfig{
				Name:        "example.com/component",
				Namespace:   "default",
				Version:     "1.0.0",
				Labels:      map[string]string{"key": "value"},
				Annotations: map[string]string{"annotation": "value"},
				Manifest:    "https://github.com/kyma-project/template-operator/releases/download/1.0.1/template-operator.yaml",
				Resources:   contentprovider.Resources{"someResource": "https://some.other/location/template-operator.yaml"},
			},
			assertions: func(t *testing.T, mockFS *mockFileSystem) {
				t.Helper()
				require.NotContains(t, mockFS.writtenTemplate, "kind: Sample")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFS := &mockFileSystem{}
			svc, _ := templategenerator.NewService(mockFS)
			descriptor := testutils.CreateComponentDescriptor("example.com/component", "1.0.0")

			err := svc.GenerateModuleTemplate(tt.moduleConfig, descriptor, tt.data, true, "output.yaml")

			require.NoError(t, err)
			require.Equal(t, "output.yaml", mockFS.path)
			tt.assertions(t, mockFS)
		})
	}
}

type mockFileSystem struct {
	path, writtenTemplate string
}

func (m *mockFileSystem) WriteFile(path, content string) error {
	m.path = path
	m.writtenTemplate = content
	return nil
}

func assertCommonTemplateProperties(t *testing.T, mockFS *mockFileSystem) {
	t.Helper()
	require.Contains(t, mockFS.writtenTemplate, "version: 1.0.0")
	require.Contains(t, mockFS.writtenTemplate, "moduleName: component")
	require.Contains(t, mockFS.writtenTemplate, "component-1.0.0")
	require.Contains(t, mockFS.writtenTemplate, "default")
	require.Contains(t, mockFS.writtenTemplate, "example.com/component")
	require.Contains(t, mockFS.writtenTemplate, "rawManifest")
	require.NotContains(t, mockFS.writtenTemplate, "---")
	require.Contains(t, mockFS.writtenTemplate, "apiVersion: operator.kyma-project.io/v1alpha1")
	require.Contains(t, mockFS.writtenTemplate, "kind: Sample")
	require.Contains(t, mockFS.writtenTemplate, "metadata:")
	require.Contains(t, mockFS.writtenTemplate, "name: sample-yaml")
	require.Contains(t, mockFS.writtenTemplate, "spec:")
	require.Contains(t, mockFS.writtenTemplate, "resourceFilePath: ./module-data.yaml")
}
