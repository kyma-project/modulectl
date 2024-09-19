package componentdescriptor

import (
	"fmt"

	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"ocm.software/ocm/api/ocm/compdesc"
	ocmv1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
)

const (
	moduleImageResourceName = "module-image"
	rawManifestResourceName = "raw-manifest"
	defaultCRResourceName   = "default-cr"
	ociArtifactType         = "ociArtifact"
	directoryType           = "directory"
	OCIRegistryCredLabel    = "oci-registry-cred" //nolint:gosec // it's a label
)

type Resource struct {
	compdesc.Resource
	Path string
}

func GenerateModuleResources(moduleVersion, manifestPath, defaultCRPath, registryCredSelector string) ([]Resource,
	error,
) {
	moduleImageResource := Resource{
		Resource: compdesc.Resource{
			ResourceMeta: compdesc.ResourceMeta{
				ElementMeta: compdesc.ElementMeta{
					Name: moduleImageResourceName,
				},
				Type:     ociArtifactType,
				Relation: ocmv1.ExternalRelation,
			},
		},
	}

	rawManifestResource := Resource{
		Resource: compdesc.Resource{
			ResourceMeta: compdesc.ResourceMeta{
				ElementMeta: compdesc.ElementMeta{
					Name: rawManifestResourceName,
				},
				Type:     directoryType,
				Relation: ocmv1.LocalRelation,
			},
		},
		Path: manifestPath,
	}
	resources := []Resource{moduleImageResource, rawManifestResource}

	if defaultCRPath != "" {
		defaultCRResource := Resource{
			Resource: compdesc.Resource{
				ResourceMeta: compdesc.ResourceMeta{
					ElementMeta: compdesc.ElementMeta{
						Name: defaultCRResourceName,
					},
					Type:     directoryType,
					Relation: ocmv1.LocalRelation,
				},
			},
			Path: defaultCRPath,
		}
		resources = append(resources, defaultCRResource)
	}

	credentialsLabel, err := createCredMatchLabels(registryCredSelector)
	if err != nil {
		return nil, fmt.Errorf("failed to create credentials label: %w", err)
	}
	for _, r := range resources {
		r.Version = moduleVersion
		r.SetLabels([]ocmv1.Label{
			{
				Name:  OCIRegistryCredLabel,
				Value: credentialsLabel,
			},
		})
	}
	return resources, nil
}

func createCredMatchLabels(registryCredSelector string) ([]byte, error) {
	var matchLabels []byte
	if registryCredSelector != "" {
		selector, err := metav1.ParseToLabelSelector(registryCredSelector)
		if err != nil {
			return nil, fmt.Errorf("failed to parse label selector: %w", err)
		}
		matchLabels, err = yaml.Marshal(selector.MatchLabels)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal match labels: %w", err)
		}
	}
	return matchLabels, nil
}
