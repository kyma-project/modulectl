package contentprovider

import (
	"errors"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	commonerrors "github.com/kyma-project/modulectl/internal/common/errors"
	"github.com/kyma-project/modulectl/internal/common/types"
)

var ErrDuplicateMapEntries = errors.New("map contains duplicate entries")

type ModuleConfigProvider struct {
	yamlConverter ObjectToYAMLConverter
}

func NewModuleConfigProvider(yamlConverter ObjectToYAMLConverter) (*ModuleConfigProvider, error) {
	if yamlConverter == nil {
		return nil, fmt.Errorf("yamlConverter must not be nil: %w", commonerrors.ErrInvalidArg)
	}

	return &ModuleConfigProvider{
		yamlConverter: yamlConverter,
	}, nil
}

func (s *ModuleConfigProvider) GetDefaultContent(args types.KeyValueArgs) (string, error) {
	if err := s.validateArgs(args); err != nil {
		return "", err
	}

	moduleConfig := s.getModuleConfig(args)

	return s.yamlConverter.ConvertToYaml(moduleConfig), nil
}

func (s *ModuleConfigProvider) getModuleConfig(args types.KeyValueArgs) ModuleConfig {
	return ModuleConfig{
		Name:      args[ArgModuleName],
		Version:   args[ArgModuleVersion],
		Manifest:  args[ArgManifestFile],
		Security:  args[ArgSecurityConfigFile],
		DefaultCR: args[ArgDefaultCRFile],
	}
}

func (s *ModuleConfigProvider) validateArgs(args types.KeyValueArgs) error {
	if args == nil {
		return fmt.Errorf("args must not be nil: %w", ErrInvalidArg)
	}

	if value, ok := args[ArgModuleName]; !ok {
		return fmt.Errorf("%s: %w", ArgModuleName, ErrMissingArg)
	} else if value == "" {
		return fmt.Errorf("%s: %w ,must not be empty", ArgModuleName, ErrInvalidArg)
	}

	if value, ok := args[ArgModuleVersion]; !ok {
		return fmt.Errorf("%s: %w", ArgModuleVersion, ErrMissingArg)
	} else if value == "" {
		return fmt.Errorf("%s: %w ,must not be empty", ArgModuleVersion, ErrInvalidArg)
	}

	return nil
}

type ModuleConfig struct {
	Name                string                     `yaml:"name" comment:"required, the name of the Module"`
	Version             string                     `yaml:"version" comment:"required, the version of the Module"`
	Manifest            string                     `yaml:"manifest" comment:"required, relative path or remote URL to the manifests"`
	Repository          string                     `yaml:"repository" comment:"required, link to the repository"`
	Documentation       string                     `yaml:"documentation" comment:"required, link to documentation"`
	Icons               Icons                      `yaml:"icons,omitempty" comment:"required, list of icons to represent the module in the UI"`
	Mandatory           bool                       `yaml:"mandatory" comment:"optional, default=false, indicates whether the module is mandatory to be installed on all clusters"`
	DefaultCR           string                     `yaml:"defaultCR" comment:"optional, relative path or remote URL to a YAML file containing the default CR for the module"`
	Namespace           string                     `yaml:"namespace" comment:"optional, default=kcp-system, the namespace where the ModuleTemplate will be deployed"`
	Security            string                     `yaml:"security" comment:"optional, name of the security scanners config file"`
	Internal            bool                       `yaml:"internal" comment:"optional, default=false, determines whether the ModuleTemplate should have the internal flag or not"`
	Beta                bool                       `yaml:"beta" comment:"optional, default=false, determines whether the ModuleTemplate should have the beta flag or not"`
	Labels              map[string]string          `yaml:"labels" comment:"optional, additional labels for the ModuleTemplate"`
	Annotations         map[string]string          `yaml:"annotations" comment:"optional, additional annotations for the ModuleTemplate"`
	AssociatedResources []*metav1.GroupVersionKind `yaml:"associatedResources" comment:"optional, GVK of the resources which are associated with the module and have to be deleted with module deletion"`
	Manager             *Manager                   `yaml:"manager" comment:"optional, the module resource that can be used to indicate the installation readiness of the module. This is typically the manager deployment of the module"`
	Resources           Resources                  `yaml:"resources,omitempty" comment:"optional, additional resources of the ModuleTemplate that may be fetched"`
}

type Manager struct {
	Name                    string `yaml:"name" comment:"required, the name of the manager"`
	Namespace               string `yaml:"namespace" comment:"optional, the path to the manager"`
	metav1.GroupVersionKind `yaml:",inline" comment:"required, the GVK of the manager"`
}

// Icons represents a map of icon names to links.
type Icons map[string]string

// UnmarshalYAML unmarshals Icons from YAML format.
func (i *Icons) UnmarshalYAML(unmarshal func(interface{}) error) error {
	dataMap, err := unmarshalToMap(unmarshal)
	if err != nil {
		return fmt.Errorf("failed to unmarshal Icons: %w", err)
	}
	*i = dataMap
	return nil
}

// MarshalYAML marshals Icons to YAML format.
func (i *Icons) MarshalYAML() (interface{}, error) {
	return marshalFromMap(*i)
}

// Resources represents a map of resource names to links.
type Resources map[string]string

// UnmarshalYAML unmarshals Resources from YAML format.
func (rm *Resources) UnmarshalYAML(unmarshal func(interface{}) error) error {
	dataMap, err := unmarshalToMap(unmarshal)
	if err != nil {
		return fmt.Errorf("failed to unmarshal Resources: %w", err)
	}
	*rm = dataMap
	return nil
}

// MarshalYAML marshals Resources to YAML format.
func (rm *Resources) MarshalYAML() (interface{}, error) {
	return marshalFromMap(*rm)
}

func unmarshalToMap(unmarshal func(interface{}) error) (map[string]string, error) {
	var items []nameLinkItem
	if err := unmarshal(&items); err == nil {
		resultMap := make(map[string]string)
		for _, item := range items {
			if _, exists := resultMap[item.Name]; exists {
				return nil, ErrDuplicateMapEntries
			}
			resultMap[item.Name] = item.Link
		}
		return resultMap, nil
	}

	resultMap := make(map[string]string)
	if err := unmarshal(&resultMap); err != nil {
		return nil, err
	}

	return resultMap, nil
}

func marshalFromMap(dataMap map[string]string) (interface{}, error) {
	items := make([]nameLinkItem, 0, len(dataMap))
	for name, link := range dataMap {
		items = append(items, nameLinkItem{Name: name, Link: link})
	}
	return items, nil
}

type nameLinkItem struct {
	Name string `yaml:"name"`
	Link string `yaml:"link"`
}
