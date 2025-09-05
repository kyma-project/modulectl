package componentconstructor_test

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kyma-project/modulectl/internal/common"
	"github.com/kyma-project/modulectl/internal/common/types/component"
	"github.com/kyma-project/modulectl/internal/service/componentconstructor"
	"github.com/kyma-project/modulectl/internal/service/contentprovider"
	iotools "github.com/kyma-project/modulectl/tools/io"
)

const (
	testModuleName     = "test-module"
	testModuleVersion  = "1.0.0"
	testManifestPath   = "/path/to/manifest.yaml"
	testDefaultCRPath  = "/path/to/defaultcr.yaml"
	testOutputFileName = "output.yaml"
)

func TestService_AddResourcesAndCreateConstructorFile_Success(t *testing.T) {
	service := componentconstructor.NewService()

	constructor := component.NewConstructor(testModuleName, testModuleVersion)

	moduleConfig := &contentprovider.ModuleConfig{
		Name:    testModuleName,
		Version: testModuleVersion,
	}

	var outputBuffer bytes.Buffer
	cmdOutput := iotools.NewDefaultOut(&outputBuffer)

	tempDir := t.TempDir()
	outputFile := filepath.Join(tempDir, testOutputFileName)

	err := service.AddResourcesAndCreateConstructorFile(
		constructor,
		moduleConfig,
		testManifestPath,
		"",
		cmdOutput,
		outputFile,
	)

	require.NoError(t, err)
	require.FileExists(t, outputFile)
	require.Len(t, constructor.Components[0].Resources, 2)

	resources := constructor.Components[0].Resources
	resourceNames := make([]string, len(resources))
	for i, resource := range resources {
		resourceNames[i] = resource.Name
	}
	require.Contains(t, resourceNames, common.RawManifestResourceName)
	require.Contains(t, resourceNames, common.MetadataResourceName)
}

func TestService_AddResourcesAndCreateConstructorFile_WithDefaultCR(t *testing.T) {
	service := componentconstructor.NewService()

	constructor := component.NewConstructor(testModuleName, testModuleVersion)

	moduleConfig := &contentprovider.ModuleConfig{
		Name:    testModuleName,
		Version: testModuleVersion,
	}

	var outputBuffer bytes.Buffer
	cmdOutput := iotools.NewDefaultOut(&outputBuffer)

	tempDir := t.TempDir()
	outputFile := filepath.Join(tempDir, testOutputFileName)

	err := service.AddResourcesAndCreateConstructorFile(
		constructor,
		moduleConfig,
		testManifestPath,
		testDefaultCRPath,
		cmdOutput,
		outputFile,
	)

	require.NoError(t, err)
	require.FileExists(t, outputFile)
	require.Len(t, constructor.Components[0].Resources, 3)

	resources := constructor.Components[0].Resources
	resourceNames := make([]string, len(resources))
	for i, resource := range resources {
		resourceNames[i] = resource.Name
	}
	require.Contains(t, resourceNames, common.RawManifestResourceName)
	require.Contains(t, resourceNames, common.MetadataResourceName)
	require.Contains(t, resourceNames, common.DefaultCRResourceName)
}

func TestService_AddResourcesAndCreateConstructorFile_ReturnsError_WhenOutputPathInvalid(t *testing.T) {
	const invalidOutputPath = "/invalid/path/that/does/not/exist/output.yaml"

	service := componentconstructor.NewService()

	constructor := component.NewConstructor(testModuleName, testModuleVersion)

	moduleConfig := &contentprovider.ModuleConfig{
		Name:    testModuleName,
		Version: testModuleVersion,
	}

	var outputBuffer bytes.Buffer
	cmdOutput := iotools.NewDefaultOut(&outputBuffer)

	err := service.AddResourcesAndCreateConstructorFile(
		constructor,
		moduleConfig,
		testManifestPath,
		testDefaultCRPath,
		cmdOutput,
		invalidOutputPath,
	)

	require.Error(t, err)
}

func TestService_AddResourcesAndCreateConstructorFile_ReturnsError_WhenConfigNil(t *testing.T) {
	service := componentconstructor.NewService()

	constructor := component.NewConstructor(testModuleName, testModuleVersion)

	var outputBuffer bytes.Buffer
	cmdOutput := iotools.NewDefaultOut(&outputBuffer)

	tempDir := t.TempDir()
	outputFile := filepath.Join(tempDir, testOutputFileName)

	err := service.AddResourcesAndCreateConstructorFile(
		constructor,
		nil,
		testManifestPath,
		testDefaultCRPath,
		cmdOutput,
		outputFile,
	)

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to add metadata resource")
}

func TestService_AddImagesToConstructor_Success(t *testing.T) {
	service := componentconstructor.NewService()
	constructor := component.NewConstructor(testModuleName, testModuleVersion)

	images := []string{
		"ghcr.io/example/image:v1.0.0",
		"docker.io/library/nginx:1.21.0",
		"registry.k8s.io/pause:3.7@sha256:bb1c58b0e4cb9f8e0e7b1c84f8d8d7c8a7a3a1e1e1e1e1e1e1e1e1e1e1e1e1e1",
	}

	err := service.AddImagesToConstructor(constructor, images)

	require.NoError(t, err)

	resources := constructor.Components[0].Resources
	imageResourceCount := 0
	for _, resource := range resources {
		if resource.Type == component.OCIArtifactResourceType && resource.Relation == component.OCIArtifactResourceRelation {
			imageResourceCount++
			require.NotEmpty(t, resource.Name)
			require.NotEmpty(t, resource.Version)
			require.NotEmpty(t, resource.Access)
			require.Equal(t, component.OCIArtifactAccessType, resource.Access.Type)
			require.NotEmpty(t, resource.Access.ImageReference)
			require.Len(t, resource.Labels, 1)
			require.Equal(t, common.ThirdPartyImageLabelValue, resource.Labels[0].Value)
		}
	}
	require.Equal(t, len(images), imageResourceCount)
}

func TestService_AddImagesToConstructor_EmptyImages(t *testing.T) {
	service := componentconstructor.NewService()
	constructor := component.NewConstructor(testModuleName, testModuleVersion)

	images := []string{}

	err := service.AddImagesToConstructor(constructor, images)

	require.NoError(t, err)

	resources := constructor.Components[0].Resources
	imageResourceCount := 0
	for _, resource := range resources {
		if resource.Type == component.OCIArtifactResourceType && resource.Relation == component.OCIArtifactResourceRelation {
			imageResourceCount++
		}
	}
	require.Equal(t, 0, imageResourceCount)
}

func TestService_AddImagesToConstructor_InvalidImage(t *testing.T) {
	service := componentconstructor.NewService()
	constructor := component.NewConstructor(testModuleName, testModuleVersion)

	images := []string{
		"ghcr.io/example/image:v1.0.0",
		"invalid-image",
		"docker.io/library/nginx:1.21.0",
	}

	err := service.AddImagesToConstructor(constructor, images)

	require.Error(t, err)
	require.Contains(t, err.Error(), "image validation failed for invalid-image")
}

func TestService_AddImagesToConstructor_ImageWithLatestTag(t *testing.T) {
	service := componentconstructor.NewService()
	constructor := component.NewConstructor(testModuleName, testModuleVersion)

	images := []string{
		"ghcr.io/example/image:latest",
	}

	err := service.AddImagesToConstructor(constructor, images)

	require.Error(t, err)
	require.Contains(t, err.Error(), "image validation failed")
	require.Contains(t, err.Error(), "image tag is disallowed")
}

func TestService_AddImagesToConstructor_ImageWithMainTag(t *testing.T) {
	service := componentconstructor.NewService()
	constructor := component.NewConstructor(testModuleName, testModuleVersion)

	images := []string{
		"ghcr.io/example/image:main",
	}

	err := service.AddImagesToConstructor(constructor, images)

	require.Error(t, err)
	require.Contains(t, err.Error(), "image validation failed")
	require.Contains(t, err.Error(), "image tag is disallowed")
}

func TestService_AddImagesToConstructor_EmptyImageURL(t *testing.T) {
	service := componentconstructor.NewService()
	constructor := component.NewConstructor(testModuleName, testModuleVersion)

	images := []string{
		"ghcr.io/example/image:v1.0.0",
		"",
		"docker.io/library/nginx:1.21.0",
	}

	err := service.AddImagesToConstructor(constructor, images)

	require.Error(t, err)
	require.Contains(t, err.Error(), "image validation failed")
	require.Contains(t, err.Error(), "empty image URL")
}

func TestService_AddImagesToConstructor_ImageWithoutTag(t *testing.T) {
	service := componentconstructor.NewService()
	constructor := component.NewConstructor(testModuleName, testModuleVersion)

	images := []string{
		"ghcr.io/example/image",
	}

	err := service.AddImagesToConstructor(constructor, images)

	require.Error(t, err)
	require.Contains(t, err.Error(), "image validation failed")
	require.Contains(t, err.Error(), "no tag or digest found")
}

func TestService_AddImagesToConstructor_SingleImage(t *testing.T) {
	service := componentconstructor.NewService()
	constructor := component.NewConstructor(testModuleName, testModuleVersion)

	images := []string{
		"ghcr.io/example/test-image:v2.1.3",
	}

	err := service.AddImagesToConstructor(constructor, images)

	require.NoError(t, err)

	resources := constructor.Components[0].Resources
	imageResources := []component.Resource{}
	for _, resource := range resources {
		if resource.Type == component.OCIArtifactResourceType && resource.Relation == component.OCIArtifactResourceRelation {
			imageResources = append(imageResources, resource)
		}
	}
	require.Len(t, imageResources, 1)

	imageResource := imageResources[0]
	require.Equal(t, component.OCIArtifactResourceType, imageResource.Type)
	require.Equal(t, component.OCIArtifactResourceRelation, imageResource.Relation)
	require.Equal(t, "ghcr.io/example/test-image:v2.1.3", imageResource.Access.ImageReference)
	require.Equal(t, component.OCIArtifactAccessType, imageResource.Access.Type)
	require.NotEmpty(t, imageResource.Name)
	require.NotEmpty(t, imageResource.Version)
}
