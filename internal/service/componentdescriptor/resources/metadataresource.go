package resources

import (
	"errors"
	"fmt"

	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"ocm.software/ocm/api/ocm/compdesc"
	ocmv1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/extensions/artifacttypes"

	"github.com/kyma-project/modulectl/internal/service/componentdescriptor/resources/accesshandler"
	"github.com/kyma-project/modulectl/internal/service/contentprovider"
)

const (
	metadataResourceName = "metadata"
)

var ErrNilModuleConfig = errors.New("module config must not be nil")

type MetadataResource struct {
	Spec struct {
		Mandatory bool                     `yaml:"mandatory,omitempty"`
		Manager   *contentprovider.Manager `yaml:"manager,omitempty"`
		Info      struct {
			Repository    string                `yaml:"repository"`
			Documentation string                `yaml:"documentation"`
			Icons         contentprovider.Icons `yaml:"icons"`
		} `yaml:"info"`
		Resources           contentprovider.Resources  `yaml:"resources,omitempty"`
		AssociatedResources []*metav1.GroupVersionKind `yaml:"associatedResources,omitempty"`
	} `yaml:"spec"`
}

func GenerateMetadataResource(config *contentprovider.ModuleConfig) (Resource, error) {
	if config == nil {
		return Resource{}, ErrNilModuleConfig
	}

	metadataResource := MetadataResource{}
	metadataResource.Spec.Mandatory = config.Mandatory
	metadataResource.Spec.Manager = config.Manager
	metadataResource.Spec.Info.Repository = config.Repository
	metadataResource.Spec.Info.Documentation = config.Documentation
	metadataResource.Spec.Info.Icons = config.Icons
	metadataResource.Spec.Resources = config.Resources
	metadataResource.Spec.AssociatedResources = config.AssociatedResources

	data, err := yaml.Marshal(metadataResource)
	if err != nil {
		return Resource{}, fmt.Errorf("failed to marshal metadata resource: %w", err)
	}

	return Resource{
		Resource: compdesc.Resource{
			ResourceMeta: compdesc.ResourceMeta{
				ElementMeta: compdesc.ElementMeta{
					Name: metadataResourceName,
				},
				Type:     artifacttypes.PLAIN_TEXT,
				Relation: ocmv1.LocalRelation,
			},
		},
		AccessHandler: &accesshandler.Yaml{
			String: string(data),
		},
	}, nil
}
