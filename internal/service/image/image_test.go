package image_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"ocm.software/ocm/api/ocm/compdesc"
	ocmv1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"

	"github.com/kyma-project/modulectl/internal/service/image"
)

func TestNewService_WhenCalledWithValidParser_ReturnsValidService(t *testing.T) {
	mockParser := &mockManifestParser{}
	service, err := image.NewService(mockParser)

	require.NoError(t, err)
	require.NotNil(t, service)
}

func TestNewService_WhenCalledWithNilParser_ReturnsError(t *testing.T) {
	service, err := image.NewService(nil)

	require.Error(t, err)
	require.Nil(t, service)
	require.Contains(t, err.Error(), "manifestParser must not be nil")
}

func TestExtractImagesFromManifest_WhenParserReturnsError_ReturnsError(t *testing.T) {
	mockParser := &mockManifestParser{}
	service, _ := image.NewService(mockParser)

	expectedError := errors.New("parser error")
	mockParser.On("Parse", "test.yaml").Return(nil, expectedError)

	result, err := service.ExtractImagesFromManifest("test.yaml")

	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "failed to parse manifest")
	mockParser.AssertExpectations(t)
}

func TestExtractImagesFromManifest_WhenParserSucceeds_ReturnsImages(t *testing.T) {
	mockParser := &mockManifestParser{}
	service, _ := image.NewService(mockParser)

	expectedManifests := []*unstructured.Unstructured{
		createDeployment("telemetry-manager", []containerSpec{
			{name: "manager", image: "europe-docker.pkg.dev/kyma-project/prod/telemetry-manager:v1.2.0"},
		}),
	}
	mockParser.manifests = expectedManifests

	images, err := service.ExtractImagesFromManifest("test.yaml")

	require.NoError(t, err)
	require.Len(t, images, 1)
	require.Contains(t, images, "europe-docker.pkg.dev/kyma-project/prod/telemetry-manager:v1.2.0")
}

func TestExtractImagesFromManifest_WhenCalledWithEmptyManifests_ReturnsEmptySlice(t *testing.T) {
	mockParser := &mockManifestParser{}
	service, _ := image.NewService(mockParser)
	mockParser.manifests = []*unstructured.Unstructured{}

	images, err := service.ExtractImagesFromManifest("test.yaml")

	require.NoError(t, err)
	require.Empty(t, images)
}

func TestExtractImagesFromManifest_WhenCalledWithDeployment_ReturnsImages(t *testing.T) {
	mockParser := &mockManifestParser{}
	service, _ := image.NewService(mockParser)

	manifests := []*unstructured.Unstructured{
		createDeployment("telemetry-manager", []containerSpec{
			{name: "manager", image: "europe-docker.pkg.dev/kyma-project/prod/telemetry-manager:v1.2.0"},
			{name: "fluent-bit", image: "europe-docker.pkg.dev/kyma-project/prod/fluent-bit:v2.1.8"},
		}),
	}
	mockParser.manifests = manifests

	images, err := service.ExtractImagesFromManifest("test.yaml")

	require.NoError(t, err)
	require.Len(t, images, 2)
	require.Contains(t, images, "europe-docker.pkg.dev/kyma-project/prod/telemetry-manager:v1.2.0")
	require.Contains(t, images, "europe-docker.pkg.dev/kyma-project/prod/fluent-bit:v2.1.8")
}

func TestExtractImagesFromManifest_WhenCalledWithStatefulSet_ReturnsImages(t *testing.T) {
	mockParser := &mockManifestParser{}
	service, _ := image.NewService(mockParser)

	manifests := []*unstructured.Unstructured{
		createStatefulSet("istio-proxy", []containerSpec{
			{name: "proxy", image: "gcr.io/istio-release/proxyv2:1.19.0"},
		}),
	}
	mockParser.manifests = manifests

	images, err := service.ExtractImagesFromManifest("test.yaml")

	require.NoError(t, err)
	require.Len(t, images, 1)
	require.Contains(t, images, "gcr.io/istio-release/proxyv2:1.19.0")
}

func TestExtractImagesFromManifest_WhenCalledWithUnsupportedWorkload_ReturnsEmptySlice(t *testing.T) {
	mockParser := &mockManifestParser{}
	service, _ := image.NewService(mockParser)

	manifests := []*unstructured.Unstructured{
		{
			Object: map[string]interface{}{
				"kind": "Service",
				"metadata": map[string]interface{}{
					"name": "telemetry-service",
				},
			},
		},
	}
	mockParser.manifests = manifests

	images, err := service.ExtractImagesFromManifest("test.yaml")

	require.NoError(t, err)
	require.Empty(t, images)
}

func TestExtractImagesFromManifest_WhenCalledWithEnvironmentImages_ReturnsImages(t *testing.T) {
	mockParser := &mockManifestParser{}
	service, _ := image.NewService(mockParser)

	manifests := []*unstructured.Unstructured{
		createDeploymentWithEnvImages("telemetry-manager", []containerSpec{
			{
				name:  "manager",
				image: "europe-docker.pkg.dev/kyma-project/prod/telemetry-manager:v1.2.0",
				envVars: []envVar{
					{name: "WEBHOOK_IMAGE", value: "europe-docker.pkg.dev/kyma-project/prod/telemetry-webhook:v1.0.0"},
				},
			},
		}),
	}
	mockParser.manifests = manifests

	images, err := service.ExtractImagesFromManifest("test.yaml")

	require.NoError(t, err)
	require.Len(t, images, 2)
	require.Contains(t, images, "europe-docker.pkg.dev/kyma-project/prod/telemetry-manager:v1.2.0")
	require.Contains(t, images, "europe-docker.pkg.dev/kyma-project/prod/telemetry-webhook:v1.0.0")
}

func TestExtractImagesFromManifest_WhenCalledWithDuplicateImages_ReturnsDeduplicatedImages(t *testing.T) {
	mockParser := &mockManifestParser{}
	service, _ := image.NewService(mockParser)

	manifests := []*unstructured.Unstructured{
		createDeployment("telemetry-manager-1", []containerSpec{
			{name: "manager", image: "europe-docker.pkg.dev/kyma-project/prod/telemetry-manager:v1.2.0"},
		}),
		createDeployment("telemetry-manager-2", []containerSpec{
			{name: "manager", image: "europe-docker.pkg.dev/kyma-project/prod/telemetry-manager:v1.2.0"},
		}),
	}
	mockParser.manifests = manifests

	images, err := service.ExtractImagesFromManifest("test.yaml")

	require.NoError(t, err)
	require.Len(t, images, 1)
	require.Contains(t, images, "europe-docker.pkg.dev/kyma-project/prod/telemetry-manager:v1.2.0")
}

func TestExtractImagesFromManifest_WhenCalledWithMultipleWorkloadTypes_ReturnsAllImages(t *testing.T) {
	mockParser := &mockManifestParser{}
	service, _ := image.NewService(mockParser)

	manifests := []*unstructured.Unstructured{
		createDeployment("telemetry-manager", []containerSpec{
			{name: "manager", image: "europe-docker.pkg.dev/kyma-project/prod/telemetry-manager:v1.2.0"},
		}),
		createStatefulSet("istio-proxy", []containerSpec{
			{name: "proxy", image: "gcr.io/istio-release/proxyv2:1.19.0"},
		}),
		{
			Object: map[string]interface{}{
				"kind": "Service",
				"metadata": map[string]interface{}{
					"name": "telemetry-service",
				},
			},
		},
	}
	mockParser.manifests = manifests

	images, err := service.ExtractImagesFromManifest("test.yaml")

	require.NoError(t, err)
	require.Len(t, images, 2)
	require.Contains(t, images, "europe-docker.pkg.dev/kyma-project/prod/telemetry-manager:v1.2.0")
	require.Contains(t, images, "gcr.io/istio-release/proxyv2:1.19.0")
}

func TestAddImagesToOcmDescriptor_WhenCalledWithEmptyImages_ReturnsNoError(t *testing.T) {
	mockParser := &mockManifestParser{}
	service, _ := image.NewService(mockParser)

	descriptor := &compdesc.ComponentDescriptor{
		ComponentSpec: compdesc.ComponentSpec{
			ObjectMeta: ocmv1.ObjectMeta{
				Name:     "kyma-project.io/module/telemetry",
				Version:  "1.0.0",
				Provider: ocmv1.Provider{Name: "kyma-project.io"},
			},
			Resources: []compdesc.Resource{},
		},
	}

	err := service.AddImagesToOcmDescriptor(descriptor, []string{})

	require.NoError(t, err)
	require.Empty(t, descriptor.Resources)
}

func TestAddImagesToOcmDescriptor_WhenCalledWithValidImages_AppendsResources(t *testing.T) {
	mockParser := &mockManifestParser{}
	service, _ := image.NewService(mockParser)

	descriptor := &compdesc.ComponentDescriptor{
		ComponentSpec: compdesc.ComponentSpec{
			ObjectMeta: ocmv1.ObjectMeta{
				Name:     "kyma-project.io/module/telemetry",
				Version:  "1.0.0",
				Provider: ocmv1.Provider{Name: "kyma-project.io"},
			},
			Resources: []compdesc.Resource{},
		},
	}
	images := []string{
		"europe-docker.pkg.dev/kyma-project/prod/telemetry-manager:v1.2.0",
		"gcr.io/istio-release/proxyv2:1.19.0",
	}

	err := service.AddImagesToOcmDescriptor(descriptor, images)

	require.NoError(t, err)
	require.Len(t, descriptor.Resources, 2)

	// Check first resource
	require.Equal(t, "telemetry-manager", descriptor.Resources[0].Name)
	require.Equal(t, "v1.2.0", descriptor.Resources[0].Version)
	require.Equal(t, "ociArtifact", descriptor.Resources[0].Type)
	require.Len(t, descriptor.Resources[0].Labels, 1)
	require.Equal(t, "scan.security.kyma-project.io/type", descriptor.Resources[0].Labels[0].Name)

	var labelValue string
	err = json.Unmarshal(descriptor.Resources[0].Labels[0].Value, &labelValue)
	require.NoError(t, err)
	require.Equal(t, "manifest-image", labelValue)

	// Check second resource
	require.Equal(t, "proxyv2", descriptor.Resources[1].Name)
	require.Equal(t, "1.19.0", descriptor.Resources[1].Version)
	require.Equal(t, "ociArtifact", descriptor.Resources[1].Type)
}

func TestAddImagesToOcmDescriptor_WhenCalledWithInvalidImageFormat_ReturnsError(t *testing.T) {
	mockParser := &mockManifestParser{}
	service, _ := image.NewService(mockParser)

	descriptor := &compdesc.ComponentDescriptor{
		ComponentSpec: compdesc.ComponentSpec{
			ObjectMeta: ocmv1.ObjectMeta{
				Name:    "kyma-project.io/module/telemetry",
				Version: "1.0.0",
			},
			Resources: []compdesc.Resource{},
		},
	}
	images := []string{"invalid-image-format"}

	err := service.AddImagesToOcmDescriptor(descriptor, images)

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to append image")
}

func TestGetImageNameAndTag(t *testing.T) {
	tests := []struct {
		name              string
		imageURL          string
		expectedImageName string
		expectedImageTag  string
		expectedError     error
	}{
		{
			name:              "valid image URL",
			imageURL:          "docker.io/template-operator/test:latest",
			expectedImageName: "test",
			expectedImageTag:  "latest",
			expectedError:     nil,
		},
		{
			name:              "invalid image URL - no tag",
			imageURL:          "docker.io/template-operator/test",
			expectedImageName: "",
			expectedImageTag:  "",
			expectedError:     errors.New("invalid image URL"),
		},
		{
			name:              "invalid image URL - empty URL",
			imageURL:          "",
			expectedImageName: "",
			expectedImageTag:  "",
			expectedError:     errors.New("invalid image URL"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			imgName, imgTag, err := image.GetImageNameAndTag(test.imageURL)
			if test.expectedError != nil {
				require.ErrorContains(t, err, test.expectedError.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, test.expectedImageName, imgName)
				require.Equal(t, test.expectedImageTag, imgTag)
			}
		})
	}
}

// Mock parser implementation
type mockManifestParser struct {
	manifests     []*unstructured.Unstructured
	err           error
	calledWith    string
	callCount     int
	expectedPath  string
	expectedError error
}

func (m *mockManifestParser) Parse(path string) ([]*unstructured.Unstructured, error) {
	m.calledWith = path
	m.callCount++

	if m.expectedError != nil {
		return nil, m.expectedError
	}

	return m.manifests, m.err
}

func (m *mockManifestParser) On(method, path string) *mockManifestParser {
	m.expectedPath = path
	return m
}

func (m *mockManifestParser) Return(manifests []*unstructured.Unstructured, err error) {
	m.manifests = manifests
	m.expectedError = err
}

func (m *mockManifestParser) AssertExpectations(t *testing.T) {
	t.Helper()
	if m.expectedPath != "" {
		require.Equal(t, m.expectedPath, m.calledWith)
	}
}

// Helper types and functions for creating test data
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

func createDeploymentWithEnvImages(name string, containers []containerSpec) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"kind": "Deployment",
			"metadata": map[string]interface{}{
				"name": name,
			},
			"spec": map[string]interface{}{
				"template": map[string]interface{}{
					"spec": map[string]interface{}{
						"containers": createContainersWithEnv(containers),
					},
				},
			},
		},
	}
}

func createContainers(containers []containerSpec) []interface{} {
	result := make([]interface{}, len(containers))
	for i, container := range containers {
		result[i] = map[string]interface{}{
			"name":  container.name,
			"image": container.image,
		}
	}
	return result
}

func createContainersWithEnv(containers []containerSpec) []interface{} {
	result := make([]interface{}, len(containers))
	for i, container := range containers {
		containerMap := map[string]interface{}{
			"name":  container.name,
			"image": container.image,
		}

		if len(container.envVars) > 0 {
			env := make([]interface{}, len(container.envVars))
			for j, envVar := range container.envVars {
				env[j] = map[string]interface{}{
					"name":  envVar.name,
					"value": envVar.value,
				}
			}
			containerMap["env"] = env
		}

		result[i] = containerMap
	}
	return result
}
