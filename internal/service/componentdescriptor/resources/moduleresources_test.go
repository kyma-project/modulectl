//nolint:gosec // some registry var names are used in tests
package resources_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ocmv1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"

	"github.com/kyma-project/modulectl/internal/service/componentdescriptor/resources"
	"github.com/kyma-project/modulectl/internal/service/componentdescriptor/resources/accesshandler"
	"github.com/kyma-project/modulectl/internal/service/contentprovider"
)

func TestCreateCredMatchLabels_ReturnCorrectLabels(t *testing.T) {
	registryCredSelector := "operator.kyma-project.io/oci-registry-cred=test-operator"
	label, err := resources.CreateCredMatchLabels(registryCredSelector)

	expectedLabel := map[string]string{
		"operator.kyma-project.io/oci-registry-cred": "test-operator",
	}
	var returnedLabel map[string]string

	require.NoError(t, err)

	err = json.Unmarshal(label, &returnedLabel)
	require.NoError(t, err)
	assert.Equal(t, expectedLabel, returnedLabel)
}

func TestCreateCredMatchLabels_ReturnErrorOnInvalidSelector(t *testing.T) {
	registryCredSelector := "@test2"
	_, err := resources.CreateCredMatchLabels(registryCredSelector)
	assert.ErrorContains(t, err, "failed to parse label selector")
}

func TestCreateCredMatchLabels_ReturnEmptyLabelWhenEmptySelector(t *testing.T) {
	registryCredSelector := ""
	label, err := resources.CreateCredMatchLabels(registryCredSelector)

	require.NoError(t, err)
	assert.Empty(t, label)
}

func TestGenerateModuleResources_ReturnErrorWhenInvalidSelector(t *testing.T) {
	moduleConfig := &contentprovider.ModuleConfig{
		Version: "1.0.0",
	}
	_, err := resources.GenerateModuleResources(moduleConfig, "path", "path", "@test2")
	assert.ErrorContains(t, err, "failed to create credentials label")
}

func TestGenerateModuleResources_ReturnCorrectResourcesWithDefaultCRPath(t *testing.T) {
	moduleConfig := &contentprovider.ModuleConfig{
		Version: "1.0.0",
	}
	manifestPath := "path/to/manifest"
	defaultCRPath := "path/to/defaultCR"
	registryCredSelector := "operator.kyma-project.io/oci-registry-cred=test-operator"

	resources, err := resources.GenerateModuleResources(moduleConfig, manifestPath, defaultCRPath,
		registryCredSelector)
	require.NoError(t, err)
	require.Len(t, resources, 4)

	require.Equal(t, "module-image", resources[0].Name)
	require.Equal(t, "ociArtifact", resources[0].Type)
	require.Equal(t, ocmv1.ExternalRelation, resources[0].Relation)
	require.Nil(t, resources[0].AccessHandler)

	require.Equal(t, "metadata", resources[1].Name)
	require.Equal(t, "plainText", resources[1].Type)
	require.Equal(t, ocmv1.LocalRelation, resources[1].Relation)
	metadataResourceHandler, ok := resources[1].AccessHandler.(*accesshandler.Yaml)
	require.True(t, ok)
	require.NotEmpty(t, metadataResourceHandler.String)

	require.Equal(t, "raw-manifest", resources[2].Name)
	require.Equal(t, "directory", resources[2].Type)
	require.Equal(t, ocmv1.LocalRelation, resources[2].Relation)
	manifestResourceHandler, ok := resources[2].AccessHandler.(*accesshandler.Tar)
	require.True(t, ok)
	require.Equal(t, "path/to/manifest", manifestResourceHandler.Path)

	require.Equal(t, "default-cr", resources[3].Name)
	require.Equal(t, "directory", resources[3].Type)
	require.Equal(t, ocmv1.LocalRelation, resources[3].Relation)
	defaultCRResourceHandler, ok := resources[3].AccessHandler.(*accesshandler.Tar)
	require.True(t, ok)
	require.Equal(t, "path/to/defaultCR", defaultCRResourceHandler.Path)

	for _, resource := range resources {
		require.Equal(t, "1.0.0", resource.Version)
		require.Equal(t, "oci-registry-cred", resource.Labels[0].Name)
		var returnedLabel map[string]string
		err = json.Unmarshal(resource.Labels[0].Value, &returnedLabel)
		require.NoError(t, err)
		expectedLabel := map[string]string{
			"operator.kyma-project.io/oci-registry-cred": "test-operator",
		}
		require.Equal(t, expectedLabel, returnedLabel)
	}
}

func TestGenerateModuleResources_ReturnCorrectResourcesWithoutDefaultCRPath(t *testing.T) {
	moduleConfig := &contentprovider.ModuleConfig{
		Version: "1.0.0",
	}
	manifestPath := "path/to/manifest"
	registryCredSelector := "operator.kyma-project.io/oci-registry-cred=test-operator"

	resources, err := resources.GenerateModuleResources(moduleConfig, manifestPath, "",
		registryCredSelector)
	require.NoError(t, err)
	require.Len(t, resources, 3)

	require.Equal(t, "module-image", resources[0].Name)
	require.Equal(t, "ociArtifact", resources[0].Type)
	require.Equal(t, ocmv1.ExternalRelation, resources[0].Relation)
	require.Nil(t, resources[0].AccessHandler)

	require.Equal(t, "metadata", resources[1].Name)
	require.Equal(t, "plainText", resources[1].Type)
	require.Equal(t, ocmv1.LocalRelation, resources[1].Relation)
	metadataResourceHandler, ok := resources[1].AccessHandler.(*accesshandler.Yaml)
	require.True(t, ok)
	require.NotEmpty(t, metadataResourceHandler.String)

	require.Equal(t, "raw-manifest", resources[2].Name)
	require.Equal(t, "directory", resources[2].Type)
	require.Equal(t, ocmv1.LocalRelation, resources[2].Relation)
	manifestResourceHandler, ok := resources[2].AccessHandler.(*accesshandler.Tar)
	require.True(t, ok)
	require.Equal(t, "path/to/manifest", manifestResourceHandler.Path)

	for _, resource := range resources {
		require.Equal(t, "1.0.0", resource.Version)
		require.Equal(t, "oci-registry-cred", resource.Labels[0].Name)
		var returnedLabel map[string]string
		err = json.Unmarshal(resource.Labels[0].Value, &returnedLabel)
		expectedLabel := map[string]string{
			"operator.kyma-project.io/oci-registry-cred": "test-operator",
		}
		require.NoError(t, err)
		require.Equal(t, expectedLabel, returnedLabel)
	}
}

func TestGenerateModuleResources_ReturnCorrectResourcesWithNoSelector(t *testing.T) {
	moduleConfig := &contentprovider.ModuleConfig{
		Version: "1.0.0",
	}
	manifestPath := "path/to/manifest"
	defaultCRPath := "path/to/defaultCR"

	resources, err := resources.GenerateModuleResources(moduleConfig, manifestPath, defaultCRPath,
		"")
	require.NoError(t, err)
	require.Len(t, resources, 4)

	require.Equal(t, "module-image", resources[0].Name)
	require.Equal(t, "ociArtifact", resources[0].Type)
	require.Equal(t, ocmv1.ExternalRelation, resources[0].Relation)
	require.Nil(t, resources[0].AccessHandler)

	require.Equal(t, "metadata", resources[1].Name)
	require.Equal(t, "plainText", resources[1].Type)
	require.Equal(t, ocmv1.LocalRelation, resources[1].Relation)
	metadataResourceHandler, ok := resources[1].AccessHandler.(*accesshandler.Yaml)
	require.True(t, ok)
	require.NotEmpty(t, metadataResourceHandler.String)

	require.Equal(t, "raw-manifest", resources[2].Name)
	require.Equal(t, "directory", resources[2].Type)
	require.Equal(t, ocmv1.LocalRelation, resources[2].Relation)
	manifestResourceHandler, ok := resources[2].AccessHandler.(*accesshandler.Tar)
	require.True(t, ok)
	require.Equal(t, "path/to/manifest", manifestResourceHandler.Path)

	require.Equal(t, "default-cr", resources[3].Name)
	require.Equal(t, "directory", resources[3].Type)
	require.Equal(t, ocmv1.LocalRelation, resources[3].Relation)
	defaultCRResourceHandler, ok := resources[3].AccessHandler.(*accesshandler.Tar)
	require.True(t, ok)
	require.Equal(t, "path/to/defaultCR", defaultCRResourceHandler.Path)

	for _, resource := range resources {
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
		manifestPath := "test/path"
		resource := resources.GenerateRawManifestResource(manifestPath)
		require.Equal(t, "raw-manifest", resource.Name)
		require.Equal(t, "directory", resource.Type)
		require.Equal(t, ocmv1.LocalRelation, resource.Relation)

		handler, ok := resource.AccessHandler.(*accesshandler.Tar)
		require.True(t, ok)
		require.Equal(t, manifestPath, handler.Path)
	})

	t.Run("default CR resource", func(t *testing.T) {
		crPath := "test/cr/path"
		resource := resources.GenerateDefaultCRResource(crPath)
		require.Equal(t, "default-cr", resource.Name)
		require.Equal(t, "directory", resource.Type)
		require.Equal(t, ocmv1.LocalRelation, resource.Relation)

		handler, ok := resource.AccessHandler.(*accesshandler.Tar)
		require.True(t, ok)
		require.Equal(t, crPath, handler.Path)
	})
}
