package resources_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"ocm.software/ocm/api/ocm/compdesc"
	ocmv1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/ociartifact"
	ociartifacttypes "ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/types/ociartifact"

	"github.com/kyma-project/modulectl/internal/service/componentdescriptor/resources"
	"github.com/kyma-project/modulectl/internal/service/image"
)

func TestErrInvalidImageFormat_WhenAccessed_ReturnsCorrectMessage(t *testing.T) {
	err := resources.ErrInvalidImageFormat
	require.Equal(t, "invalid image url format", err.Error())
}

func TestNewOciArtifactResource_WhenImageInfoIsNil_ReturnsError(t *testing.T) {
	result, err := resources.NewOciArtifactResource(nil)

	require.Nil(t, result)
	require.Error(t, err)
	require.ErrorIs(t, err, resources.ErrInvalidImageFormat)
	require.Contains(t, err.Error(), "image info is nil or empty")
}

func TestNewOciArtifactResource_WhenImageInfoHasEmptyURL_ReturnsError(t *testing.T) {
	imageInfo := createImageInfo("", "test-image", "v1.0.0", "")

	result, err := resources.NewOciArtifactResource(imageInfo)

	require.Nil(t, result)
	require.Error(t, err)
	require.ErrorIs(t, err, resources.ErrInvalidImageFormat)
	require.Contains(t, err.Error(), "image info is nil or empty")
}

func TestNewOciArtifactResource_WhenValidImageWithSemverTag_CreatesResourceCorrectly(t *testing.T) {
	imageInfo := createImageInfo("registry.io/myimage:v1.2.3", "myimage", "v1.2.3", "")

	result, err := resources.NewOciArtifactResource(imageInfo)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, ociartifacttypes.TYPE, result.Type)
	require.Equal(t, ocmv1.ExternalRelation, result.Relation)
	require.Equal(t, "myimage", result.Name)
	require.Equal(t, "v1.2.3", result.Version)
	require.Len(t, result.Labels, 1)
	require.Equal(t, ociartifact.Type, result.Access.GetType())

	label := result.Labels[0]
	require.Equal(t, "scan.security.kyma-project.io/type", label.Name)
	require.Equal(t, "\"third-party-image\"", string(label.Value))
}

func TestNewOciArtifactResource_WhenValidImageWithSemverNoVPrefix_CreatesResourceCorrectly(t *testing.T) {
	imageInfo := createImageInfo("registry.io/myimage:1.2.3", "myimage", "1.2.3", "")

	result, err := resources.NewOciArtifactResource(imageInfo)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, "1.2.3", result.Version)
}

func TestNewOciArtifactResource_WhenValidImageWithDigestAndSemverTag_CreatesResourceWithDigestVersion(t *testing.T) {
	digest := "sha256:abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
	imageInfo := createImageInfo("registry.io/myimage@"+digest, "myimage", "v1.2.3", digest)

	result, err := resources.NewOciArtifactResource(imageInfo)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, "v1.2.3+sha256.sha256:abcde", result.Version)
	require.Equal(t, "myimage-sha256:a", result.Name)
}

func TestNewOciArtifactResource_WhenValidImageWithDigestAndInvalidTag_CreatesResourceWithNormalizedTag(t *testing.T) {
	digest := "sha256:abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
	imageInfo := createImageInfo("registry.io/myimage@"+digest, "myimage", "latest", digest)

	result, err := resources.NewOciArtifactResource(imageInfo)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, "0.0.0-latest+sha256.sha256:abcde", result.Version)
	require.Equal(t, "myimage-sha256:a", result.Name)
}

func TestNewOciArtifactResource_WhenDigestWithoutTag_CreatesResourceWithDefaultVersion(t *testing.T) {
	digest := "sha256:abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
	imageInfo := createImageInfo("registry.io/myimage@"+digest, "myimage", "", digest)

	result, err := resources.NewOciArtifactResource(imageInfo)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, "0.0.0+sha256.sha256:abcde", result.Version)
	require.Equal(t, "myimage-sha256:a", result.Name)
}

func TestNewOciArtifactResource_WhenVeryShortDigest_HandlesCorrectly(t *testing.T) {
	digest := "sha256:abcdefghijkl"
	imageInfo := createImageInfo("registry.io/myimage@"+digest, "myimage", "", digest)

	result, err := resources.NewOciArtifactResource(imageInfo)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, "0.0.0+sha256.sha256:abcde", result.Version)
	require.Equal(t, "myimage-sha256:a", result.Name)
}

func TestNewOciArtifactResource_WhenValidImageWithInvalidTag_CreatesResourceWithNormalizedVersion(t *testing.T) {
	imageInfo := createImageInfo("registry.io/myimage:latest", "myimage", "latest", "")

	result, err := resources.NewOciArtifactResource(imageInfo)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, "0.0.0-latest", result.Version)
	require.Equal(t, "myimage", result.Name)
}

func TestNewOciArtifactResource_WhenComplexTagNormalization_HandlesSpecialCharacters(t *testing.T) {
	imageInfo := createImageInfo("registry.io/myimage:feature/branch@with#special$chars", "myimage", "feature/branch@with#special$chars", "")

	result, err := resources.NewOciArtifactResource(imageInfo)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, "0.0.0-feature-branch-with-special-chars", result.Version)
}

func TestNewOciArtifactResource_WhenTagWithLeadingTrailingSpecialChars_NormalizesCorrectly(t *testing.T) {
	imageInfo := createImageInfo("registry.io/myimage:---.valid-tag.---", "myimage", "---.valid-tag.---", "")

	result, err := resources.NewOciArtifactResource(imageInfo)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, "0.0.0-valid-tag", result.Version)
}

func TestNewOciArtifactResource_WhenEmptyTagNormalization_UsesUnknown(t *testing.T) {
	imageInfo := createImageInfo("registry.io/myimage:@#$%", "myimage", "@#$%", "")

	result, err := resources.NewOciArtifactResource(imageInfo)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, "0.0.0-unknown", result.Version)
}

func TestNewOciArtifactResource_WhenTagWithOnlySpecialChars_UsesUnknown(t *testing.T) {
	imageInfo := createImageInfo("registry.io/myimage:...-...", "myimage", "...-...", "")

	result, err := resources.NewOciArtifactResource(imageInfo)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, "0.0.0-unknown", result.Version)
}

func TestNewOciArtifactResource_WhenSemverWithPrerelease_IdentifiesAsValidSemver(t *testing.T) {
	imageInfo := createImageInfo("registry.io/myimage:v1.2.3-alpha.1", "myimage", "v1.2.3-alpha.1", "")

	result, err := resources.NewOciArtifactResource(imageInfo)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, "v1.2.3-alpha.1", result.Version)
}

func TestNewOciArtifactResource_WhenSemverWithBuildMetadata_IdentifiesAsValidSemver(t *testing.T) {
	imageInfo := createImageInfo("registry.io/myimage:1.2.3+build.123", "myimage", "1.2.3+build.123", "")

	result, err := resources.NewOciArtifactResource(imageInfo)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, "1.2.3+build.123", result.Version)
}

func TestNewOciArtifactResource_WhenSemverWithPrereleaseAndBuildMetadata_IdentifiesAsValidSemver(t *testing.T) {
	imageInfo := createImageInfo("registry.io/myimage:v2.1.0-rc.1+build.456", "myimage", "v2.1.0-rc.1+build.456", "")

	result, err := resources.NewOciArtifactResource(imageInfo)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, "v2.1.0-rc.1+build.456", result.Version)
}

func TestNewOciArtifactResource_WhenInvalidSemverMissingPatch_CreatesNormalizedVersion(t *testing.T) {
	imageInfo := createImageInfo("registry.io/myimage:v1.2", "myimage", "v1.2", "")

	result, err := resources.NewOciArtifactResource(imageInfo)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, "0.0.0-v1.2", result.Version)
}

func TestNewOciArtifactResource_WhenInvalidSemverWithLetters_CreatesNormalizedVersion(t *testing.T) {
	imageInfo := createImageInfo("registry.io/myimage:v1.2.3a", "myimage", "v1.2.3a", "")

	result, err := resources.NewOciArtifactResource(imageInfo)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, "0.0.0-v1.2.3a", result.Version)
}

func TestAddResourceIfNotExists_WhenResourceDoesNotExist_AddsResource(t *testing.T) {
	descriptor := &compdesc.ComponentDescriptor{}
	resource := createResource("test-resource", "v1.0.0")

	resources.AddResourceIfNotExists(descriptor, resource)

	require.Len(t, descriptor.Resources, 1)
	require.Equal(t, "test-resource", descriptor.Resources[0].Name)
	require.Equal(t, "v1.0.0", descriptor.Resources[0].Version)
}

func TestAddResourceIfNotExists_WhenResourceExists_DoesNotAddDuplicate(t *testing.T) {
	descriptor := &compdesc.ComponentDescriptor{}
	descriptor.Resources = append(descriptor.Resources, *createResource("test-resource", "v1.0.0"))
	newResource := createResource("test-resource", "v1.0.0")

	resources.AddResourceIfNotExists(descriptor, newResource)

	require.Len(t, descriptor.Resources, 1)
	require.Equal(t, "test-resource", descriptor.Resources[0].Name)
	require.Equal(t, "v1.0.0", descriptor.Resources[0].Version)
}

func TestAddResourceIfNotExists_WhenResourceWithSameNameButDifferentVersion_AddsResource(t *testing.T) {
	descriptor := &compdesc.ComponentDescriptor{}
	descriptor.Resources = append(descriptor.Resources, *createResource("test-resource", "v1.0.0"))
	newResource := createResource("test-resource", "v2.0.0")

	resources.AddResourceIfNotExists(descriptor, newResource)

	require.Len(t, descriptor.Resources, 2)
	require.Equal(t, "test-resource", descriptor.Resources[0].Name)
	require.Equal(t, "v1.0.0", descriptor.Resources[0].Version)
	require.Equal(t, "test-resource", descriptor.Resources[1].Name)
	require.Equal(t, "v2.0.0", descriptor.Resources[1].Version)
}

func TestAddResourceIfNotExists_WhenResourceWithDifferentNameButSameVersion_AddsResource(t *testing.T) {
	descriptor := &compdesc.ComponentDescriptor{}
	descriptor.Resources = append(descriptor.Resources, *createResource("test-resource-1", "v1.0.0"))
	newResource := createResource("test-resource-2", "v1.0.0")

	resources.AddResourceIfNotExists(descriptor, newResource)

	require.Len(t, descriptor.Resources, 2)
	require.Equal(t, "test-resource-1", descriptor.Resources[0].Name)
	require.Equal(t, "test-resource-2", descriptor.Resources[1].Name)
}

func TestAddResourceIfNotExists_WhenMultipleResourcesExist_ChecksAllForDuplicates(t *testing.T) {
	descriptor := &compdesc.ComponentDescriptor{}
	descriptor.Resources = append(descriptor.Resources, *createResource("resource-1", "v1.0.0"))
	descriptor.Resources = append(descriptor.Resources, *createResource("resource-2", "v1.0.0"))
	descriptor.Resources = append(descriptor.Resources, *createResource("resource-3", "v1.0.0"))

	newResource := createResource("resource-2", "v1.0.0")

	resources.AddResourceIfNotExists(descriptor, newResource)

	require.Len(t, descriptor.Resources, 3)
}

func TestAddResourceIfNotExists_WhenEmptyResourcesList_AddsResource(t *testing.T) {
	descriptor := &compdesc.ComponentDescriptor{}
	resource := createResource("new-resource", "v1.0.0")

	resources.AddResourceIfNotExists(descriptor, resource)

	require.Len(t, descriptor.Resources, 1)
	require.Equal(t, "new-resource", descriptor.Resources[0].Name)
}

func createImageInfo(fullURL, name, tag, digest string) *image.ImageInfo {
	return &image.ImageInfo{
		FullURL: fullURL,
		Name:    name,
		Tag:     tag,
		Digest:  digest,
	}
}

func createResource(name, version string) *compdesc.Resource {
	return &compdesc.Resource{
		ResourceMeta: compdesc.ResourceMeta{
			ElementMeta: compdesc.ElementMeta{
				Name:    name,
				Version: version,
			},
		},
	}
}
