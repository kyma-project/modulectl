package contentprovider_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/kyma-project/modulectl/internal/common/types"
	"github.com/kyma-project/modulectl/internal/service/contentprovider"
)

func Test_Manifest_GetDefaultContent_ReturnsExpectedValue(t *testing.T) {
	mockParser := &mockManifestParser{}
	manifestContentProvider, err := contentprovider.NewManifest(mockParser)
	require.NoError(t, err)

	expectedDefault := `# This file holds the Manifest of your module, encompassing all resources installed in the cluster once the module is activated.
# It should include the Custom Resource Definition for your module'm default CustomResource, if it exists.

`
	manifestGeneratedDefaultContentWithNil, _ := manifestContentProvider.GetDefaultContent(nil)
	manifestGeneratedDefaultContentWithEmptyMap, _ := manifestContentProvider.GetDefaultContent(make(types.KeyValueArgs))

	t.Parallel()
	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{
			name:     "Manifest Default Content with Nil",
			value:    manifestGeneratedDefaultContentWithNil,
			expected: expectedDefault,
		}, {
			name:     "Manifest Default Content with Empty Map",
			value:    manifestGeneratedDefaultContentWithEmptyMap,
			expected: expectedDefault,
		},
	}

	for _, testcase := range tests {
		testName := "TestCorrectContentProviderFor_" + testcase.name
		t.Run(testName, func(t *testing.T) {
			t.Parallel()
			if testcase.value != testcase.expected {
				t.Errorf("ContentProvider for '%s' did not return correct default: expected = '%s', but got = '%s'",
					testcase.name, testcase.expected, testcase.value)
			}
		})
	}
}

func TestExtractImagesFromManifest_Deployment(t *testing.T) {
	mockParser := &mockManifestParser{
		manifests: []*unstructured.Unstructured{
			createDeployment("app", []containerSpec{
				{name: "app", image: "app:v1.0.0"},
				{name: "sidecar", image: "sidecar:v2.0.0"},
			}),
		},
	}
	manifest, _ := contentprovider.NewManifest(mockParser)

	images, err := manifest.ExtractImagesFromManifest("test.yaml")
	require.NoError(t, err)
	require.Len(t, images, 2)
	require.Contains(t, images, "app:v1.0.0")
	require.Contains(t, images, "sidecar:v2.0.0")
}

func TestExtractImagesFromManifest_StatefulSet(t *testing.T) {
	mockParser := &mockManifestParser{
		manifests: []*unstructured.Unstructured{
			createStatefulSet("db", []containerSpec{
				{name: "db", image: "postgres:13"},
			}),
		},
	}
	manifest, _ := contentprovider.NewManifest(mockParser)

	images, err := manifest.ExtractImagesFromManifest("test.yaml")
	require.NoError(t, err)
	require.Len(t, images, 1)
	require.Contains(t, images, "postgres:13")
}

func TestExtractImagesFromManifest_InitContainers(t *testing.T) {
	mockParser := &mockManifestParser{
		manifests: []*unstructured.Unstructured{
			createDeploymentWithInitContainers("app", []containerSpec{
				{name: "main", image: "app:v1.0.0"},
			}, []containerSpec{
				{name: "init", image: "init:v1.0.0"},
			}),
		},
	}
	manifest, _ := contentprovider.NewManifest(mockParser)

	images, err := manifest.ExtractImagesFromManifest("test.yaml")
	require.NoError(t, err)
	require.Len(t, images, 2)
	require.Contains(t, images, "app:v1.0.0")
	require.Contains(t, images, "init:v1.0.0")
}

func TestExtractImagesFromManifest_EnvImages(t *testing.T) {
	mockParser := &mockManifestParser{
		manifests: []*unstructured.Unstructured{
			createDeploymentWithEnvImages([]containerSpec{
				{
					name:  "app",
					image: "app:v1.0.0",
					envVars: []envVar{
						{name: "HELPER_IMAGE", value: "helper:v1.0.0"},
						{name: "TOOL_IMAGE", value: "tool:v2.0.0"},
					},
				},
			}),
		},
	}
	manifest, _ := contentprovider.NewManifest(mockParser)

	images, err := manifest.ExtractImagesFromManifest("test.yaml")
	require.NoError(t, err)
	require.Len(t, images, 3)
	require.Contains(t, images, "app:v1.0.0")
	require.Contains(t, images, "helper:v1.0.0")
	require.Contains(t, images, "tool:v2.0.0")
}

func TestExtractImagesFromManifest_DisallowedTag(t *testing.T) {
	mockParser := &mockManifestParser{
		manifests: []*unstructured.Unstructured{
			createDeployment("app", []containerSpec{
				{name: "app", image: "app:latest"},
			}),
		},
	}
	manifest, _ := contentprovider.NewManifest(mockParser)

	images, err := manifest.ExtractImagesFromManifest("test.yaml")
	require.Error(t, err)
	require.Nil(t, images)
	require.Contains(t, err.Error(), "image tag is disallowed")
}

func TestExtractImagesFromManifest_DuplicateImages(t *testing.T) {
	mockParser := &mockManifestParser{
		manifests: []*unstructured.Unstructured{
			createDeployment("app1", []containerSpec{
				{name: "app1", image: "shared:v1.0.0"},
			}),
			createDeployment("app2", []containerSpec{
				{name: "app2", image: "shared:v1.0.0"},
			}),
		},
	}
	manifest, _ := contentprovider.NewManifest(mockParser)

	images, err := manifest.ExtractImagesFromManifest("test.yaml")
	require.NoError(t, err)
	require.Len(t, images, 1)
	require.Contains(t, images, "shared:v1.0.0")
}

func TestExtractImagesFromManifest_ParserError(t *testing.T) {
	mockParser := &mockManifestParser{
		err: errors.New("parser error"),
	}
	manifest, _ := contentprovider.NewManifest(mockParser)

	images, err := manifest.ExtractImagesFromManifest("test.yaml")
	require.Error(t, err)
	require.Nil(t, images)
	require.Contains(t, err.Error(), "failed to parse manifest")
}

func TestExtractImagesFromManifest_EmptyContainers(t *testing.T) {
	mockParser := &mockManifestParser{
		manifests: []*unstructured.Unstructured{
			createDeployment("app", []containerSpec{}),
			createStatefulSet("db", []containerSpec{}),
		},
	}
	manifest, _ := contentprovider.NewManifest(mockParser)

	images, err := manifest.ExtractImagesFromManifest("test.yaml")
	require.NoError(t, err)
	require.Empty(t, images)
}

func TestExtractImagesFromManifest_UnsupportedResourceType_IgnoresResource(t *testing.T) {
	mockParser := &mockManifestParser{
		manifests: []*unstructured.Unstructured{
			createUnsupportedResource("Service", "my-service"),
			createUnsupportedResource("ConfigMap", "my-config"),
			createDeployment("app", []containerSpec{
				{name: "app", image: "app:v1.0.0"},
			}),
		},
	}
	manifest, _ := contentprovider.NewManifest(mockParser)

	images, err := manifest.ExtractImagesFromManifest("test.yaml")
	require.NoError(t, err)
	require.Len(t, images, 1)
	require.Contains(t, images, "app:v1.0.0")
}

type mockManifestParser struct {
	manifests []*unstructured.Unstructured
	err       error
}

func (m *mockManifestParser) Parse(_ string) ([]*unstructured.Unstructured, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.manifests, nil
}

// --- Test helpers (copy from image_extraction_test.go) ---

type containerSpec struct {
	name    string
	image   string
	envVars []envVar
}

type envVar struct {
	name  string
	value string
}

func createDeployment(name string, containers []containerSpec) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"kind": "Deployment",
			"metadata": map[string]interface{}{
				"name": name,
			},
			"spec": map[string]interface{}{
				"template": map[string]interface{}{
					"spec": map[string]interface{}{
						"containers": createContainers(containers),
					},
				},
			},
		},
	}
}

func createStatefulSet(name string, containers []containerSpec) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"kind": "StatefulSet",
			"metadata": map[string]interface{}{
				"name": name,
			},
			"spec": map[string]interface{}{
				"template": map[string]interface{}{
					"spec": map[string]interface{}{
						"containers": createContainers(containers),
					},
				},
			},
		},
	}
}

func createDeploymentWithInitContainers(name string, containers []containerSpec, initContainers []containerSpec) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"kind": "Deployment",
			"metadata": map[string]interface{}{
				"name": name,
			},
			"spec": map[string]interface{}{
				"template": map[string]interface{}{
					"spec": map[string]interface{}{
						"containers":     createContainers(containers),
						"initContainers": createContainers(initContainers),
					},
				},
			},
		},
	}
}

func createDeploymentWithEnvImages(containers []containerSpec) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"kind": "Deployment",
			"metadata": map[string]interface{}{
				"name": "app",
			},
			"spec": map[string]interface{}{
				"template": map[string]interface{}{
					"spec": map[string]interface{}{
						"containers": createContainers(containers),
					},
				},
			},
		},
	}
}

func createContainers(containers []containerSpec) []interface{} {
	result := make([]interface{}, 0, len(containers))
	for _, container := range containers {
		containerObj := map[string]interface{}{
			"name": container.name,
			"img":  container.image,
		}
		if len(container.envVars) > 0 {
			envVars := make([]interface{}, 0, len(container.envVars))
			for _, env := range container.envVars {
				envVars = append(envVars, map[string]interface{}{
					"name":  env.name,
					"value": env.value,
				})
			}
			containerObj["env"] = envVars
		}
		result = append(result, containerObj)
	}
	return result
}

func createUnsupportedResource(kind, name string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"kind": kind,
			"metadata": map[string]interface{}{
				"name": name,
			},
		},
	}
}
