package component_test

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/kyma-project/modulectl/internal/common"
	"github.com/kyma-project/modulectl/internal/common/types/component"
	"github.com/kyma-project/modulectl/internal/service/contentprovider"
	"github.com/kyma-project/modulectl/internal/service/git"
	"github.com/kyma-project/modulectl/internal/service/image"
	"gopkg.in/yaml.v3"
)

func TestNewConstructor(t *testing.T) {
	constructor := component.NewConstructor("test-component", "1.0.0")

	if constructor == nil {
		t.Fatal("expected constructor to be non-nil")
	}

	if len(constructor.Components) != 1 {
		t.Errorf("expected 1 component, got length %d", len(constructor.Components))
	}

	// Test that the component is properly initialized
	moduleComponent := constructor.Components[0]
	if moduleComponent.Name != "test-component" {
		t.Errorf("expected component name test-component, got %s", moduleComponent.Name)
	}

	if moduleComponent.Version != "1.0.0" {
		t.Errorf("expected component version 1.0.0, got %s", moduleComponent.Version)
	}
}

func TestConstructor_Initialize(t *testing.T) {
	constructor := component.NewConstructor("test-component", "1.0.0")

	if len(constructor.Components) != 1 {
		t.Fatalf("expected 1 component, got %d", len(constructor.Components))
	}

	moduleComponent := constructor.Components[0]
	if moduleComponent.Name != "test-component" {
		t.Errorf("expected component name test-component, got %s", moduleComponent.Name)
	}

	if moduleComponent.Version != "1.0.0" {
		t.Errorf("expected component version 1.0.0, got %s", moduleComponent.Version)
	}

	if moduleComponent.Provider.Name != common.ProviderName {
		t.Errorf("expected provider name %s, got %s", common.ProviderName, moduleComponent.Provider.Name)
	}

	if len(moduleComponent.Provider.Labels) != 1 {
		t.Fatalf("expected 1 provider label, got %d", len(moduleComponent.Provider.Labels))
	}

	providerLabel := moduleComponent.Provider.Labels[0]
	if providerLabel.Name != common.BuiltByLabelKey {
		t.Errorf("expected label name %s, got %s", common.BuiltByLabelKey, providerLabel.Name)
	}

	if providerLabel.Value != common.BuiltByLabelValue {
		t.Errorf("expected label value %s, got %s", common.BuiltByLabelValue, providerLabel.Value)
	}

	if providerLabel.Version != common.VersionV1 {
		t.Errorf("expected label version %s, got %s", common.VersionV1, providerLabel.Version)
	}

	if len(moduleComponent.Resources) != 0 {
		t.Errorf("expected empty Resources slice, got length %d", len(moduleComponent.Resources))
	}

	if len(moduleComponent.Sources) != 0 {
		t.Errorf("expected empty Sources slice, got length %d", len(moduleComponent.Sources))
	}
}

func TestConstructor_AddGitSource(t *testing.T) {
	constructor := component.NewConstructor("test-component", "1.0.0")

	constructor.AddGitSource("https://github.com/kyma-project/modulectl", "abc123def456")

	if len(constructor.Components[0].Sources) != 1 {
		t.Fatalf("expected 1 source, got %d", len(constructor.Components[0].Sources))
	}

	source := constructor.Components[0].Sources[0]
	if source.Name != common.OCMIdentityName {
		t.Errorf("expected source name %s, got %s", common.OCMIdentityName, source.Name)
	}

	if source.Type != component.GithubSourceType {
		t.Errorf("expected source type %s, got %s", component.GithubSourceType, source.Type)
	}

	if source.Version != "1.0.0" {
		t.Errorf("expected source version 1.0.0, got %s", source.Version)
	}

	if len(source.Labels) != 1 {
		t.Fatalf("expected 1 label, got %d", len(source.Labels))
	}

	label := source.Labels[0]
	if label.Name != common.RefLabel {
		t.Errorf("expected label name %s, got %s", common.RefLabel, label.Name)
	}

	if label.Value != git.HeadRef {
		t.Errorf("expected label value %s, got %s", git.HeadRef, label.Value)
	}

	if label.Version != common.OCMVersion {
		t.Errorf("expected label version %s, got %s", common.OCMVersion, label.Version)
	}

	if source.Access.Type != component.GithubAccessType {
		t.Errorf("expected access type %s, got %s", component.GithubAccessType, source.Access.Type)
	}

	if source.Access.RepoUrl != "https://github.com/kyma-project/modulectl" {
		t.Errorf("expected repo URL https://github.com/kyma-project/modulectl, got %s", source.Access.RepoUrl)
	}

	if source.Access.Commit != "abc123def456" {
		t.Errorf("expected commit abc123def456, got %s", source.Access.Commit)
	}
}

func TestConstructor_AddLabel(t *testing.T) {
	constructor := component.NewConstructor("test-component", "1.0.0")

	initialLabelCount := len(constructor.Components[0].Labels)

	constructor.AddLabel("test-key", "test-value", common.VersionV1)

	expectedLabelCount := initialLabelCount + 1
	if len(constructor.Components[0].Labels) != expectedLabelCount {
		t.Fatalf("expected %d labels, got %d", expectedLabelCount, len(constructor.Components[0].Labels))
	}

	var addedLabel *component.Label
	for _, label := range constructor.Components[0].Labels {
		if label.Name == "test-key" {
			addedLabel = &label
			break
		}
	}

	if addedLabel == nil {
		t.Fatal("added label not found")
	}

	if addedLabel.Name != "test-key" {
		t.Errorf("expected label name test-key, got %s", addedLabel.Name)
	}

	if addedLabel.Value != "test-value" {
		t.Errorf("expected label value test-value, got %s", addedLabel.Value)
	}

	if addedLabel.Version != common.VersionV1 {
		t.Errorf("expected label version %s, got %s", common.VersionV1, addedLabel.Version)
	}
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

	expectedCount := len(labels)
	if len(constructor.Components[0].Labels) != expectedCount {
		t.Errorf("expected %d labels, got %d", expectedCount, len(constructor.Components[0].Labels))
	}

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
		if !found {
			t.Errorf("label with key %s not found in component labels", expectedLabel.key)
		}
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
		expectedCount := initialLabelCounts[i] + 1
		if len(source.Labels) != expectedCount {
			t.Errorf("source %d: expected %d labels, got %d", i, expectedCount, len(source.Labels))
		}

		var foundLabel *component.Label
		for _, label := range source.Labels {
			if label.Name == "test-key" {
				foundLabel = &label
				break
			}
		}

		if foundLabel == nil {
			t.Errorf("source %d: added label not found", i)
			continue
		}

		if foundLabel.Value != "test-value" {
			t.Errorf("source %d: expected label value test-value, got %s", i, foundLabel.Value)
		}

		if foundLabel.Version != common.VersionV1 {
			t.Errorf("source %d: expected label version %s, got %s", i, common.VersionV1, foundLabel.Version)
		}
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

	if len(constructor.Components[0].Resources) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(constructor.Components[0].Resources))
	}

	resource := constructor.Components[0].Resources[0]
	if resource.Type != component.OCIArtifactResourceType {
		t.Errorf("expected resource type %s, got %s", component.OCIArtifactResourceType, resource.Type)
	}

	if resource.Relation != component.OCIArtifactResourceRelation {
		t.Errorf("expected resource relation %s, got %s", component.OCIArtifactResourceRelation, resource.Relation)
	}

	if len(resource.Labels) != 1 {
		t.Fatalf("expected 1 label, got %d", len(resource.Labels))
	}

	expectedLabelName := common.SecScanBaseLabelKey + "/" + common.TypeLabelKey
	if resource.Labels[0].Name != expectedLabelName {
		t.Errorf("expected label name %s, got %s", expectedLabelName, resource.Labels[0].Name)
	}

	if resource.Labels[0].Value != common.ThirdPartyImageLabelValue {
		t.Errorf("expected label value %s, got %s", common.ThirdPartyImageLabelValue, resource.Labels[0].Value)
	}

	if resource.Labels[0].Version != common.OCMVersion {
		t.Errorf("expected label version %s, got %s", common.OCMVersion, resource.Labels[0].Version)
	}

	if resource.Access.Type != component.OCIArtifactAccessType {
		t.Errorf("expected access type %s, got %s", component.OCIArtifactAccessType, resource.Access.Type)
	}

	if resource.Access.ImageReference != imageInfo.FullURL {
		t.Errorf("expected image reference %s, got %s", imageInfo.FullURL, resource.Access.ImageReference)
	}
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

	if len(constructor.Components[0].Resources) != 2 {
		t.Fatalf("expected 2 resources, got %d", len(constructor.Components[0].Resources))
	}

	for i, resource := range constructor.Components[0].Resources {
		if resource.Access.ImageReference != imageInfos[i].FullURL {
			t.Errorf("resource %d: expected image reference %s, got %s", i, imageInfos[i].FullURL,
				resource.Access.ImageReference)
		}
	}
}

func TestConstructor_AddRawManifestResource(t *testing.T) {
	constructor := component.NewConstructor("test-component", "1.0.0")

	initialResourceCount := len(constructor.Components[0].Resources)

	constructor.AddRawManifestResource("/path/to/manifest.yaml")

	expectedCount := initialResourceCount + 1
	if len(constructor.Components[0].Resources) != expectedCount {
		t.Fatalf("expected %d resources, got %d", expectedCount, len(constructor.Components[0].Resources))
	}

	resource := constructor.Components[0].Resources[len(constructor.Components[0].Resources)-1]
	if resource.Name != common.RawManifestResourceName {
		t.Errorf("expected resource name %s, got %s", common.RawManifestResourceName, resource.Name)
	}

	if resource.Type != component.DirectoryTreeResourceType {
		t.Errorf("expected resource type %s, got %s", component.DirectoryTreeResourceType, resource.Type)
	}

	if resource.Version != "1.0.0" {
		t.Errorf("expected resource version 1.0.0, got %s", resource.Version)
	}

	if resource.Input.Type != component.DirectoryInputType {
		t.Errorf("expected input type dir, got %s", resource.Input.Type)
	}

	if resource.Input.Path != "/path/to/manifest.yaml" {
		t.Errorf("expected input path /path/to/manifest.yaml, got %s", resource.Input.Path)
	}
}

func TestConstructor_AddDefaultCRResource(t *testing.T) {
	constructor := component.NewConstructor("test-component", "1.0.0")

	initialResourceCount := len(constructor.Components[0].Resources)

	constructor.AddDefaultCRResource("/path/to/defaultcr.yaml")

	expectedCount := initialResourceCount + 1
	if len(constructor.Components[0].Resources) != expectedCount {
		t.Fatalf("expected %d resources, got %d", expectedCount, len(constructor.Components[0].Resources))
	}

	resource := constructor.Components[0].Resources[len(constructor.Components[0].Resources)-1]
	if resource.Name != common.DefaultCRResourceName {
		t.Errorf("expected resource name %s, got %s", common.DefaultCRResourceName, resource.Name)
	}

	if resource.Type != component.DirectoryTreeResourceType {
		t.Errorf("expected resource type %s, got %s", component.DirectoryTreeResourceType, resource.Type)
	}

	if resource.Version != "1.0.0" {
		t.Errorf("expected resource version 1.0.0, got %s", resource.Version)
	}

	if resource.Input.Type != component.DirectoryInputType {
		t.Errorf("expected input type dir, got %s", resource.Input.Type)
	}

	if resource.Input.Path != "/path/to/defaultcr.yaml" {
		t.Errorf("expected input path /path/to/defaultcr.yaml, got %s", resource.Input.Path)
	}
}

func TestConstructor_AddMetadataResource(t *testing.T) {
	constructor := component.NewConstructor("test-component", "1.0.0")

	moduleConfig := &contentprovider.ModuleConfig{
		Name:    "test-module",
		Version: "1.0.0",
	}

	initialResourceCount := len(constructor.Components[0].Resources)

	constructor.AddMetadataResource(moduleConfig)

	if len(constructor.Components[0].Resources) > initialResourceCount {
		resource := constructor.Components[0].Resources[len(constructor.Components[0].Resources)-1]
		if resource.Name != common.MetadataResourceName {
			t.Errorf("expected resource name %s, got %s", common.MetadataResourceName, resource.Name)
		}

		if resource.Type != component.PlainTextResourceType {
			t.Errorf("expected resource type %s, got %s", component.PlainTextResourceType, resource.Type)
		}

		if resource.Version != "1.0.0" {
			t.Errorf("expected resource version 1.0.0, got %s", resource.Version)
		}

		if resource.Input.Type != component.BinaryResourceInput {
			t.Errorf("expected input type %s, got %s", component.BinaryResourceInput, resource.Input.Type)
		}

		if resource.Input.Data == "" {
			t.Error("expected input data to be non-empty")
		}

		_, err := base64.StdEncoding.DecodeString(resource.Input.Data)
		if err != nil {
			t.Errorf("expected input data to be valid base64: %v", err)
		}
	}
}

func TestConstructor_CreateComponentConstructorFile(t *testing.T) {
	constructor := component.NewConstructor("test-component", "1.0.0")

	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "constructor.yaml")

	err := constructor.CreateComponentConstructorFile(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("constructor file was not created")
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read constructor file: %v", err)
	}

	var loadedConstructor component.Constructor
	err = yaml.Unmarshal(content, &loadedConstructor)
	if err != nil {
		t.Fatalf("failed to unmarshal constructor file: %v", err)
	}

	compareTwoConstructors(t, loadedConstructor, constructor)
}

func compareTwoConstructors(t *testing.T, loadedConstructor component.Constructor, constructor *component.Constructor) {
	if len(loadedConstructor.Components) != 1 {
		t.Errorf("expected 1 component in loaded constructor, got %d", len(loadedConstructor.Components))
	}
	if loadedConstructor.Components[0].Name != "test-component" {
		t.Errorf("expected component name test-component, got %s", loadedConstructor.Components[0].Name)
	}
	if loadedConstructor.Components[0].Version != "1.0.0" {
		t.Errorf("expected component version 1.0.0, got %s", loadedConstructor.Components[0].Version)
	}
	if !reflect.DeepEqual(loadedConstructor.Components[0].Provider, constructor.Components[0].Provider) {
		t.Errorf("expected provider %+v, got %+v", constructor.Components[0].Provider,
			loadedConstructor.Components[0].Provider)
	}
	if len(loadedConstructor.Components[0].Resources) != len(constructor.Components[0].Resources) {
		t.Errorf("expected %d resources, got %d", len(constructor.Components[0].Resources),
			len(loadedConstructor.Components[0].Resources))
	}
	if len(loadedConstructor.Components[0].Sources) != len(constructor.Components[0].Sources) {
		t.Errorf("expected %d sources, got %d", len(constructor.Components[0].Sources),
			len(loadedConstructor.Components[0].Sources))
	}
}
