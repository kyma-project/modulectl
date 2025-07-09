package image_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"ocm.software/ocm/api/ocm/compdesc"
	ocmv1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/ociartifact"

	"github.com/kyma-project/modulectl/internal/service/image"
)

// Ensure mock implements the interface at compile time
var _ image.ManifestParser = (*mockManifestParser)(nil)

func TestNewService_ReturnsValidService(t *testing.T) {
	service := image.NewService()
	require.NotNil(t, service)
}

func TestParseManifest_WhenCalledWithNilParser_ReturnsError(t *testing.T) {
	service := image.NewService()

	_, err := service.ManifestParse("test.yaml", nil)

	require.Error(t, err)
	require.Contains(t, err.Error(), "parser cannot be nil")
}

func TestParseManifest_WhenParserReturnsError_ReturnsError(t *testing.T) {
	mockParser := &mockManifestParser{}
	service := image.NewService()

	expectedError := errors.New("parser error")
	mockParser.On("Parse", "test.yaml").Return(nil, expectedError)

	result, err := service.ManifestParse("test.yaml", mockParser)

	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "failed to parse manifest at test.yaml: parser error")
	mockParser.AssertExpectations(t)
}

func TestParseManifest_WhenParserSucceeds_ReturnsManifests(t *testing.T) {
	service := image.NewService()
	expectedManifests := []*unstructured.Unstructured{
		createDeployment("telemetry-manager", []containerSpec{
			{name: "manager", image: "europe-docker.pkg.dev/kyma-project/prod/telemetry-manager:v1.2.0"},
		}),
	}
	mockParser := &mockManifestParser{
		manifests: expectedManifests,
	}

	manifests, err := service.ManifestParse("test.yaml", mockParser)

	require.NoError(t, err)
	require.Equal(t, expectedManifests, manifests)
}

func TestGetAllImages_WhenCalledWithEmptyManifests_ReturnsEmptySlice(t *testing.T) {
	service := image.NewService()

	images := service.GetAllImages([]*unstructured.Unstructured{})

	require.Empty(t, images)
}

func TestGetAllImages_WhenCalledWithDeployment_ReturnsImages(t *testing.T) {
	service := image.NewService()
	manifests := []*unstructured.Unstructured{
		createDeployment("telemetry-manager", []containerSpec{
			{name: "manager", image: "europe-docker.pkg.dev/kyma-project/prod/telemetry-manager:v1.2.0"},
			{name: "fluent-bit", image: "europe-docker.pkg.dev/kyma-project/prod/fluent-bit:v2.1.8"},
		}),
	}

	images := service.GetAllImages(manifests)

	require.Len(t, images, 2)
	require.Contains(t, images, "europe-docker.pkg.dev/kyma-project/prod/telemetry-manager:v1.2.0")
	require.Contains(t, images, "europe-docker.pkg.dev/kyma-project/prod/fluent-bit:v2.1.8")
}

func TestGetAllImages_WhenCalledWithStatefulSet_ReturnsImages(t *testing.T) {
	service := image.NewService()
	manifests := []*unstructured.Unstructured{
		createStatefulSet("istio-proxy", []containerSpec{
			{name: "proxy", image: "gcr.io/istio-release/proxyv2:1.19.0"},
		}),
	}

	images := service.GetAllImages(manifests)

	require.Len(t, images, 1)
	require.Contains(t, images, "gcr.io/istio-release/proxyv2:1.19.0")
}

func TestGetAllImages_WhenCalledWithUnsupportedWorkload_ReturnsEmptySlice(t *testing.T) {
	service := image.NewService()
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

	images := service.GetAllImages(manifests)

	require.Empty(t, images)
}

func TestGetAllImages_WhenCalledWithEnvironmentImages_ReturnsImages(t *testing.T) {
	service := image.NewService()
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

	images := service.GetAllImages(manifests)

	require.Len(t, images, 2)
	require.Contains(t, images, "europe-docker.pkg.dev/kyma-project/prod/telemetry-manager:v1.2.0")
	require.Contains(t, images, "europe-docker.pkg.dev/kyma-project/prod/telemetry-webhook:v1.0.0")
}

func TestGetAllImages_WhenCalledWithDuplicateImages_ReturnsDeduplicatedImages(t *testing.T) {
	service := image.NewService()
	manifests := []*unstructured.Unstructured{
		createDeployment("telemetry-manager-1", []containerSpec{
			{name: "manager", image: "europe-docker.pkg.dev/kyma-project/prod/telemetry-manager:v1.2.0"},
		}),
		createDeployment("telemetry-manager-2", []containerSpec{
			{name: "manager", image: "europe-docker.pkg.dev/kyma-project/prod/telemetry-manager:v1.2.0"},
		}),
	}

	images := service.GetAllImages(manifests)

	require.Len(t, images, 1)
	require.Contains(t, images, "europe-docker.pkg.dev/kyma-project/prod/telemetry-manager:v1.2.0")
}

func TestGetAllImages_WhenCalledWithManifestWithoutContainers_ReturnsEmptySlice(t *testing.T) {
	service := image.NewService()
	manifests := []*unstructured.Unstructured{
		{
			Object: map[string]interface{}{
				"kind": "Deployment",
				"metadata": map[string]interface{}{
					"name": "telemetry-manager",
				},
				"spec": map[string]interface{}{
					"template": map[string]interface{}{
						"spec": map[string]interface{}{},
					},
				},
			},
		},
	}

	images := service.GetAllImages(manifests)

	require.Empty(t, images)
}

func TestGetAllImages_WhenCalledWithInvalidContainerStructure_ReturnsEmptySlice(t *testing.T) {
	service := image.NewService()
	manifests := []*unstructured.Unstructured{
		{
			Object: map[string]interface{}{
				"kind": "Deployment",
				"metadata": map[string]interface{}{
					"name": "telemetry-manager",
				},
				"spec": map[string]interface{}{
					"template": map[string]interface{}{
						"spec": map[string]interface{}{
							"containers": "invalid-structure",
						},
					},
				},
			},
		},
	}

	images := service.GetAllImages(manifests)

	require.Empty(t, images)
}

func TestGetAllImages_WhenCalledWithEnvironmentVariablesButNoImageReferences_ReturnsOnlyContainerImages(t *testing.T) {
	service := image.NewService()
	manifests := []*unstructured.Unstructured{
		createDeploymentWithEnvImages("telemetry-manager", []containerSpec{
			{
				name:  "manager",
				image: "europe-docker.pkg.dev/kyma-project/prod/telemetry-manager:v1.2.0",
				envVars: []envVar{
					{name: "LOG_LEVEL", value: "info"},
					{name: "METRICS_PORT", value: "8080"},
				},
			},
		}),
	}

	images := service.GetAllImages(manifests)

	require.Len(t, images, 1)
	require.Contains(t, images, "europe-docker.pkg.dev/kyma-project/prod/telemetry-manager:v1.2.0")
}

func TestGetAllImages_WhenCalledWithMultipleWorkloadTypes_ReturnsAllImages(t *testing.T) {
	service := image.NewService()
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

	images := service.GetAllImages(manifests)

	require.Len(t, images, 2)
	require.Contains(t, images, "europe-docker.pkg.dev/kyma-project/prod/telemetry-manager:v1.2.0")
	require.Contains(t, images, "gcr.io/istio-release/proxyv2:1.19.0")
}

func TestAppendManifestImages_WhenCalledWithEmptyImages_ReturnsNoError(t *testing.T) {
	service := image.NewService()
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

	err := service.AppendManifestImages(descriptor, []string{})

	require.NoError(t, err)
	require.Empty(t, descriptor.Resources)
}

func TestAppendManifestImages_WhenCalledWithValidImages_AppendsResources(t *testing.T) {
	service := image.NewService()
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

	err := service.AppendManifestImages(descriptor, images)

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

func TestAppendManifestImages_WhenCalledWithInvalidImageFormat_ReturnsError(t *testing.T) {
	service := image.NewService()
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

	err := service.AppendManifestImages(descriptor, images)

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to append image")
}

func TestAppendManifestImages_WhenCalledWithComplexImageNames_AppendsResources(t *testing.T) {
	service := image.NewService()
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
		"registry.k8s.io/ingress-nginx/controller:v1.8.0",
	}

	err := service.AppendManifestImages(descriptor, images)

	require.NoError(t, err)
	require.Len(t, descriptor.Resources, 3)

	// Check that all resources have the correct label
	for _, resource := range descriptor.Resources {
		require.Len(t, resource.Labels, 1)
		require.Equal(t, "scan.security.kyma-project.io/type", resource.Labels[0].Name)

		var labelValue string
		err = json.Unmarshal(descriptor.Resources[0].Labels[0].Value, &labelValue)
		require.NoError(t, err)
		require.Equal(t, "manifest-image", labelValue)
		require.Equal(t, "v1", resource.Labels[0].Version)
	}
}

func TestAppendManifestImages_WhenCalledWithExistingResources_AppendsToExistingResources(t *testing.T) {
	service := image.NewService()
	existingAccess := ociartifact.New("existing-image:1.0.0")
	existingAccess.SetType(ociartifact.Type)

	existingResource := compdesc.Resource{
		ResourceMeta: compdesc.ResourceMeta{
			Type:     "existing-type",
			Relation: ocmv1.ExternalRelation,
			ElementMeta: compdesc.ElementMeta{
				Name:    "existing-resource",
				Version: "1.0.0",
			},
		},
		Access: existingAccess,
	}

	descriptor := &compdesc.ComponentDescriptor{
		ComponentSpec: compdesc.ComponentSpec{
			ObjectMeta: ocmv1.ObjectMeta{
				Name:     "kyma-project.io/module/telemetry",
				Version:  "1.0.0",
				Provider: ocmv1.Provider{Name: "kyma-project.io"},
			},
			Resources: []compdesc.Resource{existingResource},
		},
	}
	images := []string{"europe-docker.pkg.dev/kyma-project/prod/telemetry-manager:v1.2.0"}

	err := service.AppendManifestImages(descriptor, images)

	require.NoError(t, err)
	require.Len(t, descriptor.Resources, 2)
	require.Equal(t, "existing-resource", descriptor.Resources[0].Name)
	require.Equal(t, "telemetry-manager", descriptor.Resources[1].Name)
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
			name:              "invalid image URL - multiple tags",
			imageURL:          "docker.io/template-operator/test:latest:latest",
			expectedImageName: "",
			expectedImageTag:  "",
			expectedError:     errors.New("invalid image URL"),
		},
		{
			name:              "invalid image URL - no slashes",
			imageURL:          "docker.io",
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

// Mock parser implementation with call tracking
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
