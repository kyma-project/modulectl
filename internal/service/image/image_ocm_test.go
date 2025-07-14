package image_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"ocm.software/ocm/api/ocm/compdesc"
	ocmv1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"

	"fmt"
	"github.com/kyma-project/modulectl/internal/service/image"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/ociartifact"
)

func TestAddImagesToOcmDescriptor_WhenCalledWithValidImages_AppendsResources(t *testing.T) {
	mockParser := &mockManifestParser{}
	service, _ := image.NewService(mockParser)

	descriptor := createEmptyDescriptor()
	images := []string{
		"alpine:3.15.4",
		"nginx:1.21.0",
	}

	err := service.AddImagesToOcmDescriptor(descriptor, images)

	require.NoError(t, err)
	require.Len(t, descriptor.Resources, 2)

	resource1 := descriptor.Resources[0]
	fmt.Println(resource1)
	require.Equal(t, "alpine", resource1.Name)
	require.Equal(t, "3.15.4", resource1.Version)
	require.Equal(t, "ociArtifact", resource1.Type)
	require.Len(t, resource1.Labels, 1)
	require.Equal(t, "scan.security.kyma-project.io/type", resource1.Labels[0].Name)

	var labelValue1 string
	err = json.Unmarshal(resource1.Labels[0].Value, &labelValue1)
	require.NoError(t, err)
	require.Equal(t, "third-party-image", labelValue1)

	resource2 := descriptor.Resources[1]
	require.Equal(t, "nginx", resource2.Name)
	require.Equal(t, "1.21.0", resource2.Version)
	require.Equal(t, "ociArtifact", resource2.Type)
}

func TestAddImagesToOcmDescriptor_WhenCalledWithComplexRegistryPath_AppendsResource(t *testing.T) {
	mockParser := &mockManifestParser{}
	service, _ := image.NewService(mockParser)

	descriptor := createEmptyDescriptor()
	images := []string{
		"europe-docker.pkg.dev/kyma-project/prod/external/istio/proxyv2:1.25.3-distroless",
	}

	err := service.AddImagesToOcmDescriptor(descriptor, images)

	require.NoError(t, err)
	require.Len(t, descriptor.Resources, 1)

	resource := descriptor.Resources[0]
	require.Equal(t, "proxyv2", resource.Name)
	require.Equal(t, "1.25.3-distroless", resource.Version)
	require.Equal(t, "ociArtifact", resource.Type)
}

func TestAddImagesToOcmDescriptor_WhenCalledWithGcrImage_AppendsResource(t *testing.T) {
	mockParser := &mockManifestParser{}
	service, _ := image.NewService(mockParser)

	descriptor := createEmptyDescriptor()
	images := []string{
		"gcr.io/kubebuilder/kube-rbac-proxy:v0.13.1",
	}

	err := service.AddImagesToOcmDescriptor(descriptor, images)

	require.NoError(t, err)
	require.Len(t, descriptor.Resources, 1)

	resource := descriptor.Resources[0]
	require.Equal(t, "kube-rbac-proxy", resource.Name)
	require.Equal(t, "v0.13.1", resource.Version)
	require.Equal(t, "ociArtifact", resource.Type)
}

func TestAddImagesToOcmDescriptor_WhenCalledWithInvalidImage_ReturnsError(t *testing.T) {
	mockParser := &mockManifestParser{}
	service, _ := image.NewService(mockParser)

	descriptor := createEmptyDescriptor()
	images := []string{
		"alpine:v1.0.0",
		"invalid-image-no-tag",
	}

	err := service.AddImagesToOcmDescriptor(descriptor, images)

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to append image")
	require.Contains(t, err.Error(), "invalid-image-no-tag")
}

func TestAddImagesToOcmDescriptor_WhenCalledWithEmptyImageList_DoesNothing(t *testing.T) {
	mockParser := &mockManifestParser{}
	service, _ := image.NewService(mockParser)

	descriptor := createEmptyDescriptor()
	images := []string{}

	err := service.AddImagesToOcmDescriptor(descriptor, images)

	require.NoError(t, err)
	require.Empty(t, descriptor.Resources)
}

func TestAddImagesToOcmDescriptor_WhenCalledWithRegistryPortImage_AppendsResource(t *testing.T) {
	mockParser := &mockManifestParser{}
	service, _ := image.NewService(mockParser)

	descriptor := createEmptyDescriptor()
	images := []string{
		"localhost:5000/myimage:v1.0.0",
	}

	err := service.AddImagesToOcmDescriptor(descriptor, images)

	require.NoError(t, err)
	require.Len(t, descriptor.Resources, 1)

	resource := descriptor.Resources[0]
	require.Equal(t, "myimage", resource.Name)
	require.Equal(t, "v1.0.0", resource.Version)
	require.Equal(t, "ociArtifact", resource.Type)
}

func TestAddImagesToOcmDescriptor_WhenCalledWithDockerHubImage_AppendsResource(t *testing.T) {
	mockParser := &mockManifestParser{}
	service, _ := image.NewService(mockParser)

	descriptor := createEmptyDescriptor()
	images := []string{
		"istio/proxyv2:1.19.0",
	}

	err := service.AddImagesToOcmDescriptor(descriptor, images)

	require.NoError(t, err)
	require.Len(t, descriptor.Resources, 1)

	resource := descriptor.Resources[0]
	require.Equal(t, "proxyv2", resource.Name)
	require.Equal(t, "1.19.0", resource.Version)
	require.Equal(t, "ociArtifact", resource.Type)
}

func TestAddImagesToOcmDescriptor_WhenCalledWithMultipleImages_CreatesCorrectLabels(t *testing.T) {
	mockParser := &mockManifestParser{}
	service, _ := image.NewService(mockParser)

	descriptor := createEmptyDescriptor()
	images := []string{
		"alpine:3.15.4",
		"nginx:1.21.0",
	}

	err := service.AddImagesToOcmDescriptor(descriptor, images)

	require.NoError(t, err)
	require.Len(t, descriptor.Resources, 2)

	for _, resource := range descriptor.Resources {
		require.Len(t, resource.Labels, 1)
		require.Equal(t, "scan.security.kyma-project.io/type", resource.Labels[0].Name)
		require.Equal(t, "v1", resource.Labels[0].Version)
		require.NotNil(t, resource.Access)
	}
}

func TestAddImagesToOcmDescriptor_WhenCalledWithInvalidDigest_ReturnsError(t *testing.T) {
	mockParser := &mockManifestParser{}
	service, _ := image.NewService(mockParser)

	descriptor := createEmptyDescriptor()
	images := []string{
		"alpine@sha256:invalid-digest",
	}

	err := service.AddImagesToOcmDescriptor(descriptor, images)

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to append image")
}

func TestAddImagesToOcmDescriptor_WhenCalledWithMalformedImage_ReturnsError(t *testing.T) {
	mockParser := &mockManifestParser{}
	service, _ := image.NewService(mockParser)

	descriptor := createEmptyDescriptor()
	images := []string{
		"",
		"alpine:",
		"alpine@",
	}

	for _, img := range images {
		err := service.AddImagesToOcmDescriptor(descriptor, []string{img})
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to append image")
	}
}

func TestAddImagesToOcmDescriptor_WhenCalledWithExistingResources_AppendsToExisting(t *testing.T) {
	mockParser := &mockManifestParser{}
	service, _ := image.NewService(mockParser)

	existingResource := compdesc.Resource{
		ResourceMeta: compdesc.ResourceMeta{
			Type:     "existing-type",
			Relation: ocmv1.LocalRelation,
			ElementMeta: compdesc.ElementMeta{
				Name:    "existing",
				Version: "1.0.0",
			},
		},
		Access: ociartifact.New("existing:1.0.0"),
	}

	descriptor := createDescriptorWithResource(existingResource)
	images := []string{
		"alpine:3.15.4",
	}

	err := service.AddImagesToOcmDescriptor(descriptor, images)

	require.NoError(t, err)
	require.Len(t, descriptor.Resources, 2)
	require.Equal(t, "existing", descriptor.Resources[0].Name)
	require.Equal(t, "alpine", descriptor.Resources[1].Name)
}

func createEmptyDescriptor() *compdesc.ComponentDescriptor {
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
	compdesc.DefaultResources(descriptor)
	return descriptor
}

func createDescriptorWithResource(resource compdesc.Resource) *compdesc.ComponentDescriptor {
	descriptor := &compdesc.ComponentDescriptor{
		ComponentSpec: compdesc.ComponentSpec{
			ObjectMeta: ocmv1.ObjectMeta{
				Name:     "kyma-project.io/module/telemetry",
				Version:  "1.0.0",
				Provider: ocmv1.Provider{Name: "kyma-project.io"},
			},
			Resources: []compdesc.Resource{resource},
		},
	}
	compdesc.DefaultResources(descriptor)
	return descriptor
}
