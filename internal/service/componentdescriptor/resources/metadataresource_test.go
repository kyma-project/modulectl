package resources_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ocmv1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"

	"github.com/kyma-project/modulectl/internal/service/componentdescriptor/resources"
	"github.com/kyma-project/modulectl/internal/service/componentdescriptor/resources/accesshandler"
	"github.com/kyma-project/modulectl/internal/service/contentprovider"
)

func TestGenerateMetadataResource(t *testing.T) {
	t.Run("should return error when module config is nil", func(t *testing.T) {
		// when
		resource, err := resources.GenerateMetadataResource(nil)

		// then
		require.Error(t, err)
		require.Equal(t, resources.ErrNilModuleConfig, err)
		require.Equal(t, resources.Resource{}, resource)
	})

	t.Run("should generate metadata resource with all fields", func(t *testing.T) {
		// given
		config := &contentprovider.ModuleConfig{
			Mandatory: true,
			Manager: &contentprovider.Manager{
				Name:      "test-manager",
				Namespace: "kyma-system",
				GroupVersionKind: metav1.GroupVersionKind{
					Group:   "apps",
					Version: "v1",
					Kind:    "Deployment",
				},
			},
			Repository:    "https://github.com/test/repo",
			Documentation: "https://test.docs",
			Icons: contentprovider.Icons{
				"module-icon": "https://test.docs/icon.png",
			},
			Resources: contentprovider.Resources{
				"rawManifest": "https://github.com/test/repo/releases/download/1.0.0/raw-manifest.yaml",
			},
			AssociatedResources: []*metav1.GroupVersionKind{
				{
					Group:   "test.group",
					Version: "v1",
					Kind:    "TestKind",
				},
			},
		}

		// when
		resource, err := resources.GenerateMetadataResource(config)

		// then
		require.NoError(t, err)
		assert.Equal(t, "metadata", resource.Name)
		assert.Equal(t, "plainText", resource.Type)
		assert.Equal(t, ocmv1.LocalRelation, resource.Relation)

		yamlHandler, ok := resource.AccessHandler.(*accesshandler.Yaml)
		require.True(t, ok)
		require.Contains(t, yamlHandler.String, "mandatory: true")
		require.Contains(t, yamlHandler.String, "manager:")
		require.Contains(t, yamlHandler.String, "name: test-manager")
		require.Contains(t, yamlHandler.String, "namespace: kyma-system")
		require.Contains(t, yamlHandler.String, "group: apps")
		require.Contains(t, yamlHandler.String, "version: v1")
		require.Contains(t, yamlHandler.String, "kind: Deployment")
		require.Contains(t, yamlHandler.String, "info:")
		require.Contains(t, yamlHandler.String, "repository: https://github.com/test/repo")
		require.Contains(t, yamlHandler.String, "documentation: https://test.docs")
		require.Contains(t, yamlHandler.String, "module-icon: https://test.docs/icon.png")
		require.Contains(t, yamlHandler.String, "resources:")
		require.Contains(t, yamlHandler.String,
			"rawManifest: https://github.com/test/repo/releases/download/1.0.0/raw-manifest.yaml")
		require.Contains(t, yamlHandler.String, "associatedResources:")
		require.Contains(t, yamlHandler.String, "group: test.group")
		require.Contains(t, yamlHandler.String, "version: v1")
		require.Contains(t, yamlHandler.String, "kind: TestKind")
	})

	t.Run("should generate metadata resource without optional fields when not provided", func(t *testing.T) {
		// given
		config := &contentprovider.ModuleConfig{
			Repository:    "https://github.com/test/repo",
			Documentation: "https://test.docs",
			Icons: contentprovider.Icons{
				"module-icon": "https://test.docs/icon.png",
			},
		}
		// when
		resource, err := resources.GenerateMetadataResource(config)
		// then
		require.NoError(t, err)
		assert.Equal(t, "metadata", resource.Name)
		assert.Equal(t, "plainText", resource.Type)
		assert.Equal(t, ocmv1.LocalRelation, resource.Relation)

		yamlHandler, ok := resource.AccessHandler.(*accesshandler.Yaml)
		require.True(t, ok)
		require.NotContains(t, yamlHandler.String, "mandatory:")
		require.NotContains(t, yamlHandler.String, "manager:")
		require.NotContains(t, yamlHandler.String, "resources:")
		require.NotContains(t, yamlHandler.String, "associatedResources:")
	})
}
