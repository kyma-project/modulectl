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

	defer func() {
		if err := s.moduleConfigService.CleanupTempFiles(); err != nil {
			opts.Out.Write(fmt.Sprintf("failed to cleanup temporary files: %v\n", err))
		}
	}()

	moduleConfig, err := s.moduleConfigService.ParseAndValidateModuleConfig(opts.ModuleConfigFile)
	if err != nil {
		return fmt.Errorf("failed to parse module config: %w", err)
	}

	descriptor, err := componentdescriptor.InitializeComponentDescriptor(moduleConfig.Name, moduleConfig.Version)
	if err != nil {
		return fmt.Errorf("failed to populate component descriptor metadata: %w", err)
	}

	moduleResources, err := componentdescriptor.GenerateModuleResources(moduleConfig.Version, moduleConfig.ManifestPath,
		moduleConfig.DefaultCRPath, opts.RegistryCredSelector)
	if err != nil {
		return fmt.Errorf("failed to generate module resources: %w", err)
	}

	if opts.GitRemote != "" {
		if err = s.gitSourcesService.AddGitSources(descriptor, opts.GitRemote,
			moduleConfig.Version); err != nil {
			return fmt.Errorf("failed to add git sources: %w", err)
		}
		if moduleConfig.Security != "" {
			err = s.configureSecScannerConf(descriptor, moduleConfig, opts)
			if err != nil {
				return fmt.Errorf("failed to configure security scanners: %w", err)
			}
		}
	}

	opts.Out.Write("- Creating component archive\n")
	archive, err := s.componentArchiveService.CreateComponentArchive(descriptor)
	if err != nil {
		return fmt.Errorf("failed to create component archive: %w", err)
	}
	if err = s.componentArchiveService.AddModuleResourcesToArchive(archive,
		moduleResources); err != nil {
		return fmt.Errorf("failed to add module resources to component archive: %w", err)
	}

	if opts.RegistryURL != "" {
		return s.pushImgAndCreateTemplate(archive, moduleConfig, opts)
	}
	return nil
}

func (s *Service) pushImgAndCreateTemplate(archive *comparch.ComponentArchive, moduleConfig *contentprovider.ModuleConfig, opts Options) error {
	opts.Out.Write("- Pushing component version\n")
	isCRDClusterScoped, err := s.crdParserService.IsCRDClusterScoped(moduleConfig.DefaultCRPath, moduleConfig.ManifestPath)
	if err != nil {
		return fmt.Errorf("failed to determine if CRD is cluster scoped: %w", err)
	}

	if err := s.registryService.PushComponentVersion(archive, opts.Insecure, opts.Credentials,
		opts.RegistryURL); err != nil {
		return fmt.Errorf("%w: failed to push component archive", err)
	}

	componentVersionAccess, err := s.registryService.GetComponentVersion(archive, opts.Insecure, opts.Credentials, opts.RegistryURL)
	if err != nil {
		return fmt.Errorf("%w: failed to get component version", err)
	}

	var crData []byte
	if moduleConfig.DefaultCRPath != "" {
		crData, err = s.moduleConfigService.GetDefaultCRData(moduleConfig.DefaultCRPath)
		if err != nil {
			return fmt.Errorf("%w: failed to get default CR data", err)
		}
	}

	opts.Out.Write("- Generating ModuleTemplate\n")
	descriptor := componentVersionAccess.GetDescriptor()
	if err = s.moduleTemplateService.GenerateModuleTemplate(moduleConfig, descriptor,
		crData, isCRDClusterScoped, opts.TemplateOutput); err != nil {
		return fmt.Errorf("%w: failed to generate module template", err)
	}
	return nil
}

func (s *Service) configureSecScannerConf(descriptor *compdesc.ComponentDescriptor, moduleConfig *contentprovider.ModuleConfig, opts Options) error {
	opts.Out.Write("- Configuring security scanners config\n")
	securityConfig, err := s.securityConfigService.ParseSecurityConfigData(opts.GitRemote, moduleConfig.Security)
	if err != nil {
		return fmt.Errorf("%w: failed to parse security config data", err)
	}

	if err = s.securityConfigService.AppendSecurityScanConfig(descriptor, *securityConfig); err != nil {
		return fmt.Errorf("%w: failed to append security scan config", err)
	}
	return nil
}
