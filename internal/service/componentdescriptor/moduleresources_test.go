package componentdescriptor_test

import (
	"encoding/json"
	"testing"

	"github.com/kyma-project/modulectl/internal/service/componentdescriptor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
}

func TestGenerateModuleResources_ReturnCorrectResourcesWithoutDefaultCRPath(t *testing.T) {

}

func TestGenerateModuleResources_ReturnCorrectResourcesWithNoSelector(t *testing.T) {

}
