package contentprovider

import (
	"fmt"

	commonerrors "github.com/kyma-project/modulectl/internal/common/errors"
	"github.com/kyma-project/modulectl/internal/common/types"
)

type ModuleConfig struct {
	yamlConverter ObjectToYAMLConverter
}

func NewModuleConfig(yamlConverter ObjectToYAMLConverter) (*ModuleConfig, error) {
	if yamlConverter == nil {
		return nil, fmt.Errorf("%w: yamlConverter must not be nil", commonerrors.ErrInvalidArg)
	}

	return &ModuleConfig{
		yamlConverter: yamlConverter,
	}, nil
}

func (s *ModuleConfig) GetDefaultContent(args types.KeyValueArgs) (string, error) {
	if err := s.validateArgs(args); err != nil {
		return "", err
	}

	moduleConfig := s.getModuleConfig(args)

	return s.yamlConverter.ConvertToYaml(moduleConfig), nil
}

func (s *ModuleConfig) getModuleConfig(args types.KeyValueArgs) moduleConfig {
	return moduleConfig{
		Name:          args[ArgModuleName],
		Version:       args[ArgModuleVersion],
		Channel:       args[ArgModuleChannel],
		ManifestPath:  args[ArgManifestFile],
		Security:      args[ArgSecurityConfigFile],
		DefaultCRPath: args[ArgDefaultCRFile],
	}
}

func (s *ModuleConfig) validateArgs(args types.KeyValueArgs) error {
	if args == nil {
		return fmt.Errorf("%w: args must not be nil", ErrInvalidArg)
	}

	if value, ok := args[ArgModuleName]; !ok {
		return fmt.Errorf("%w: %s", ErrMissingArg, ArgModuleName)
	} else if value == "" {
		return fmt.Errorf("%w: %s must not be empty", ErrInvalidArg, ArgModuleName)
	}

	if value, ok := args[ArgModuleVersion]; !ok {
		return fmt.Errorf("%w: %s", ErrMissingArg, ArgModuleVersion)
	} else if value == "" {
		return fmt.Errorf("%w: %s must not be empty", ErrInvalidArg, ArgModuleVersion)
	}

	if value, ok := args[ArgModuleChannel]; !ok {
		return fmt.Errorf("%w: %s", ErrMissingArg, ArgModuleChannel)
	} else if value == "" {
		return fmt.Errorf("%w: %s must not be empty", ErrInvalidArg, ArgModuleChannel)
	}

	return nil
}

type moduleConfig struct {
	Name          string            `yaml:"name" comment:"required, the name of the Module"`
	Version       string            `yaml:"version" comment:"required, the version of the Module"`
	Channel       string            `yaml:"channel" comment:"required, channel that should be used in the ModuleTemplate"`
	ManifestPath  string            `yaml:"manifest" comment:"required, relative path or remote URL to the manifests"`
	Mandatory     bool              `yaml:"mandatory" comment:"optional, default=false, indicates whether the module is mandatory to be installed on all clusters"`
	DefaultCRPath string            `yaml:"defaultCR" comment:"optional, relative path or remote URL to a YAML file containing the default CR for the module"`
	ResourceName  string            `yaml:"resourceName" comment:"optional, default={name}-{channel}, when channel is 'none', the default is {name}-{version}, the name for the ModuleTemplate that will be created"`
	Namespace     string            `yaml:"namespace" comment:"optional, default=kcp-system, the namespace where the ModuleTemplate will be deployed"`
	Security      string            `yaml:"security" comment:"optional, name of the security scanners config file"`
	Internal      bool              `yaml:"internal" comment:"optional, default=false, determines whether the ModuleTemplate should have the internal flag or not"`
	Beta          bool              `yaml:"beta" comment:"optional, default=false, determines whether the ModuleTemplate should have the beta flag or not"`
	Labels        map[string]string `yaml:"labels" comment:"optional, additional labels for the ModuleTemplate"`
	Annotations   map[string]string `yaml:"annotations" comment:"optional, additional annotations for the ModuleTemplate"`
}
