package create_test

import (
	"os"
	"testing"

	"github.com/kyma-project/lifecycle-manager/api/shared"
	"github.com/kyma-project/lifecycle-manager/api/v1beta2"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/ocireg"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/github"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/localblob"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/ociartifact"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	ocmmetav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	compdescv2 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/versions/v2"
	ocmocireg "github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ocireg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/util/yaml"
)

const (
	rawManifestLayerName = "raw-manifest"
	typeYaml             = "yaml"
)

func Test_ModuleTemplate(t *testing.T) {
	// TODO remove debugging env
	// err := os.Setenv("OCI_REPOSITORY_URL", "http://k3d-oci.localhost:5001")
	// err = os.Setenv("TEST_REPOSITORY_URL", "https://github.com/lindnerby/template-operator.git")
	// err = os.Setenv("MODULE_TEMPLATE_PATH", "/tmp/module-config-template.yaml")
	// err = os.Setenv("MODULE_TEMPLATE_VERSION", "1.0.0")

	ociRepoURL := os.Getenv("OCI_REPOSITORY_URL")
	testRepoURL := os.Getenv("TEST_REPOSITORY_URL")
	templatePath := os.Getenv("MODULE_TEMPLATE_PATH")

	template, err := readModuleTemplate(templatePath)
	require.NoError(t, err)
	descriptor := getDescriptor(template)
	assert.NotNil(t, descriptor)
	assert.Equal(t, compdescv2.SchemaVersion, descriptor.SchemaVersion())

	t.Run("test annotations", func(t *testing.T) {
		annotations := template.Annotations
		expectedModuleTemplateVersion := os.Getenv("MODULE_TEMPLATE_VERSION")
		assert.Equal(t, expectedModuleTemplateVersion, annotations[shared.ModuleVersionAnnotation])
		assert.Equal(t, "false", annotations[shared.IsClusterScopedAnnotation])
	})

	t.Run("test descriptor.component.repositoryContexts", func(t *testing.T) {
		assert.Len(t, 1, len(descriptor.RepositoryContexts))
		repo := descriptor.GetEffectiveRepositoryContext()
		assert.Equal(t, ociRepoURL, repo.Object["baseUrl"])
		assert.Equal(t, string(ocmocireg.OCIRegistryURLPathMapping), repo.Object["componentNameMapping"])
		assert.Equal(t, ocireg.Type, repo.Object["type"])
	})

	t.Run("test descriptor.component.resources", func(t *testing.T) {
		assert.Len(t, 2, len(descriptor.Resources))

		resource := descriptor.Resources[0]
		assert.Equal(t, "template-operator", resource.Name)
		assert.Equal(t, ocmmetav1.ExternalRelation, resource.Relation)
		assert.Equal(t, "ociImage", resource.Type)
		expectedModuleTemplateVersion := os.Getenv("MODULE_TEMPLATE_VERSION")
		assert.Equal(t, expectedModuleTemplateVersion, resource.Version)

		resource = descriptor.Resources[1]
		assert.Equal(t, rawManifestLayerName, resource.Name)
		assert.Equal(t, ocmmetav1.LocalRelation, resource.Relation)
		assert.Equal(t, typeYaml, resource.Type)
		assert.Equal(t, expectedModuleTemplateVersion, resource.Version)
	})

	t.Run("test descriptor.component.resources[0].access", func(t *testing.T) {
		resourceAccessSpec, err := ocm.DefaultContext().AccessSpecForSpec(descriptor.Resources[0].Access)
		require.NoError(t, err)
		ociArtifactAccessSpec, ok := resourceAccessSpec.(*ociartifact.AccessSpec)
		assert.True(t, ok)
		assert.Equal(t, ociartifact.Type, ociArtifactAccessSpec.GetType())
		assert.Equal(t, "europe-docker.pkg.dev/kyma-project/prod/template-operator:1.0.0",
			ociArtifactAccessSpec.ImageReference)
	})

	t.Run("test descriptor.component.resources[1].access", func(t *testing.T) {
		resourceAccessSpec, err := ocm.DefaultContext().AccessSpecForSpec(descriptor.Resources[1].Access)
		require.NoError(t, err)
		localBlobAccessSpec, ok := resourceAccessSpec.(*localblob.AccessSpec)
		assert.True(t, ok)
		assert.Equal(t, localblob.Type, localBlobAccessSpec.GetType())
		assert.Contains(t, localBlobAccessSpec.LocalReference, "sha256:")
	})

	t.Run("test descriptor.component.sources", func(t *testing.T) {
		assert.Len(t, len(descriptor.Sources), 1)
		source := descriptor.Sources[0]
		sourceAccessSpec, err := ocm.DefaultContext().AccessSpecForSpec(source.Access)
		require.NoError(t, err)
		githubAccessSpec, ok := sourceAccessSpec.(*github.AccessSpec)
		assert.True(t, ok)
		assert.Equal(t, github.Type, githubAccessSpec.Type)
		assert.Contains(t, testRepoURL, githubAccessSpec.RepoURL)
	})

	t.Run("test spec.mandatory", func(t *testing.T) {
		assert.False(t, template.Spec.Mandatory)
	})

	t.Run("test security scan labels", func(t *testing.T) {
		secScanLabels := descriptor.Sources[0].Labels
		flattenedLabels := flatten(secScanLabels)
		assert.Equal(t, map[string]string{
			"git.kyma-project.io/ref":                   "refs/heads/main",
			"scan.security.kyma-project.io/rc-tag":      "1.0.0",
			"scan.security.kyma-project.io/language":    "golang-mod",
			"scan.security.kyma-project.io/dev-branch":  "main",
			"scan.security.kyma-project.io/subprojects": "false",
			"scan.security.kyma-project.io/exclude":     "**/test/**,**/*_test.go",
		}, flattenedLabels)
	})
}

func readModuleTemplate(filepath string) (*v1beta2.ModuleTemplate, error) {
	moduleTemplate := &v1beta2.ModuleTemplate{}
	moduleFile, err := os.ReadFile(filepath)
	if err != nil && len(moduleFile) > 0 {
		return nil, err
	}
	err = yaml.Unmarshal(moduleFile, moduleTemplate)
	if err != nil {
		return nil, err
	}
	return moduleTemplate, err
}

func getDescriptor(template *v1beta2.ModuleTemplate) *v1beta2.Descriptor {
	if template.Spec.Descriptor.Object != nil {
		desc, ok := template.Spec.Descriptor.Object.(*v1beta2.Descriptor)
		if !ok || desc == nil {
			return nil
		}
		return desc
	}
	ocmDesc, err := compdesc.Decode(
		template.Spec.Descriptor.Raw,
		[]compdesc.DecodeOption{compdesc.DisableValidation(true)}...)
	if err != nil {
		return nil
	}
	template.Spec.Descriptor.Object = &v1beta2.Descriptor{ComponentDescriptor: ocmDesc}
	desc, ok := template.Spec.Descriptor.Object.(*v1beta2.Descriptor)
	if !ok {
		return nil
	}

	return desc
}

func flatten(labels ocmmetav1.Labels) map[string]string {
	labelsMap := make(map[string]string)
	for _, l := range labels {
		var value string
		_ = yaml.Unmarshal(l.Value, &value)
		labelsMap[l.Name] = value
	}

	return labelsMap
}
