package resources

import (
	"errors"

	"github.com/kyma-project/modulectl/internal/service/contentprovider"
	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func GenerateMetadataYaml(config *contentprovider.ModuleConfig) ([]byte, error) {
	if config == nil {
		return []byte{}, ErrNilModuleConfig
	}
	metadataResource := MetadataResource{}
	metadataResource.Spec.Mandatory = config.Mandatory
	metadataResource.Spec.Manager = config.Manager
	metadataResource.Spec.Info.Repository = config.Repository
	metadataResource.Spec.Info.Documentation = config.Documentation
	metadataResource.Spec.Info.Icons = config.Icons
	metadataResource.Spec.Resources = config.Resources
	metadataResource.Spec.AssociatedResources = config.AssociatedResources

	return yaml.Marshal(metadataResource)
}
