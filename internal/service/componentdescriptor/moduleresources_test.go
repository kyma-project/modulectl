//nolint:gosec // some registry var names are used in tests
package componentdescriptor_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ocmv1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"

	"github.com/kyma-project/modulectl/internal/service/componentdescriptor"
)

func TestCreateCredMatchLabels_ReturnCorrectLabels(t *testing.T) {
	registryCredSelector := "operator.kyma-project.io/oci-registry-cred=test-operator"
	label, err := componentdescriptor.CreateCredMatchLabels(registryCredSelector)

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
	_, err := componentdescriptor.CreateCredMatchLabels(registryCredSelector)
	assert.ErrorContains(t, err, "failed to parse label selector")
}

func TestCreateCredMatchLabels_ReturnEmptyLabelWhenEmptySelector(t *testing.T) {
	registryCredSelector := ""
	label, err := componentdescriptor.CreateCredMatchLabels(registryCredSelector)

	require.NoError(t, err)
	assert.Empty(t, label)
}

func TestGenerateModuleResources_ReturnErrorWhenInvalidSelector(t *testing.T) {
	_, err := componentdescriptor.GenerateModuleResources("1.0.0", "path", "path", "@test2")
	assert.ErrorContains(t, err, "failed to create credentials label")
}

func TestGenerateModuleResources_ReturnCorrectResourcesWithDefaultCRPath(t *testing.T) {
	moduleVersion := "1.0.0"
	manifestPath := "path/to/manifest"
	defaultCRPath := "path/to/defaultCR"
	registryCredSelector := "operator.kyma-project.io/oci-registry-cred=test-operator"

	resources, err := componentdescriptor.GenerateModuleResources(moduleVersion, manifestPath, defaultCRPath,
		registryCredSelector)
	require.NoError(t, err)
	require.Len(t, resources, 3)

	require.Equal(t, "module-image", resources[0].Name)
	require.Equal(t, "ociArtifact", resources[0].Type)
	require.Equal(t, ocmv1.ExternalRelation, resources[0].Relation)
	require.Empty(t, resources[0].Path)

	require.Equal(t, "raw-manifest", resources[1].Name)
	require.Equal(t, "directory", resources[1].Type)
	require.Equal(t, ocmv1.LocalRelation, resources[1].Relation)
	require.Equal(t, "path/to/manifest", resources[1].Path)

	require.Equal(t, "default-cr", resources[2].Name)
	require.Equal(t, "directory", resources[2].Type)
	require.Equal(t, ocmv1.LocalRelation, resources[2].Relation)
	require.Equal(t, "path/to/defaultCR", resources[2].Path)

	for _, resource := range resources {
		require.Equal(t, moduleVersion, resource.Version)
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
	moduleVersion := "1.0.0"
	manifestPath := "path/to/manifest"
	registryCredSelector := "operator.kyma-project.io/oci-registry-cred=test-operator"

	resources, err := componentdescriptor.GenerateModuleResources(moduleVersion, manifestPath, "",
		registryCredSelector)
	require.NoError(t, err)
	require.Len(t, resources, 2)

	require.Equal(t, "module-image", resources[0].Name)
	require.Equal(t, "ociArtifact", resources[0].Type)
	require.Equal(t, ocmv1.ExternalRelation, resources[0].Relation)
	require.Empty(t, resources[0].Path)

	require.Equal(t, "raw-manifest", resources[1].Name)
	require.Equal(t, "directory", resources[1].Type)
	require.Equal(t, ocmv1.LocalRelation, resources[1].Relation)
	require.Equal(t, "path/to/manifest", resources[1].Path)

	for _, resource := range resources {
		require.Equal(t, moduleVersion, resource.Version)
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
	moduleVersion := "1.0.0"
	manifestPath := "path/to/manifest"
	defaultCRPath := "path/to/defaultCR"

	resources, err := componentdescriptor.GenerateModuleResources(moduleVersion, manifestPath, defaultCRPath,
		"")
	require.NoError(t, err)
	require.Len(t, resources, 3)

	require.Equal(t, "module-image", resources[0].Name)
	require.Equal(t, "ociArtifact", resources[0].Type)
	require.Equal(t, ocmv1.ExternalRelation, resources[0].Relation)
	require.Empty(t, resources[0].Path)

	require.Equal(t, "raw-manifest", resources[1].Name)
	require.Equal(t, "directory", resources[1].Type)
	require.Equal(t, ocmv1.LocalRelation, resources[1].Relation)
	require.Equal(t, "path/to/manifest", resources[1].Path)

	require.Equal(t, "default-cr", resources[2].Name)
	require.Equal(t, "directory", resources[2].Type)
	require.Equal(t, ocmv1.LocalRelation, resources[2].Relation)
	require.Equal(t, "path/to/defaultCR", resources[2].Path)

	for _, resource := range resources {
		require.Equal(t, moduleVersion, resource.Version)
		require.Empty(t, resource.Labels)
	}
}
