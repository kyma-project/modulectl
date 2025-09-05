package component_test

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/kyma-project/modulectl/internal/common"
	"github.com/kyma-project/modulectl/internal/common/types/component"
	"github.com/kyma-project/modulectl/internal/service/contentprovider"
	"github.com/kyma-project/modulectl/internal/service/git"
	"github.com/kyma-project/modulectl/internal/service/image"
)

func TestNewConstructor(t *testing.T) {
	constructor := component.NewConstructor("test-component", "1.0.0")

	require.NotNil(t, constructor)
	require.Len(t, constructor.Components, 1)

	moduleComponent := constructor.Components[0]
	require.Equal(t, "test-component", moduleComponent.Name)
	require.Equal(t, "1.0.0", moduleComponent.Version)
}

func TestConstructor_Initialize(t *testing.T) {
	constructor := component.NewConstructor("test-component", "1.0.0")

	require.Len(t, constructor.Components, 1)
	moduleComponent := constructor.Components[0]
	require.Equal(t, "test-component", moduleComponent.Name)
	require.Equal(t, "1.0.0", moduleComponent.Version)
	require.Equal(t, common.ProviderName, moduleComponent.Provider.Name)
	require.Len(t, moduleComponent.Provider.Labels, 1)
	providerLabel := moduleComponent.Provider.Labels[0]
	require.Equal(t, common.BuiltByLabelKey, providerLabel.Name)
	require.Equal(t, common.BuiltByLabelValue, providerLabel.Value)
	require.Equal(t, common.VersionV1, providerLabel.Version)
	require.Empty(t, moduleComponent.Resources)
	require.Empty(t, moduleComponent.Sources)
}

func TestConstructor_AddGitSource(t *testing.T) {
	constructor := component.NewConstructor("test-component", "1.0.0")

	constructor.AddGitSource("https://github.com/kyma-project/modulectl", "abc123def456")

	require.Len(t, constructor.Components[0].Sources, 1)
	source := constructor.Components[0].Sources[0]
	require.Equal(t, common.OCMIdentityName, source.Name)
	require.Equal(t, component.GithubSourceType, source.Type)
	require.Equal(t, "1.0.0", source.Version)
	require.Len(t, source.Labels, 1)
	label := source.Labels[0]
	require.Equal(t, common.RefLabel, label.Name)
	require.Equal(t, git.HeadRef, label.Value)
	require.Equal(t, common.OCMVersion, label.Version)
	require.Equal(t, component.GithubAccessType, source.Access.Type)
	require.Equal(t, "https://github.com/kyma-project/modulectl", source.Access.RepoUrl)
	require.Equal(t, "abc123def456", source.Access.Commit)
}

func TestConstructor_AddLabel(t *testing.T) {
	constructor := component.NewConstructor("test-component", "1.0.0")

	initialLabelCount := len(constructor.Components[0].Labels)

	constructor.AddLabel("test-key", "test-value", common.VersionV1)

	require.Len(t, constructor.Components[0].Labels, initialLabelCount+1)

	var addedLabel *component.Label
	for _, label := range constructor.Components[0].Labels {
		if label.Name == "test-key" {
			addedLabel = &label
			break
		}
	}

	require.NotNil(t, addedLabel, "added label not found")
	require.Equal(t, "test-key", addedLabel.Name)
	require.Equal(t, "test-value", addedLabel.Value)
	require.Equal(t, common.VersionV1, addedLabel.Version)
}

func TestConstructor_AddLabel_Multiple(t *testing.T) {
	constructor := component.NewConstructor("test-component", "1.0.0")

	labels := []struct {
		key, value, version string
	}{
		{"environment", "production", common.VersionV1},
		{"team", "platform", common.VersionV2},
		{"criticality", "high", common.VersionV1},
		{"region", "us-east-1", common.VersionV1},
	}

	for _, label := range labels {
		constructor.AddLabel(label.key, label.value, label.version)
	}

	require.Len(t, constructor.Components[0].Labels, len(labels))

	for _, expectedLabel := range labels {
		found := false
		for _, actualLabel := range constructor.Components[0].Labels {
			if actualLabel.Name == expectedLabel.key &&
				actualLabel.Value == expectedLabel.value &&
				actualLabel.Version == expectedLabel.version {
				found = true
				break
			}
		}
		require.True(t, found, "label with key %s not found in component labels", expectedLabel.key)
	}
}

func TestConstructor_AddLabelToSources(t *testing.T) {
	constructor := component.NewConstructor("test-component", "1.0.0")

	constructor.AddGitSource("https://github.com/test/repo1", "commit1")
	constructor.AddGitSource("https://github.com/test/repo2", "commit2")

	initialLabelCounts := make([]int, len(constructor.Components[0].Sources))
	for i, source := range constructor.Components[0].Sources {
		initialLabelCounts[i] = len(source.Labels)
	}

	constructor.AddLabelToSources("test-key", "test-value", common.VersionV1)

	for i, source := range constructor.Components[0].Sources {
		require.Len(t, source.Labels, initialLabelCounts[i]+1, "source %d: label count mismatch", i)

		var foundLabel *component.Label
		for _, label := range source.Labels {
			if label.Name == "test-key" {
				foundLabel = &label
				break
			}
		}

		require.NotNil(t, foundLabel, "source %d: added label not found", i)
		require.Equal(t, "test-value", foundLabel.Value, "source %d: label value mismatch", i)
		require.Equal(t, common.VersionV1, foundLabel.Version, "source %d: label version mismatch", i)
	}
}

func TestConstructor_AddImageAsResource(t *testing.T) {
	constructor := component.NewConstructor("test-component", "1.0.0")

	imageInfo := &image.ImageInfo{
		Name:    "test-image",
		Tag:     "1.0.0",
		Digest:  "sha256:abc123",
		FullURL: "registry.io/test-image:1.0.0",
	}

	constructor.AddImageAsResource([]*image.ImageInfo{imageInfo})

	require.Len(t, constructor.Components[0].Resources, 1)
	resource := constructor.Components[0].Resources[0]
	require.Equal(t, component.OCIArtifactResourceType, resource.Type)
	require.Equal(t, component.OCIArtifactResourceRelation, resource.Relation)
	require.Len(t, resource.Labels, 1)

	expectedLabelName := common.SecScanBaseLabelKey + "/" + common.TypeLabelKey
	require.Equal(t, expectedLabelName, resource.Labels[0].Name)
	require.Equal(t, common.ThirdPartyImageLabelValue, resource.Labels[0].Value)
	require.Equal(t, common.OCMVersion, resource.Labels[0].Version)
	require.Equal(t, component.OCIArtifactAccessType, resource.Access.Type)
	require.Equal(t, imageInfo.FullURL, resource.Access.ImageReference)
}

func TestConstructor_AddImageAsResource_Multiple(t *testing.T) {
	constructor := component.NewConstructor("test-component", "1.0.0")

	imageInfos := []*image.ImageInfo{
		{
			Name:    "image1",
			Tag:     "1.0.0",
			Digest:  "sha256:abc123",
			FullURL: "registry.io/image1:1.0.0",
		},
		{
			Name:    "image2",
			Tag:     "2.0.0",
			Digest:  "sha256:def456",
			FullURL: "registry.io/image2:2.0.0",
		},
	}

	constructor.AddImageAsResource(imageInfos)

	require.Len(t, constructor.Components[0].Resources, 2)

	for i, resource := range constructor.Components[0].Resources {
		require.Equal(t, imageInfos[i].FullURL, resource.Access.ImageReference, "resource %d: image reference mismatch", i)
	}
}

func TestConstructor_AddRawManifestResource(t *testing.T) {
	constructor := component.NewConstructor("test-component", "1.0.0")

	constructor.AddRawManifestResource("/path/to/manifest.yaml")

	require.Len(t, constructor.Components[0].Resources, 1)
	resource := constructor.Components[0].Resources[len(constructor.Components[0].Resources)-1]
	require.Equal(t, common.RawManifestResourceName, resource.Name)
	require.Equal(t, component.DirectoryTreeResourceType, resource.Type)
	require.Equal(t, "1.0.0", resource.Version)
	require.Equal(t, component.DirectoryInputType, resource.Input.Type)
	require.Equal(t, "/path/to", resource.Input.Path)
	require.NotNil(t, resource.Input.IncludeFiles)
	require.Len(t, resource.Input.IncludeFiles, 1)
	require.Equal(t, "manifest.yaml", resource.Input.IncludeFiles[0])
}

func TestConstructor_AddDefaultCRResource(t *testing.T) {
	constructor := component.NewConstructor("test-component", "1.0.0")

	constructor.AddDefaultCRResource("/path/to/defaultcr.yaml")

	require.Len(t, constructor.Components[0].Resources, 1)
	resource := constructor.Components[0].Resources[len(constructor.Components[0].Resources)-1]
	require.Equal(t, common.DefaultCRResourceName, resource.Name)
	require.Equal(t, component.DirectoryTreeResourceType, resource.Type)
	require.Equal(t, "1.0.0", resource.Version)
	require.Equal(t, component.DirectoryInputType, resource.Input.Type)
	require.Equal(t, "/path/to", resource.Input.Path)
	require.NotNil(t, resource.Input.IncludeFiles)
	require.Len(t, resource.Input.IncludeFiles, 1)
	require.Equal(t, "defaultcr.yaml", resource.Input.IncludeFiles[0])
}

func TestConstructor_AddMetadataResource_Success(t *testing.T) {
	constructor := component.NewConstructor("test-component", "1.0.0")

	moduleConfig := &contentprovider.ModuleConfig{
		Name:    "test-module",
		Version: "1.0.0",
	}

	err := constructor.AddMetadataResource(moduleConfig)
	require.NoError(t, err)

	resource := constructor.Components[0].Resources[0]
	require.Equal(t, common.MetadataResourceName, resource.Name)
	require.Equal(t, component.PlainTextResourceType, resource.Type)
	require.Equal(t, "1.0.0", resource.Version)
	require.Equal(t, component.BinaryResourceInput, resource.Input.Type)
	require.NotEmpty(t, resource.Input.Data)

	_, err = base64.StdEncoding.DecodeString(resource.Input.Data)
	require.NoError(t, err, "expected input data to be valid base64")
}

func TestConstructor_AddMetadataResource_ReturnsError_WhenConfigNil(t *testing.T) {
	constructor := component.NewConstructor("test-component", "1.0.0")

	err := constructor.AddMetadataResource(nil)

	require.Error(t, err)
	require.Contains(t, err.Error(), "config is nil")
	require.Empty(t, constructor.Components[0].Resources)
}

func TestConstructor_CreateComponentConstructorFile_Success(t *testing.T) {
	constructor := component.NewConstructor("test-component", "1.0.0")

	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "constructor.yaml")

	err := constructor.CreateComponentConstructorFile(filePath)
	require.NoError(t, err)

	_, err = os.Stat(filePath)
	require.NoError(t, err, "constructor file was not created")

	content, err := os.ReadFile(filePath)
	require.NoError(t, err, "failed to read constructor file")

	var loadedConstructor component.Constructor
	err = yaml.Unmarshal(content, &loadedConstructor)
	require.NoError(t, err, "failed to unmarshal constructor file")

	require.Len(t, loadedConstructor.Components, 1, "expected 1 component in loaded constructor")
	require.Equal(t, "test-component", loadedConstructor.Components[0].Name)
	require.Equal(t, "1.0.0", loadedConstructor.Components[0].Version)
	require.Equal(t, constructor.Components[0].Provider, loadedConstructor.Components[0].Provider)
	require.Len(t, loadedConstructor.Components[0].Resources, len(constructor.Components[0].Resources))
	require.Len(t, loadedConstructor.Components[0].Sources, len(constructor.Components[0].Sources))
}

func TestConstructor_CreateComponentConstructorFile_ReturnsError_WhenOutputFilePathInvalid(t *testing.T) {
	constructor := component.NewConstructor("test-component", "1.0.0")

	filePath := "/invalid/path/constructor.yaml"
	err := constructor.CreateComponentConstructorFile(filePath)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unable to write component constructor file")
}
