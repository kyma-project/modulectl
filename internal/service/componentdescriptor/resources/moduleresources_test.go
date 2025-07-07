package resources_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	ocmv1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"

	"github.com/kyma-project/modulectl/internal/service/componentdescriptor/resources"
	"github.com/kyma-project/modulectl/internal/service/componentdescriptor/resources/accesshandler"
	"github.com/kyma-project/modulectl/internal/service/contentprovider"
)

func TestModuleResourceService_ReturnErrorWhenFileSystemNil(t *testing.T) {
	_, err := resources.NewService(nil)
	require.ErrorIs(t, err, resources.ErrNilTarGenerator)
}

func TestGenerateModuleResources_ReturnCorrectResourcesWithDefaultCRPath(t *testing.T) {
	moduleConfig := &contentprovider.ModuleConfig{
		Version: "1.0.0",
	}
	mockFs := &fileSystemStub{}
	moduleResourceService, err := resources.NewService(mockFs)
	require.NoError(t, err)
	manifestPath := "path/to/manifest"
	defaultCRPath := "path/to/defaultCR"

	res, err := moduleResourceService.GenerateModuleResources(moduleConfig, manifestPath, defaultCRPath)
	require.NoError(t, err)
	require.Len(t, res, 4)

	require.Equal(t, "module-image", res[0].Name)
	require.Equal(t, "ociArtifact", res[0].Type)
	require.Equal(t, ocmv1.ExternalRelation, res[0].Relation)
	require.Nil(t, res[0].AccessHandler)

	require.Equal(t, "metadata", res[1].Name)
	require.Equal(t, "plainText", res[1].Type)
	require.Equal(t, ocmv1.LocalRelation, res[1].Relation)
	metadataResourceHandler, ok := res[1].AccessHandler.(*accesshandler.Yaml)
	require.True(t, ok)
	require.NotEmpty(t, metadataResourceHandler.String)

	require.Equal(t, "raw-manifest", res[2].Name)
	require.Equal(t, "directoryTree", res[2].Type)
	require.Equal(t, ocmv1.LocalRelation, res[2].Relation)
	manifestResourceHandler, ok := res[2].AccessHandler.(*accesshandler.Tar)
	require.True(t, ok)
	require.Equal(t, "path/to/manifest", manifestResourceHandler.GetPath())

	require.Equal(t, "default-cr", res[3].Name)
	require.Equal(t, "directoryTree", res[3].Type)
	require.Equal(t, ocmv1.LocalRelation, res[3].Relation)
	defaultCRResourceHandler, ok := res[3].AccessHandler.(*accesshandler.Tar)
	require.True(t, ok)
	require.Equal(t, "path/to/defaultCR", defaultCRResourceHandler.GetPath())

	for _, resource := range res {
		require.Equal(t, "1.0.0", resource.Version)
	}
}

func TestGenerateModuleResources_ReturnCorrectResourcesWithoutDefaultCRPath(t *testing.T) {
	moduleConfig := &contentprovider.ModuleConfig{
		Version: "1.0.0",
	}
	mockFs := &fileSystemStub{}
	moduleResourceService, err := resources.NewService(mockFs)
	require.NoError(t, err)
	manifestPath := "path/to/manifest"

	res, err := moduleResourceService.GenerateModuleResources(moduleConfig, manifestPath, "")
	require.NoError(t, err)
	require.Len(t, res, 3)

	require.Equal(t, "module-image", res[0].Name)
	require.Equal(t, "ociArtifact", res[0].Type)
	require.Equal(t, ocmv1.ExternalRelation, res[0].Relation)
	require.Nil(t, res[0].AccessHandler)

	require.Equal(t, "metadata", res[1].Name)
	require.Equal(t, "plainText", res[1].Type)
	require.Equal(t, ocmv1.LocalRelation, res[1].Relation)
	metadataResourceHandler, ok := res[1].AccessHandler.(*accesshandler.Yaml)
	require.True(t, ok)
	require.NotEmpty(t, metadataResourceHandler.String)

	require.Equal(t, "raw-manifest", res[2].Name)
	require.Equal(t, "directoryTree", res[2].Type)
	require.Equal(t, ocmv1.LocalRelation, res[2].Relation)
	manifestResourceHandler, ok := res[2].AccessHandler.(*accesshandler.Tar)
	require.True(t, ok)
	require.Equal(t, "path/to/manifest", manifestResourceHandler.GetPath())

	for _, resource := range res {
		require.Equal(t, "1.0.0", resource.Version)
	}
}

func TestGenerateModuleResources_ReturnCorrectResources(t *testing.T) {
	moduleConfig := &contentprovider.ModuleConfig{
		Version: "1.0.0",
	}
	mockFs := &fileSystemStub{}
	moduleResourceService, err := resources.NewService(mockFs)
	require.NoError(t, err)
	manifestPath := "path/to/manifest"
	defaultCRPath := "path/to/defaultCR"

	res, err := moduleResourceService.GenerateModuleResources(moduleConfig, manifestPath, defaultCRPath)
	require.NoError(t, err)
	require.Len(t, res, 4)

	require.Equal(t, "module-image", res[0].Name)
	require.Equal(t, "ociArtifact", res[0].Type)
	require.Equal(t, ocmv1.ExternalRelation, res[0].Relation)
	require.Nil(t, res[0].AccessHandler)

	require.Equal(t, "metadata", res[1].Name)
	require.Equal(t, "plainText", res[1].Type)
	require.Equal(t, ocmv1.LocalRelation, res[1].Relation)
	metadataResourceHandler, ok := res[1].AccessHandler.(*accesshandler.Yaml)
	require.True(t, ok)
	require.NotEmpty(t, metadataResourceHandler.String)

	require.Equal(t, "raw-manifest", res[2].Name)
	require.Equal(t, "directoryTree", res[2].Type)
	require.Equal(t, ocmv1.LocalRelation, res[2].Relation)
	manifestResourceHandler, ok := res[2].AccessHandler.(*accesshandler.Tar)
	require.True(t, ok)
	require.Equal(t, "path/to/manifest", manifestResourceHandler.GetPath())

	require.Equal(t, "default-cr", res[3].Name)
	require.Equal(t, "directoryTree", res[3].Type)
	require.Equal(t, ocmv1.LocalRelation, res[3].Relation)
	defaultCRResourceHandler, ok := res[3].AccessHandler.(*accesshandler.Tar)
	require.True(t, ok)
	require.Equal(t, "path/to/defaultCR", defaultCRResourceHandler.GetPath())

	for _, resource := range res {
		require.Equal(t, "1.0.0", resource.Version)
		require.Empty(t, resource.Labels)
	}
}

func TestResourceGenerators(t *testing.T) {
	t.Run("module image resource", func(t *testing.T) {
		resource := resources.GenerateModuleImageResource()
		require.Equal(t, "module-image", resource.Name)
		require.Equal(t, "ociArtifact", resource.Type)
		require.Equal(t, ocmv1.ExternalRelation, resource.Relation)
		require.Nil(t, resource.AccessHandler)
	})

	t.Run("raw manifest resource", func(t *testing.T) {
		mockFs := &fileSystemStub{}
		manifestPath := "test/path"
		resource := resources.GenerateRawManifestResource(mockFs, manifestPath)
		require.Equal(t, "raw-manifest", resource.Name)
		require.Equal(t, "directoryTree", resource.Type)
		require.Equal(t, ocmv1.LocalRelation, resource.Relation)

		handler, ok := resource.AccessHandler.(*accesshandler.Tar)
		require.True(t, ok)
		require.Equal(t, manifestPath, handler.GetPath())
	})

	t.Run("default CR resource", func(t *testing.T) {
		mockFs := &fileSystemStub{}
		crPath := "test/cr/path"
		resource := resources.GenerateDefaultCRResource(mockFs, crPath)
		require.Equal(t, "default-cr", resource.Name)
		require.Equal(t, "directoryTree", resource.Type)
		require.Equal(t, ocmv1.LocalRelation, resource.Relation)

		handler, ok := resource.AccessHandler.(*accesshandler.Tar)
		require.True(t, ok)
		require.Equal(t, crPath, handler.GetPath())
	})
}

type fileSystemStub struct{}

func (m fileSystemStub) ArchiveFile(_ string) ([]byte, error) {
	return nil, nil
}
