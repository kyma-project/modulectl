package create

import (
	"fmt"

	commonerrors "github.com/kyma-project/modulectl/internal/common/errors"
	"github.com/kyma-project/modulectl/internal/service/componentdescriptor"
	"github.com/kyma-project/modulectl/internal/service/contentprovider"
	ocm "ocm.software/ocm/api/ocm/compdesc/versions/v2"
)

type ModuleConfigService interface {
	ParseModuleConfig(configFilePath string) (*contentprovider.ModuleConfig, error)
	ValidateModuleConfig(moduleConfig *contentprovider.ModuleConfig) error
	GetDefaultCRPath(defaultCRPath string) (string, error)
	GetManifestPath(manifestPath string) (string, error)
	CleanupTempFiles() []error
}

type SecurityConfigService interface {
	ParseSecurityConfigData(gitRepoURL, securityConfigFile string) (*contentprovider.SecurityScanConfig, error)
	AppendSecurityScanConfig(descriptor *ocm.ComponentDescriptor,
		securityConfig contentprovider.SecurityScanConfig) error
}

type GitSourcesService interface {
	AddGitSources(componentDescriptor *ocm.ComponentDescriptor, gitRepoURL, moduleVersion string) error
}

type Service struct {
	moduleConfigService   ModuleConfigService
	gitSourcesService     GitSourcesService
	securityConfigService SecurityConfigService
}

func NewService(moduleConfigService ModuleConfigService,
	gitSourcesService GitSourcesService,
	securityConfigService SecurityConfigService) (*Service, error) {
	if moduleConfigService == nil {
		return nil, fmt.Errorf("%w: moduleConfigService must not be nil", commonerrors.ErrInvalidArg)
	}

	if gitSourcesService == nil {
		return nil, fmt.Errorf("%w: gitSourcesService must not be nil", commonerrors.ErrInvalidArg)
	}

	if securityConfigService == nil {
		return nil, fmt.Errorf("%w: securityConfigService must not be nil", commonerrors.ErrInvalidArg)
	}

	return &Service{
		moduleConfigService:   moduleConfigService,
		gitSourcesService:     gitSourcesService,
		securityConfigService: securityConfigService,
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

	componentDescriptor, err := componentdescriptor.InitializeComponentDescriptor(moduleConfig.Name,
		moduleConfig.Version)
	if err != nil {
		return fmt.Errorf("%w: failed to populate component descriptor metadata", err)
	}

	// TODO: Populate component descriptor resources (module layers)

	if err := s.gitSourcesService.AddGitSources(componentDescriptor, opts.GitRemote, moduleConfig.Version); err != nil {
		return fmt.Errorf("%w: failed to add git sources", err)
	}

	if moduleConfig.Security != "" {
		securityConfig, err := s.securityConfigService.ParseSecurityConfigData(opts.GitRemote, moduleConfig.Security)
		if err != nil {
			return fmt.Errorf("%w: failed to parse security config data", err)
		}

		if err := s.securityConfigService.AppendSecurityScanConfig(componentDescriptor, *securityConfig); err != nil {
			return fmt.Errorf("%w: failed to append security scan config", err)
		}

	}

	return nil
}
