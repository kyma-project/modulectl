package image_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/kyma-project/modulectl/internal/service/image"
)

func TestExtractImagesFromManifest_WhenCalledWithDeployment_ReturnsImages(t *testing.T) {
	mockParser := &mockManifestParser{}
	service, _ := image.NewService(mockParser)

	manifests := []*unstructured.Unstructured{
		createDeployment("app", []containerSpec{
			{name: "app", image: "app:v1.0.0"},
			{name: "sidecar", image: "sidecar:v2.0.0"},
		}),
	}
	mockParser.manifests = manifests

	images, err := service.ExtractImagesFromManifest("test.yaml")

	require.NoError(t, err)
	require.Len(t, images, 2)
	require.Contains(t, images, "app:v1.0.0")
	require.Contains(t, images, "sidecar:v2.0.0")
}

func TestExtractImagesFromManifest_WhenCalledWithStatefulSet_ReturnsImages(t *testing.T) {
	mockParser := &mockManifestParser{}
	service, _ := image.NewService(mockParser)

	manifests := []*unstructured.Unstructured{
		createStatefulSet("database", []containerSpec{
			{name: "db", image: "postgres:13"},
		}),
	}
	mockParser.manifests = manifests

	images, err := service.ExtractImagesFromManifest("test.yaml")

	require.NoError(t, err)
	require.Len(t, images, 1)
	require.Contains(t, images, "postgres:13")
}

func TestExtractImagesFromManifest_WhenCalledWithInitContainers_ReturnsImages(t *testing.T) {
	mockParser := &mockManifestParser{}
	service, _ := image.NewService(mockParser)

	manifests := []*unstructured.Unstructured{
		createDeploymentWithInitContainers("app", []containerSpec{
			{name: "main", image: "app:v1.0.0"},
		}, []containerSpec{
			{name: "init", image: "init:v1.0.0"},
		}),
	}
	mockParser.manifests = manifests

	images, err := service.ExtractImagesFromManifest("test.yaml")

	require.NoError(t, err)
	require.Len(t, images, 2)
	require.Contains(t, images, "app:v1.0.0")
	require.Contains(t, images, "init:v1.0.0")
}

func TestExtractImagesFromManifest_WhenCalledWithEnvImages_ReturnsImages(t *testing.T) {
	mockParser := &mockManifestParser{}
	service, _ := image.NewService(mockParser)

	manifests := []*unstructured.Unstructured{
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
	}
	mockParser.manifests = manifests

	images, err := service.ExtractImagesFromManifest("test.yaml")

	require.NoError(t, err)
	require.Len(t, images, 3)
	require.Contains(t, images, "app:v1.0.0")
	require.Contains(t, images, "helper:v1.0.0")
	require.Contains(t, images, "tool:v2.0.0")
}

func TestExtractImagesFromManifest_WhenCalledWithLatestTag_ReturnsError(t *testing.T) {
	mockParser := &mockManifestParser{}
	service, _ := image.NewService(mockParser)

	manifests := []*unstructured.Unstructured{
		createDeployment("app", []containerSpec{
			{name: "app", image: "app:latest"},
			{name: "valid", image: "valid:v1.0.0"},
		}),
	}
	mockParser.manifests = manifests

	images, err := service.ExtractImagesFromManifest("test.yaml")

	require.Error(t, err)
	require.Nil(t, images)
	require.Contains(t, err.Error(), "image tag is disallowed")
	require.Contains(t, err.Error(), "latest")
}

func TestExtractImagesFromManifest_WhenCalledWithMainTag_ReturnsError(t *testing.T) {
	mockParser := &mockManifestParser{}
	service, _ := image.NewService(mockParser)

	manifests := []*unstructured.Unstructured{
		createDeployment("app", []containerSpec{
			{name: "app", image: "app:main"},
			{name: "valid", image: "valid:v1.0.0"},
		}),
	}
	mockParser.manifests = manifests

	images, err := service.ExtractImagesFromManifest("test.yaml")

	require.Error(t, err)
	require.Nil(t, images)
	require.Contains(t, err.Error(), "image tag is disallowed")
	require.Contains(t, err.Error(), "main")
}

func TestExtractImagesFromManifest_WhenCalledWithLatestTagUppercase_ReturnsError(t *testing.T) {
	mockParser := &mockManifestParser{}
	service, _ := image.NewService(mockParser)

	manifests := []*unstructured.Unstructured{
		createDeployment("app", []containerSpec{
			{name: "app", image: "app:Latest"},
			{name: "app2", image: "app2:LATEST"},
			{name: "valid", image: "valid:v1.0.0"},
		}),
	}
	mockParser.manifests = manifests

	images, err := service.ExtractImagesFromManifest("test.yaml")

	require.Error(t, err)
	require.Nil(t, images)
	require.Contains(t, err.Error(), "image tag is disallowed")
	require.Contains(t, err.Error(), "Latest")
}

func TestExtractImagesFromManifest_WhenCalledWithMainTagUppercase_ReturnsError(t *testing.T) {
	mockParser := &mockManifestParser{}
	service, _ := image.NewService(mockParser)

	manifests := []*unstructured.Unstructured{
		createDeployment("app", []containerSpec{
			{name: "app", image: "app:Main"},
			{name: "app2", image: "app2:MAIN"},
			{name: "valid", image: "valid:v1.0.0"},
		}),
	}
	mockParser.manifests = manifests

	images, err := service.ExtractImagesFromManifest("test.yaml")

	require.Error(t, err)
	require.Nil(t, images)
	require.Contains(t, err.Error(), "image tag is disallowed")
	require.Contains(t, err.Error(), "Main")
}

func TestExtractImagesFromManifest_WhenCalledWithInvalidImageFormat_FiltersOutImage(t *testing.T) {
	mockParser := &mockManifestParser{}
	service, _ := image.NewService(mockParser)

	manifests := []*unstructured.Unstructured{
		createDeploymentWithEnvImages([]containerSpec{
			{
				name:  "app",
				image: "app:v1.0.0",
				envVars: []envVar{
					{name: "INVALID_IMAGE", value: "invalid-no-tag"},
					{name: "VALID_IMAGE", value: "valid:v1.0.0"},
					{name: "EMPTY_IMAGE", value: ""},
					{name: "SHORT_IMAGE", value: "ab"},
				},
			},
		}),
	}
	mockParser.manifests = manifests

	images, err := service.ExtractImagesFromManifest("test.yaml")

	require.NoError(t, err)
	require.Len(t, images, 2)
	require.Contains(t, images, "app:v1.0.0")
	require.Contains(t, images, "valid:v1.0.0")
}

func TestExtractImagesFromManifest_WhenCalledWithUnsupportedResourceType_IgnoresResource(t *testing.T) {
	mockParser := &mockManifestParser{}
	service, _ := image.NewService(mockParser)

	manifests := []*unstructured.Unstructured{
		createUnsupportedResource("Service", "my-service"),
		createUnsupportedResource("ConfigMap", "my-config"),
		createUnsupportedResource("Secret", "my-secret"),
		createUnsupportedResource("Pod", "my-pod"),
		createDeployment("app", []containerSpec{
			{name: "app", image: "app:v1.0.0"},
		}),
	}
	mockParser.manifests = manifests

	images, err := service.ExtractImagesFromManifest("test.yaml")

	require.NoError(t, err)
	require.Len(t, images, 1)
	require.Contains(t, images, "app:v1.0.0")
}

func TestExtractImagesFromManifest_WhenCalledWithMultipleManifests_ReturnsAllImages(t *testing.T) {
	mockParser := &mockManifestParser{}
	service, _ := image.NewService(mockParser)

	manifests := []*unstructured.Unstructured{
		createDeployment("app1", []containerSpec{
			{name: "app1", image: "app1:v1.0.0"},
		}),
		createStatefulSet("app2", []containerSpec{
			{name: "app2", image: "app2:v2.0.0"},
		}),
		createDeploymentWithInitContainers("app3", []containerSpec{
			{name: "app3", image: "app3:v3.0.0"},
		}, []containerSpec{
			{name: "init3", image: "init3:v1.0.0"},
		}),
	}
	mockParser.manifests = manifests

	images, err := service.ExtractImagesFromManifest("test.yaml")

	require.NoError(t, err)
	require.Len(t, images, 4)
	require.Contains(t, images, "app1:v1.0.0")
	require.Contains(t, images, "app2:v2.0.0")
	require.Contains(t, images, "app3:v3.0.0")
	require.Contains(t, images, "init3:v1.0.0")
}

func TestExtractImagesFromManifest_WhenCalledWithDuplicateImages_DeduplicatesImages(t *testing.T) {
	mockParser := &mockManifestParser{}
	service, _ := image.NewService(mockParser)

	manifests := []*unstructured.Unstructured{
		createDeployment("app1", []containerSpec{
			{name: "app1", image: "shared:v1.0.0"},
		}),
		createDeployment("app2", []containerSpec{
			{name: "app2", image: "shared:v1.0.0"},
		}),
	}
	mockParser.manifests = manifests

	images, err := service.ExtractImagesFromManifest("test.yaml")

	require.NoError(t, err)
	require.Len(t, images, 1)
	require.Contains(t, images, "shared:v1.0.0")
}

func TestExtractImagesFromManifest_WhenParserFails_ReturnsError(t *testing.T) {
	mockParser := &mockManifestParser{
		err: errors.New("parser error"),
	}
	service, _ := image.NewService(mockParser)

	images, err := service.ExtractImagesFromManifest("test.yaml")

	require.Error(t, err)
	require.Nil(t, images)
	require.Contains(t, err.Error(), "failed to parse manifest")
	require.Contains(t, err.Error(), "parser error")
}

func TestExtractImagesFromManifest_WhenCalledWithEmptyContainers_ReturnsEmptyList(t *testing.T) {
	mockParser := &mockManifestParser{}
	service, _ := image.NewService(mockParser)

	manifests := []*unstructured.Unstructured{
		createDeployment("app", []containerSpec{}),
		createStatefulSet("db", []containerSpec{}),
	}
	mockParser.manifests = manifests

	images, err := service.ExtractImagesFromManifest("test.yaml")

	require.NoError(t, err)
	require.Empty(t, images)
}

func TestExtractImagesFromManifest_WhenCalledWithEmptyManifestList_ReturnsEmptyList(t *testing.T) {
	mockParser := &mockManifestParser{}
	service, _ := image.NewService(mockParser)

	mockParser.manifests = []*unstructured.Unstructured{}

	images, err := service.ExtractImagesFromManifest("test.yaml")

	require.NoError(t, err)
	require.Empty(t, images)
}

func TestExtractImagesFromManifest_WhenCalledWithMissingContainerFields_HandlesGracefully(t *testing.T) {
	mockParser := &mockManifestParser{}
	service, _ := image.NewService(mockParser)

	manifest := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"kind": "Deployment",
			"metadata": map[string]interface{}{
				"name": "test",
			},
			"spec": map[string]interface{}{
				"template": map[string]interface{}{
					"spec": map[string]interface{}{
						// Missing containers field
					},
				},
			},
		},
	}

	mockParser.manifests = []*unstructured.Unstructured{manifest}

	images, err := service.ExtractImagesFromManifest("test.yaml")

	require.NoError(t, err)
	require.Empty(t, images)
}

func TestExtractImagesFromManifest_WhenCalledWithMalformedContainerData_HandlesGracefully(t *testing.T) {
	mockParser := &mockManifestParser{}
	service, _ := image.NewService(mockParser)

	manifest := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"kind": "Deployment",
			"metadata": map[string]interface{}{
				"name": "test",
			},
			"spec": map[string]interface{}{
				"template": map[string]interface{}{
					"spec": map[string]interface{}{
						"containers": []interface{}{
							"invalid-container-string",
							map[string]interface{}{
								"name":  "valid",
								"image": "valid:v1.0.0",
							},
						},
					},
				},
			},
		},
	}

	mockParser.manifests = []*unstructured.Unstructured{manifest}

	images, err := service.ExtractImagesFromManifest("test.yaml")

	require.NoError(t, err)
	require.Len(t, images, 1)
	require.Contains(t, images, "valid:v1.0.0")
}

func TestExtractImagesFromManifest_WhenCalledWithDigestImages_ReturnsImages(t *testing.T) {
	mockParser := &mockManifestParser{}
	service, _ := image.NewService(mockParser)

	manifests := []*unstructured.Unstructured{
		createDeployment("app", []containerSpec{
			{name: "app", image: "alpine@sha256:abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234"},
		}),
	}
	mockParser.manifests = manifests

	images, err := service.ExtractImagesFromManifest("test.yaml")

	require.NoError(t, err)
	require.Len(t, images, 1)
	require.Contains(t, images, "alpine@sha256:abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234")
}

func TestExtractImagesFromManifest_WhenCalledWithEmptyTag_ReturnsError(t *testing.T) {
	mockParser := &mockManifestParser{}
	service, _ := image.NewService(mockParser)

	manifests := []*unstructured.Unstructured{
		createDeploymentWithEnvImages([]containerSpec{
			{
				name:  "app",
				image: "app:v1.0.0",
				envVars: []envVar{
					{name: "EMPTY_TAG_IMAGE", value: "image-no-tag-colon:"},
					{name: "VALID_IMAGE", value: "valid:v1.0.0"},
				},
			},
		}),
	}
	mockParser.manifests = manifests

	images, err := service.ExtractImagesFromManifest("test.yaml")

	require.Error(t, err)
	require.Nil(t, images)
	require.Contains(t, err.Error(), "invalid image reference")
	require.Contains(t, err.Error(), "image-no-tag-colon:")
}

func TestExtractImagesFromManifest_WhenCalledWithWhitespaceInImage_FiltersOutImage(t *testing.T) {
	mockParser := &mockManifestParser{}
	service, _ := image.NewService(mockParser)

	manifests := []*unstructured.Unstructured{
		createDeploymentWithEnvImages([]containerSpec{
			{
				name:  "app",
				image: "app:v1.0.0",
				envVars: []envVar{
					{name: "SPACE_IMAGE", value: "image:tag with space"},
					{name: "TAB_IMAGE", value: "image:tag\twith\ttab"},
					{name: "NEWLINE_IMAGE", value: "image:tag\nwith\nnewline"},
					{name: "VALID_IMAGE", value: "valid:v1.0.0"},
				},
			},
		}),
	}
	mockParser.manifests = manifests

	images, err := service.ExtractImagesFromManifest("test.yaml")

	require.NoError(t, err)
	require.Len(t, images, 2)
	require.Contains(t, images, "app:v1.0.0")
	require.Contains(t, images, "valid:v1.0.0")
}

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
			"name":  container.name,
			"image": container.image,
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
