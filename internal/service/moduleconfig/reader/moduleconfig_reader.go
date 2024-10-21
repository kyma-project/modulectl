package moduleconfigreader

import (
	"errors"
	"fmt"

	"gopkg.in/yaml.v3"

	commonerrors "github.com/kyma-project/modulectl/internal/common/errors"
	"github.com/kyma-project/modulectl/internal/common/validation"
	"github.com/kyma-project/modulectl/internal/service/contentprovider"
)

var ErrNoPathForDefaultCR = errors.New("no path for default CR given")

type FileSystem interface {
	ReadFile(path string) ([]byte, error)
}

type Service struct {
	fileSystem FileSystem
}

func NewService(fileSystem FileSystem) (*Service, error) {
	if fileSystem == nil {
		return nil, fmt.Errorf("%w: fileSystem must not be nil", commonerrors.ErrInvalidArg)
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

	if err := validation.ValidateModuleChannel(moduleConfig.Channel); err != nil {
		return fmt.Errorf("failed to validate module channel: %w", err)
	}

	if err := validation.ValidateModuleNamespace(moduleConfig.Namespace); err != nil {
		return fmt.Errorf("failed to validate module namespace: %w", err)
	}

	if err := validation.ValidateResources(moduleConfig.Resources); err != nil {
		return fmt.Errorf("failed to validate resources: %w", err)
	}

	if moduleConfig.Manifest == "" {
		return fmt.Errorf("manifest must not be empty: %w", commonerrors.ErrInvalidOption)
	}

	if err := ValidateManager(moduleConfig.Manager); err != nil {
		return fmt.Errorf("failed to validate manager: %w", err)
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

	if manager.Kind == "" {
		return fmt.Errorf("kind must not be empty: %w", commonerrors.ErrInvalidOption)
	}

	if manager.Group == "" {
		return fmt.Errorf("group must not be empty: %w", commonerrors.ErrInvalidOption)
	}

	if manager.Version == "" {
		return fmt.Errorf("version must not be empty: %w", commonerrors.ErrInvalidOption)
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
