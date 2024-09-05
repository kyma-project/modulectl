package moduleconfigreader

import (
	"fmt"

	"gopkg.in/yaml.v3"

	commonerrors "github.com/kyma-project/modulectl/internal/common/errors"
	"github.com/kyma-project/modulectl/internal/common/validations"
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
		return nil, fmt.Errorf("%w: fileSystem must not be nil", commonerrors.ErrInvalidArg)
	}

	return &Service{
		fileSystem: fileSystem,
	}, nil
}

func (s *Service) ParseModuleConfig(configFilePath string) (*contentprovider.ModuleConfig, error) {
	moduleConfigData, err := s.fileSystem.ReadFile(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read module config file: %w", err)
	}

	moduleConfig := &contentprovider.ModuleConfig{}
	if err := yaml.Unmarshal(moduleConfigData, moduleConfig); err != nil {
		return nil, fmt.Errorf("failed to parse module config file: %w", err)
	}

	return moduleConfig, nil
}

func (*Service) ValidateModuleConfig(moduleConfig *contentprovider.ModuleConfig) error {
	if err := validations.ValidateModuleName(moduleConfig.Name); err != nil {
		return fmt.Errorf("failed to validate module name: %w", err)
	}

	if err := validations.ValidateModuleVersion(moduleConfig.Version); err != nil {
		return fmt.Errorf("failed to validate module version: %w", err)
	}

	if err := validations.ValidateModuleChannel(moduleConfig.Channel); err != nil {
		return fmt.Errorf("failed to validate module channel: %w", err)
	}

	if err := validations.ValidateModuleNamespace(moduleConfig.Namespace); err != nil {
		return fmt.Errorf("failed to validate module namespace: %w", err)
	}

	return nil
}
