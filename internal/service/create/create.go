package create

import (
	"fmt"

	"ocm.software/ocm/api/ocm/compdesc"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/extensions/repositories/comparch"

	commonerrors "github.com/kyma-project/modulectl/internal/common/errors"
	"github.com/kyma-project/modulectl/internal/service/componentarchive"
	"github.com/kyma-project/modulectl/internal/service/componentdescriptor"
	"github.com/kyma-project/modulectl/internal/service/contentprovider"
)

type ModuleConfigService interface {
	ParseAndValidateModuleConfig(moduleConfigFile string) (*contentprovider.ModuleConfig, error)
	GetDefaultCRData(defaultCRPath string) ([]byte, error)
	CleanupTempFiles() []error
}

type SecurityConfigService interface {
	ParseSecurityConfigData(gitRepoURL, securityConfigFile string) (*contentprovider.SecurityScanConfig, error)
	AppendSecurityScanConfig(descriptor *compdesc.ComponentDescriptor,
		securityConfig contentprovider.SecurityScanConfig) error
}

type GitSourcesService interface {
	AddGitSources(componentDescriptor *compdesc.ComponentDescriptor, gitRepoURL, moduleVersion string) error
}

type ComponentArchiveService interface {
	CreateComponentArchive(componentDescriptor *compdesc.ComponentDescriptor) (*comparch.ComponentArchive,
		error)
	AddModuleResourcesToArchive(componentArchive componentarchive.ComponentArchive,
		moduleResources []componentdescriptor.Resource) error
}

type RegistryService interface {
	PushComponentVersion(archive *comparch.ComponentArchive, insecure bool, credentials, registryURL string) error
	GetComponentVersion(archive *comparch.ComponentArchive, insecure bool,
		userPasswordCreds, registryURL string) (cpi.ComponentVersionAccess, error)
}

type ModuleTemplateService interface {
	GenerateModuleTemplate(moduleConfig *contentprovider.ModuleConfig,
		descriptor *compdesc.ComponentDescriptor,
		data []byte,
		isCrdClusterScoped bool,
		templateOutput string) error
}

type CRDParserService interface {
	IsCRDClusterScoped(crPath, manifestPath string) (bool, error)
}

type Service struct {
	moduleConfigService     ModuleConfigService
	gitSourcesService       GitSourcesService
	securityConfigService   SecurityConfigService
	componentArchiveService ComponentArchiveService
	registryService         RegistryService
	moduleTemplateService   ModuleTemplateService
	crdParserService        CRDParserService
}

func NewService(moduleConfigService ModuleConfigService,
	gitSourcesService GitSourcesService,
	securityConfigService SecurityConfigService,
	componentArchiveService ComponentArchiveService,
	registryService RegistryService,
	moduleTemplateService ModuleTemplateService,
	crdParserService CRDParserService,
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

	if registryService == nil {
		return nil, fmt.Errorf("%w: registryService must not be nil", commonerrors.ErrInvalidArg)
	}

	if moduleTemplateService == nil {
		return nil, fmt.Errorf("%w: moduleTemplateService must not be nil", commonerrors.ErrInvalidArg)
	}

	if crdParserService == nil {
		return nil, fmt.Errorf("%w: crdParserService must not be nil", commonerrors.ErrInvalidArg)
	}

	return &Service{
		moduleConfigService:     moduleConfigService,
		gitSourcesService:       gitSourcesService,
		securityConfigService:   securityConfigService,
		componentArchiveService: componentArchiveService,
		registryService:         registryService,
		moduleTemplateService:   moduleTemplateService,
		crdParserService:        crdParserService,
	}, nil
}

func (s *Service) CreateModule(opts Options) error {
	if err := opts.Validate(); err != nil {
		return err
	}
	defer s.moduleConfigService.CleanupTempFiles()
	moduleConfig, err := s.moduleConfigService.ParseAndValidateModuleConfig(opts.ModuleConfigFile)
	if err != nil {
		return fmt.Errorf("%w: failed to parse module config", err)
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

	if opts.GitRemote != "" {
		if err := s.gitSourcesService.AddGitSources(componentDescriptor, opts.GitRemote,
			moduleConfig.Version); err != nil {
			return fmt.Errorf("%w: failed to add git sources", err)
		}
	}
	opts.Out.Write("- Configuring security scanners config\n")
	if moduleConfig.Security != "" && opts.GitRemote != "" {
		securityConfig, err := s.securityConfigService.ParseSecurityConfigData(opts.GitRemote, moduleConfig.Security)
		if err != nil {
			return fmt.Errorf("%w: failed to parse security config data", err)
		}

		if err := s.securityConfigService.AppendSecurityScanConfig(componentDescriptor, *securityConfig); err != nil {
			return fmt.Errorf("%w: failed to append security scan config", err)
		}
	}
	isCRDClusterScoped, err := s.crdParserService.IsCRDClusterScoped(moduleConfig.DefaultCRPath,
		moduleConfig.ManifestPath)
	if err != nil {
		return fmt.Errorf("%w: failed to determine if CRD is cluster scoped", err)
	}

	opts.Out.Write("- Creating component archive\n")
	componentArchive, err := s.componentArchiveService.CreateComponentArchive(componentDescriptor)
	if err != nil {
		return fmt.Errorf("%w: failed to create component archive", err)
	}
	if err = s.componentArchiveService.AddModuleResourcesToArchive(componentArchive,
		moduleResources); err != nil {
		return fmt.Errorf("%w: failed to add module resources to component archive", err)
	}

	if opts.RegistryURL == "" {
		return nil
	}
	opts.Out.Write("- Pushing component version\n")
	if err = s.registryService.PushComponentVersion(componentArchive, opts.Insecure, opts.Credentials,
		opts.RegistryURL); err != nil {
		return fmt.Errorf("%w: failed to push component archive", err)
	}

	componentVersionAccess, err := s.registryService.GetComponentVersion(componentArchive, opts.Insecure,
		opts.Credentials, opts.RegistryURL)
	if err != nil {
		return fmt.Errorf("%w: failed to get component version", err)
	}

	crData, err := s.moduleConfigService.GetDefaultCRData(moduleConfig.DefaultCRPath)
	if err != nil {
		return fmt.Errorf("%w: failed to get default CR data", err)
	}

	opts.Out.Write("- Generating ModuleTemplate\n")
	descriptor := componentVersionAccess.GetDescriptor()
	if err = s.moduleTemplateService.GenerateModuleTemplate(moduleConfig, descriptor,
		crData, isCRDClusterScoped, opts.TemplateOutput); err != nil {
		return fmt.Errorf("%w: failed to generate module template", err)
	}

	return nil
}
