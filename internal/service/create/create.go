package create

import (
	"fmt"

	commonerrors "github.com/kyma-project/modulectl/internal/common/errors"
	"github.com/kyma-project/modulectl/internal/service/contentprovider"
)

type ModuleConfigService interface {
	ParseModuleConfig(configFilePath string) (*contentprovider.ModuleConfig, error)
	ValidateModuleConfig(moduleConfig *contentprovider.ModuleConfig) error
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

	moduleConfig, err := s.moduleConfigService.ParseModuleConfig(opts.ModuleConfigFile)
	if err != nil {
		return fmt.Errorf("%w: failed to parse module config file", err)
	}

	if err := s.moduleConfigService.ValidateModuleConfig(moduleConfig); err != nil {
		return fmt.Errorf("%w: failed to value module config", err)
	}

	return nil
}
