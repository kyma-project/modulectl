package create

import (
	"fmt"

	commonerrors "github.com/kyma-project/modulectl/internal/common/errors"
	"github.com/kyma-project/modulectl/internal/service/contentprovider"
)

type ModuleConfigService interface {
	ParseModuleConfig(configFilePath string) (*contentprovider.ModuleConfig, error)
	ValidateModuleConfig(moduleConfig *contentprovider.ModuleConfig) error
	GetDefaultCRPath(defaultCRPath string) (string, error)
	GetManifestPath(manifestPath string) (string, error)
	CleanupTempFiles() []error
}

type Service struct {
	moduleConfigService ModuleConfigService
}

func NewService(moduleConfigService ModuleConfigService) (*Service, error) {
	if moduleConfigService == nil {
		return nil, fmt.Errorf("%w: moduleConfigService must not be nil", commonerrors.ErrInvalidArg)
	}

	return &Service{
		moduleConfigService: moduleConfigService,
	}, nil
}

func (s *Service) CreateModule(opts Options) error {
	if err := opts.Validate(); err != nil {
		return err
	}

	defer s.moduleConfigService.CleanupTempFiles()

	moduleConfig, err := s.moduleConfigService.ParseModuleConfig(opts.ModuleConfigFile)
	if err != nil {
		return fmt.Errorf("%w: failed to parse module config file", err)
	}

	if err := s.moduleConfigService.ValidateModuleConfig(moduleConfig); err != nil {
		return fmt.Errorf("%w: failed to value module config", err)
	}

	moduleConfig.DefaultCRPath, err = s.moduleConfigService.GetDefaultCRPath(moduleConfig.DefaultCRPath)
	if err != nil {
		return fmt.Errorf("%w: failed to get default CR path", err)
	}

	moduleConfig.ManifestPath, err = s.moduleConfigService.GetManifestPath(moduleConfig.ManifestPath)
	if err != nil {
		return fmt.Errorf("%w: failed to get manifest path", err)
	}

	return nil
}
