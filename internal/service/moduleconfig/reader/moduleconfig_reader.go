package moduleconfigreader

import (
	"fmt"

	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	commonerrors "github.com/kyma-project/modulectl/internal/common/errors"
	"github.com/kyma-project/modulectl/internal/common/validation"
	"github.com/kyma-project/modulectl/internal/service/contentprovider"
)

type FileSystem interface {
	ReadFile(path string) ([]byte, error)
}

type Service struct {
	fileSystem FileSystem
}

func NewService(fileSystem FileSystem) (*Service, error) {
	if fileSystem == nil {
		return nil, fmt.Errorf("fileSystem must not be nil: %w", commonerrors.ErrInvalidArg)
	}

	return &Service{
		fileSystem: fileSystem,
	}, nil
}

func (s *Service) ParseAndValidateModuleConfig(moduleConfigFile string,
) (*contentprovider.ModuleConfig, error) {
	moduleConfig, err := ParseModuleConfig(moduleConfigFile, s.fileSystem)
	if err != nil {
		return nil, fmt.Errorf("failed to parse module config file: %w", err)
	}

	if err = ValidateModuleConfig(moduleConfig); err != nil {
		return nil, fmt.Errorf("failed to validate module config: %w", err)
	}

	return moduleConfig, nil
}

func ValidateModuleConfig(moduleConfig *contentprovider.ModuleConfig) error {
	if err := validation.ValidateModuleName(moduleConfig.Name); err != nil {
		return fmt.Errorf("failed to validate module name: %w", err)
	}

	if err := validation.ValidateModuleVersion(moduleConfig.Version); err != nil {
		return fmt.Errorf("failed to validate module version: %w", err)
	}

	if err := validation.ValidateModuleNamespace(moduleConfig.Namespace); err != nil {
		return fmt.Errorf("failed to validate module namespace: %w", err)
	}

	if err := validation.ValidateIsValidHTTPSURL(moduleConfig.Manifest); err != nil {
		return fmt.Errorf("failed to validate manifest: %w", err)
	}

	if err := validation.ValidateIsValidHTTPSURL(moduleConfig.Repository); err != nil {
		return fmt.Errorf("failed to validate repository: %w", err)
	}

	if err := validation.ValidateIsValidHTTPSURL(moduleConfig.Documentation); err != nil {
		return fmt.Errorf("failed to validate documentation: %w", err)
	}

	if len(moduleConfig.Icons) == 0 {
		return fmt.Errorf("failed to validate module icons: must contain at least one icon: %w",
			commonerrors.ErrInvalidOption)
	}

	if err := validation.ValidateMapEntries(moduleConfig.Icons); err != nil {
		return fmt.Errorf("failed to validate module icons: %w", err)
	}

	if err := validation.ValidateMapEntries(moduleConfig.Resources); err != nil {
		return fmt.Errorf("failed to validate resources: %w", err)
	}

	if moduleConfig.DefaultCR != "" {
		if err := validation.ValidateIsValidHTTPSURL(moduleConfig.DefaultCR); err != nil {
			return fmt.Errorf("failed to validate default CR: %w", err)
		}
	}

	if err := ValidateAssociatedResources(moduleConfig.AssociatedResources); err != nil {
		return fmt.Errorf("failed to validate associated resources: %w", err)
	}

	if err := ValidateManager(moduleConfig.Manager); err != nil {
		return fmt.Errorf("failed to validate manager: %w", err)
	}

	return nil
}

func ValidateAssociatedResources(resources []*metav1.GroupVersionKind) error {
	for _, resource := range resources {
		if err := validation.ValidateGvk(resource.Group, resource.Version, resource.Kind); err != nil {
			return fmt.Errorf("GVK is invalid: %w", err)
		}
	}
	return nil
}

func ValidateManager(manager *contentprovider.Manager) error {
	if manager == nil {
		return nil
	}

	if manager.Name == "" {
		return fmt.Errorf("name must not be empty: %w", commonerrors.ErrInvalidOption)
	}

	if err := validation.ValidateGvk(manager.Group, manager.Version, manager.Kind); err != nil {
		return fmt.Errorf("GVK is invalid: %w", err)
	}

	if manager.Namespace != "" {
		if err := validation.ValidateNamespace(manager.Namespace); err != nil {
			return fmt.Errorf("namespace is invalid: %w", err)
		}
	}

	return nil
}

func ParseModuleConfig(configFilePath string, fileSystem FileSystem) (*contentprovider.ModuleConfig, error) {
	moduleConfigData, err := fileSystem.ReadFile(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read module config file: %w", err)
	}

	moduleConfig := &contentprovider.ModuleConfig{}
	if err := yaml.Unmarshal(moduleConfigData, moduleConfig); err != nil {
		return nil, fmt.Errorf("failed to parse module config file: %w", err)
	}

	if moduleConfig.Namespace == "" {
		moduleConfig.Namespace = "kcp-system"
	}

	return moduleConfig, nil
}
