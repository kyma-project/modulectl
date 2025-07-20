package componentdescriptor_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"ocm.software/ocm/api/ocm/compdesc"
	ocmv1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/ociartifact"
	ociartifacttypes "ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/types/ociartifact"

	"github.com/kyma-project/modulectl/internal/service/componentdescriptor"
)

func TestAddImagesToOcmDescriptor_WhenCalledWithValidImages_AppendsResources(t *testing.T) {
	descriptor := createEmptyDescriptor()
	images := []string{
		"alpine:3.15.4",
		"nginx:1.21.0",
	}

	err := componentdescriptor.AddImagesToOcmDescriptor(descriptor, images)

	require.NoError(t, err)
	require.Len(t, descriptor.Resources, 2)

	resource1 := descriptor.Resources[0]
	require.Equal(t, "alpine", resource1.Name)
	require.Equal(t, "3.15.4", resource1.Version)
	require.Equal(t, ociartifacttypes.TYPE, resource1.Type)
	require.Equal(t, ocmv1.ExternalRelation, resource1.Relation)
	require.Len(t, resource1.Labels, 1)
	require.Equal(t, "scan.security.kyma-project.io/type", resource1.Labels[0].Name)

	var labelValue1 string
	err = json.Unmarshal(resource1.Labels[0].Value, &labelValue1)
	require.NoError(t, err)
	require.Equal(t, "third-party-image", labelValue1)

	resource2 := descriptor.Resources[1]
	require.Equal(t, "nginx", resource2.Name)
	require.Equal(t, "1.21.0", resource2.Version)
	require.Equal(t, ociartifacttypes.TYPE, resource2.Type)
}

func TestAddImagesToOcmDescriptor_WhenCalledWithComplexRegistryPath_AppendsResource(t *testing.T) {
	descriptor := createEmptyDescriptor()
	images := []string{
		"europe-docker.pkg.dev/kyma-project/prod/external/istio/proxyv2:1.25.3-distroless",
	}

	err := componentdescriptor.AddImagesToOcmDescriptor(descriptor, images)

	require.NoError(t, err)
	require.Len(t, descriptor.Resources, 1)

	resource := descriptor.Resources[0]
	require.Equal(t, "proxyv2", resource.Name)
	require.Equal(t, "1.25.3-distroless", resource.Version)
	require.Equal(t, ociartifacttypes.TYPE, resource.Type)
}

func TestAddImagesToOcmDescriptor_WhenCalledWithGcrImage_AppendsResource(t *testing.T) {
	descriptor := createEmptyDescriptor()
	images := []string{
		"gcr.io/kubebuilder/kube-rbac-proxy:v0.13.1",
	}

	err := componentdescriptor.AddImagesToOcmDescriptor(descriptor, images)

	require.NoError(t, err)
	require.Len(t, descriptor.Resources, 1)

	resource := descriptor.Resources[0]
	require.Equal(t, "kube-rbac-proxy", resource.Name)
	require.Equal(t, "v0.13.1", resource.Version)
	require.Equal(t, ociartifacttypes.TYPE, resource.Type)
}

func TestAddImagesToOcmDescriptor_WhenCalledWithInvalidImage_ReturnsError(t *testing.T) {
	descriptor := createEmptyDescriptor()
	images := []string{
		"alpine:v1.0.0",
		"invalid-image-no-tag",
	}

	err := componentdescriptor.AddImagesToOcmDescriptor(descriptor, images)

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to append image")
	require.Contains(t, err.Error(), "invalid-image-no-tag")
}

func TestAddImagesToOcmDescriptor_WhenCalledWithEmptyImageList_DoesNothing(t *testing.T) {
	descriptor := createEmptyDescriptor()
	images := []string{}

	err := componentdescriptor.AddImagesToOcmDescriptor(descriptor, images)

	require.NoError(t, err)
	require.Empty(t, descriptor.Resources)
}

func TestAddImagesToOcmDescriptor_WhenCalledWithRegistryPortImage_AppendsResource(t *testing.T) {
	descriptor := createEmptyDescriptor()
	images := []string{
		"localhost:5000/myimage:v1.0.0",
	}

	err := componentdescriptor.AddImagesToOcmDescriptor(descriptor, images)

	require.NoError(t, err)
	require.Len(t, descriptor.Resources, 1)

	resource := descriptor.Resources[0]
	require.Equal(t, "myimage", resource.Name)
	require.Equal(t, "v1.0.0", resource.Version)
	require.Equal(t, ociartifacttypes.TYPE, resource.Type)
}

func TestAddImagesToOcmDescriptor_WhenCalledWithDockerHubImage_AppendsResource(t *testing.T) {
	descriptor := createEmptyDescriptor()
	images := []string{
		"istio/proxyv2:1.19.0",
	}

	err := componentdescriptor.AddImagesToOcmDescriptor(descriptor, images)

	require.NoError(t, err)
	require.Len(t, descriptor.Resources, 1)

	resource := descriptor.Resources[0]
	require.Equal(t, "proxyv2", resource.Name)
	require.Equal(t, "1.19.0", resource.Version)
	require.Equal(t, ociartifacttypes.TYPE, resource.Type)
}

func TestAddImagesToOcmDescriptor_WhenCalledWithMultipleImages_CreatesCorrectLabels(t *testing.T) {
	descriptor := createEmptyDescriptor()
	images := []string{
		"alpine:3.15.4",
		"nginx:1.21.0",
	}

	err := componentdescriptor.AddImagesToOcmDescriptor(descriptor, images)

	require.NoError(t, err)
	require.Len(t, descriptor.Resources, 2)

	for _, resource := range descriptor.Resources {
		require.Len(t, resource.Labels, 1)
		require.Equal(t, "scan.security.kyma-project.io/type", resource.Labels[0].Name)
		require.Equal(t, "v1", resource.Labels[0].Version)
		require.NotNil(t, resource.Access)
		require.Equal(t, ociartifact.Type, resource.Access.GetType())
	}
}

func TestAddImagesToOcmDescriptor_WhenCalledWithDigestImage_AppendsResourceWithConvertedVersion(t *testing.T) {
	descriptor := createEmptyDescriptor()
	images := []string{
		"alpine@sha256:abcd1234567890abcdef1234567890abcdef1234567890abcdef1234567890ab",
	}

	err := componentdescriptor.AddImagesToOcmDescriptor(descriptor, images)

	require.NoError(t, err)
	require.Len(t, descriptor.Resources, 1)

	resource := descriptor.Resources[0]
	require.Equal(t, "alpine", resource.Name)
	require.Equal(t, "0.0.0+sha256.abcd1234567890abcdef1234567890abcdef1234567890abcdef1234567890ab", resource.Version)
	require.Equal(t, ociartifacttypes.TYPE, resource.Type)
}

func TestAddImagesToOcmDescriptor_WhenCalledWithMalformedImage_ReturnsError(t *testing.T) {
	descriptor := createEmptyDescriptor()
	images := []string{
		"",
		"alpine:",
		"alpine@",
	}

	for _, img := range images {
		err := componentdescriptor.AddImagesToOcmDescriptor(descriptor, []string{img})
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to append image")
	}
}

func TestAddImagesToOcmDescriptor_WhenCalledWithExistingResources_AppendsToExisting(t *testing.T) {
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

	err := componentdescriptor.AddImagesToOcmDescriptor(descriptor, images)

	require.NoError(t, err)
	require.Len(t, descriptor.Resources, 2)
	require.Equal(t, "existing", descriptor.Resources[0].Name)
	require.Equal(t, "alpine", descriptor.Resources[1].Name)
}

func TestAddImagesToOcmDescriptor_WhenCalledWithNilDescriptor_Panics(t *testing.T) {
	var descriptor *compdesc.ComponentDescriptor
	images := []string{"alpine:3.15.4"}

	require.Panics(t, func() {
		_ = componentdescriptor.AddImagesToOcmDescriptor(descriptor, images)
	})
}

func TestAddImagesToOcmDescriptor_WhenCalledWithImageWithoutTag_ReturnsError(t *testing.T) {
	descriptor := createEmptyDescriptor()
	images := []string{
		"alpine",
	}

	err := componentdescriptor.AddImagesToOcmDescriptor(descriptor, images)

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to append image")
}

func TestAddImagesToOcmDescriptor_WhenCalledWithValidImageAfterError_StopsProcessing(t *testing.T) {
	descriptor := createEmptyDescriptor()
	images := []string{
		"alpine:3.15.4",
		"", // invalid image
		"nginx:1.21.0",
	}

	err := componentdescriptor.AddImagesToOcmDescriptor(descriptor, images)

	require.Error(t, err)
	require.Len(t, descriptor.Resources, 1) // Only first image should be added
	require.Equal(t, "alpine", descriptor.Resources[0].Name)
}

func TestAddImagesToOcmDescriptor_WhenCalledWithVariousTagFormats_AppendsResourcesWithCorrectVersions(t *testing.T) {
	descriptor := createEmptyDescriptor()
	images := []string{
		"alpine:v1.2.3",
		"nginx:1.0.0",
		"kyma-project.io/myimage:2.2.3",
		"ghcr.io/mymodule/anotherimage:4.5.6",
		"registry.k3d.localhost:5000/myapp:123",
	}

	err := componentdescriptor.AddImagesToOcmDescriptor(descriptor, images)

	require.NoError(t, err)
	require.Len(t, descriptor.Resources, 5)

	expectedVersions := []string{"v1.2.3", "1.0.0", "2.2.3", "4.5.6", "123"}
	for i, resource := range descriptor.Resources {
		require.Equal(t, expectedVersions[i], resource.Version)
	}
}

func TestAddImagesToOcmDescriptor_WhenCalledAfterDefaults_MaintainsDescriptorValidity(t *testing.T) {
	descriptor := createEmptyDescriptor()
	images := []string{
		"alpine:3.15.4",
		"nginx:1.21.0",
	}

	err := componentdescriptor.AddImagesToOcmDescriptor(descriptor, images)

	require.NoError(t, err)

	err = compdesc.Validate(descriptor)
	require.NoError(t, err)
}

func TestAddImagesToOcmDescriptor_WhenCalledWithImageWithMultipleSlashes_ExtractsCorrectName(t *testing.T) {
	descriptor := createEmptyDescriptor()
	images := []string{
		"registry.example.com/team/project/subproject/app:v1.0.0",
	}

	err := componentdescriptor.AddImagesToOcmDescriptor(descriptor, images)

	require.NoError(t, err)
	require.Len(t, descriptor.Resources, 1)

	resource := descriptor.Resources[0]
	require.Equal(t, "app", resource.Name)
	require.Equal(t, "v1.0.0", resource.Version)
}

func TestAddImagesToOcmDescriptor_WhenCalledWithShortDigest_ReturnsError(t *testing.T) {
	descriptor := createEmptyDescriptor()
	images := []string{
		"alpine@sha256:short",
	}

	err := componentdescriptor.AddImagesToOcmDescriptor(descriptor, images)

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to append image")
}

// Test helper functions

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
