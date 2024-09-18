package create

import (
	"fmt"

	ocm "ocm.software/ocm/api/ocm/compdesc"
	"ocm.software/ocm/api/ocm/extensions/repositories/comparch"

	commonerrors "github.com/kyma-project/modulectl/internal/common/errors"
	"github.com/kyma-project/modulectl/internal/service/componentdescriptor"
	"github.com/kyma-project/modulectl/internal/service/contentprovider"
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

type ComponentArchiveService interface {
	CreateComponentArchive(componentDescriptor *ocm.ComponentDescriptor) (*comparch.ComponentArchive,
		error)
	AddModuleResourcesToArchive(componentArchive *comparch.ComponentArchive,
		moduleResources []componentdescriptor.Resource) error
}

type Service struct {
	moduleConfigService     ModuleConfigService
	gitSourcesService       GitSourcesService
	securityConfigService   SecurityConfigService
	componentArchiveService ComponentArchiveService
}

func NewService(moduleConfigService ModuleConfigService,
	gitSourcesService GitSourcesService,
	securityConfigService SecurityConfigService,
	componentArchiveService ComponentArchiveService,
) (*Service, error) {
	if moduleConfigService == nil {
		return nil, fmt.Errorf("%w: moduleConfigService must not be nil", commonerrors.ErrInvalidArg)
	}

	if gitSourcesService == nil {
		return nil, fmt.Errorf("%w: gitSourcesService must not be nil", commonerrors.ErrInvalidArg)
	}

	if securityConfigService == nil {
		return nil, fmt.Errorf("%w: securityConfigService must not be nil", commonerrors.ErrInvalidArg)
	}

	if componentArchiveService == nil {
		return nil, fmt.Errorf("%w: componentArchiveService must not be nil", commonerrors.ErrInvalidArg)
	}

	return &Service{
		moduleConfigService:     moduleConfigService,
		gitSourcesService:       gitSourcesService,
		securityConfigService:   securityConfigService,
		componentArchiveService: componentArchiveService,
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

	moduleResources, err := componentdescriptor.GenerateModuleResources(moduleConfig.Version, moduleConfig.ManifestPath,
		moduleConfig.DefaultCRPath, opts.RegistryCredSelector)
	if err != nil {
		return fmt.Errorf("%w: failed to generate module resources", err)
	}

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

	componentArchive, err := s.componentArchiveService.CreateComponentArchive(componentDescriptor)
	if err != nil {
		return fmt.Errorf("%w: failed to create component archive", err)
	}

	if err := s.componentArchiveService.AddModuleResourcesToArchive(componentArchive,
		moduleResources); err != nil {
		return fmt.Errorf("%w: failed to add module resources to component archive", err)
	}

	return nil
}
