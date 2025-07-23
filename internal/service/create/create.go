package create

import (
	"errors"
	"fmt"
	"path"

	"ocm.software/ocm/api/ocm/compdesc"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/extensions/repositories/comparch"

	commonerrors "github.com/kyma-project/modulectl/internal/common/errors"
	"github.com/kyma-project/modulectl/internal/service/componentarchive"
	"github.com/kyma-project/modulectl/internal/service/componentdescriptor"
	"github.com/kyma-project/modulectl/internal/service/componentdescriptor/resources"
	"github.com/kyma-project/modulectl/internal/service/contentprovider"
	"github.com/kyma-project/modulectl/internal/utils"
)

var ErrComponentVersionExists = errors.New("component version already exists")

type ModuleConfigService interface {
	ParseAndValidateModuleConfig(moduleConfigFile string) (*contentprovider.ModuleConfig, error)
}

type FileSystem interface {
	ReadFile(path string) ([]byte, error)
}

type FileResolver interface {
	// Resolve resolves a file reference, which can be either a URL or a local file path (may be just a file name).
	// For local file paths, it will resolve the path relative to the provided basePath (absolute or relative).
	Resolve(fileRef contentprovider.UrlOrLocalFile, basePath string) (string, error)
	CleanupTempFiles() []error
}

type SecurityConfigService interface {
	ParseSecurityConfigData(securityConfigFile string) (*contentprovider.SecurityScanConfig, error)
	AppendSecurityScanConfig(descriptor *compdesc.ComponentDescriptor,
		securityConfig contentprovider.SecurityScanConfig) error
}

type GitSourcesService interface {
	AddGitSources(componentDescriptor *compdesc.ComponentDescriptor,
		gitRepoPath, gitRepoURL, moduleVersion string) error
}

type ComponentArchiveService interface {
	CreateComponentArchive(componentDescriptor *compdesc.ComponentDescriptor) (*comparch.ComponentArchive,
		error)
	AddModuleResourcesToArchive(componentArchive componentarchive.ComponentArchive,
		moduleResources []resources.Resource) error
}

type RegistryService interface {
	PushComponentVersion(archive *comparch.ComponentArchive,
		insecure bool,
		overwrite bool,
		credentials string,
		registryURL string,
	) error
	GetComponentVersion(archive *comparch.ComponentArchive,
		insecure bool,
		userPasswordCreds string,
		registryURL string,
	) (cpi.ComponentVersionAccess, error)
	ExistsComponentVersion(archive *comparch.ComponentArchive,
		insecure bool,
		credentials string,
		registryURL string,
	) (bool, error)
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

type ModuleResourceService interface {
	GenerateModuleResources(moduleConfig *contentprovider.ModuleConfig,
		manifestPath, defaultCRPath string) ([]resources.Resource, error)
}

type ImageVersionVerifierService interface {
	VerifyModuleResources(moduleConfig *contentprovider.ModuleConfig, filePath string) error
}

type ManifestService interface {
	ExtractImagesFromManifest(manifestPath string) ([]string, error)
}
type Service struct {
	moduleConfigService         ModuleConfigService
	gitSourcesService           GitSourcesService
	securityConfigService       SecurityConfigService
	componentArchiveService     ComponentArchiveService
	registryService             RegistryService
	moduleTemplateService       ModuleTemplateService
	crdParserService            CRDParserService
	moduleResourceService       ModuleResourceService
	imageVersionVerifierService ImageVersionVerifierService
	manifestService             ManifestService
	manifestFileResolver        FileResolver
	defaultCRFileResolver       FileResolver
	fileSystem                  FileSystem
}

func NewService(moduleConfigService ModuleConfigService,
	gitSourcesService GitSourcesService,
	securityConfigService SecurityConfigService,
	componentArchiveService ComponentArchiveService,
	registryService RegistryService,
	moduleTemplateService ModuleTemplateService,
	crdParserService CRDParserService,
	moduleResourceService ModuleResourceService,
	imageVersionVerifierService ImageVersionVerifierService,
	manifestService ManifestService,
	manifestFileResolver FileResolver,
	defaultCRFileResolver FileResolver,
	fileSystem FileSystem,
) (*Service, error) {
	if moduleConfigService == nil {
		return nil, fmt.Errorf("moduleConfigService must not be nil: %w", commonerrors.ErrInvalidArg)
	}

	if gitSourcesService == nil {
		return nil, fmt.Errorf("gitSourcesService must not be nil: %w", commonerrors.ErrInvalidArg)
	}

	if securityConfigService == nil {
		return nil, fmt.Errorf("securityConfigService must not be nil: %w", commonerrors.ErrInvalidArg)
	}

	if componentArchiveService == nil {
		return nil, fmt.Errorf("componentArchiveService must not be nil: %w", commonerrors.ErrInvalidArg)
	}

	if registryService == nil {
		return nil, fmt.Errorf("registryService must not be nil: %w", commonerrors.ErrInvalidArg)
	}

	if moduleTemplateService == nil {
		return nil, fmt.Errorf("moduleTemplateService must not be nil: %w", commonerrors.ErrInvalidArg)
	}

	if crdParserService == nil {
		return nil, fmt.Errorf("crdParserService must not be nil: %w", commonerrors.ErrInvalidArg)
	}

	if moduleResourceService == nil {
		return nil, fmt.Errorf("moduleResourceService must not be nil: %w", commonerrors.ErrInvalidArg)
	}

	if imageVersionVerifierService == nil {
		return nil, fmt.Errorf("imageVersionVerifierService must not be nil: %w", commonerrors.ErrInvalidArg)
	}
	if manifestService == nil {
		return nil, fmt.Errorf("manifestService must not be nil: %w", commonerrors.ErrInvalidArg)
	}

	if manifestFileResolver == nil {
		return nil, fmt.Errorf("manifestFileResolver must not be nil: %w", commonerrors.ErrInvalidArg)
	}

	if defaultCRFileResolver == nil {
		return nil, fmt.Errorf("defaultCRFileResolver must not be nil: %w", commonerrors.ErrInvalidArg)
	}

	if fileSystem == nil {
		return nil, fmt.Errorf("fileSystem must not be nil: %w", commonerrors.ErrInvalidArg)
	}

	return &Service{
		moduleConfigService:         moduleConfigService,
		gitSourcesService:           gitSourcesService,
		securityConfigService:       securityConfigService,
		componentArchiveService:     componentArchiveService,
		registryService:             registryService,
		moduleTemplateService:       moduleTemplateService,
		crdParserService:            crdParserService,
		moduleResourceService:       moduleResourceService,
		imageVersionVerifierService: imageVersionVerifierService,
		manifestService:             manifestService,
		manifestFileResolver:        manifestFileResolver,
		defaultCRFileResolver:       defaultCRFileResolver,
		fileSystem:                  fileSystem,
	}, nil
}

//nolint:funlen // this is a straight down aggregation of the individual steps
func (s *Service) Run(opts Options) error {
	if err := opts.Validate(); err != nil {
		return err
	}

	defer s.cleanupTempFiles(opts)

	moduleConfig, err := s.moduleConfigService.ParseAndValidateModuleConfig(opts.ConfigFile)
	if err != nil {
		return fmt.Errorf("failed to parse module config: %w", err)
	}

	configFilePath := path.Dir(opts.ConfigFile)
	// If the manifest is a local file reference, it's entry in the module config file will be relative to the module config file location (usually the same directory).
	manifestFilePath, err := s.manifestFileResolver.Resolve(moduleConfig.Manifest, configFilePath)
	if err != nil {
		return fmt.Errorf("failed to resolve manifest file: %w", err)
	}

	var defaultCRFilePath string
	if !moduleConfig.DefaultCR.IsEmpty() {
		// If the defaultCR is a local file reference, it's entry in the module config file will be relative to the module config file location (usually the same directory).
		defaultCRFilePath, err = s.defaultCRFileResolver.Resolve(moduleConfig.DefaultCR, configFilePath)
		if err != nil {
			return fmt.Errorf("failed to resolve default CR file: %w", err)
		}
	}

	descriptor, err := componentdescriptor.InitializeComponentDescriptor(moduleConfig.Name, moduleConfig.Version)
	if err != nil {
		return fmt.Errorf("failed to populate component descriptor metadata: %w", err)
	}

	if err = s.gitSourcesService.AddGitSources(descriptor, opts.ModuleSourcesGitDirectory, moduleConfig.Repository,
		moduleConfig.Version); err != nil {
		return fmt.Errorf("failed to add git sources: %w", err)
	}

	var securityConfigImages []string
	if moduleConfig.Security != "" {
		securityConfigImages, err = s.configureSecScannerConf(descriptor, moduleConfig, opts)
		if err != nil {
			return fmt.Errorf("failed to configure security scanners: %w", err)
		}
	}

	manifestImages, err := s.extractImagesFromManifest(manifestFilePath, opts)
	if err != nil {
		return fmt.Errorf("failed to extract images from manifest: %w", err)
	}

	images := utils.MergeAndDeduplicateSlices(securityConfigImages, manifestImages)
	err = addImagesOciArtifactsToDescriptor(descriptor, images, opts)
	if err != nil {
		return fmt.Errorf("failed to create oci artifact component for raw manifest: %w", err)
	}

	opts.Out.Write("- Creating component archive\n")
	archive, err := s.componentArchiveService.CreateComponentArchive(descriptor)
	if err != nil {
		return fmt.Errorf("failed to create component archive: %w", err)
	}

	if err := s.imageVersionVerifierService.VerifyModuleResources(moduleConfig, manifestFilePath); err != nil {
		return fmt.Errorf("failed to verify module resources: %w", err)
	}

	moduleResources, err := s.moduleResourceService.GenerateModuleResources(moduleConfig, manifestFilePath,
		defaultCRFilePath)
	if err != nil {
		return fmt.Errorf("failed to generate module resources: %w", err)
	}

	if err = s.componentArchiveService.AddModuleResourcesToArchive(archive,
		moduleResources); err != nil {
		return fmt.Errorf("failed to add module resources to component archive: %w", err)
	}

	opts.Out.Write("- Pushing component version\n")
	if !opts.DryRun {
		descriptor, err = s.pushComponentVersion(archive, opts)
		if err != nil {
			return fmt.Errorf("failed to push component version: %w", err)
		}
	} else {
		opts.Out.Write("\tSkipping push due to dry-run mode\n")
		if err = s.ensureComponentVersionDoesNotExist(archive, opts); err != nil {
			return err
		}
	}

	opts.Out.Write("- Generating ModuleTemplate\n")
	if err = s.generateModuleTemplate(moduleConfig,
		descriptor,
		manifestFilePath,
		defaultCRFilePath,
		opts.TemplateOutput); err != nil {
		return fmt.Errorf("failed to generate module template: %w", err)
	}

	return nil
}

func (s *Service) ensureComponentVersionDoesNotExist(archive *comparch.ComponentArchive, opts Options) error {
	exists, err := s.registryService.ExistsComponentVersion(archive,
		opts.Insecure,
		opts.Credentials,
		opts.RegistryURL)
	if err != nil {
		return fmt.Errorf("failed to check if component version exists: %w", err)
	}

	if !exists {
		opts.Out.Write(
			fmt.Sprintf("\tComponent %s in version %s does not exist yet\n", archive.GetName(), archive.GetVersion()))
		return nil
	}

	if opts.OverwriteComponentVersion {
		opts.Out.Write(
			fmt.Sprintf("\tComponent %s in version %s already exists and is overwritten. Use this for testing purposes only.\n",
				archive.GetName(),
				archive.GetVersion()))
		return nil
	}

	return fmt.Errorf("component %s in version %s already exists: %w", archive.GetName(), archive.GetVersion(),
		ErrComponentVersionExists)
}

func (s *Service) pushComponentVersion(archive *comparch.ComponentArchive, opts Options) (*compdesc.ComponentDescriptor,
	error,
) {
	if err := s.registryService.PushComponentVersion(archive,
		opts.Insecure,
		opts.OverwriteComponentVersion,
		opts.Credentials,
		opts.RegistryURL); err != nil {
		return nil, fmt.Errorf("failed to push component archive: %w", err)
	}

	componentVersionAccess, err := s.registryService.GetComponentVersion(archive, opts.Insecure, opts.Credentials,
		opts.RegistryURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get component version: %w", err)
	}

	return componentVersionAccess.GetDescriptor(), nil
}

func (s *Service) generateModuleTemplate(
	moduleConfig *contentprovider.ModuleConfig,
	descriptor *compdesc.ComponentDescriptor,
	manifestFilePath string,
	defaultCRFilePath string,
	templateOutput string,
) error {
	isCRDClusterScoped, err := s.crdParserService.IsCRDClusterScoped(defaultCRFilePath, manifestFilePath)
	if err != nil {
		return fmt.Errorf("failed to determine if CRD is cluster scoped: %w", err)
	}

	var crData []byte
	if defaultCRFilePath != "" {
		crData, err = s.fileSystem.ReadFile(defaultCRFilePath)
		if err != nil {
			return fmt.Errorf("failed to get default CR data: %w", err)
		}
	}

	if err := s.moduleTemplateService.GenerateModuleTemplate(moduleConfig,
		descriptor,
		crData,
		isCRDClusterScoped,
		templateOutput); err != nil {
		return fmt.Errorf("failed to generate module template: %w", err)
	}

	return nil
}

func (s *Service) configureSecScannerConf(descriptor *compdesc.ComponentDescriptor,
	moduleConfig *contentprovider.ModuleConfig, opts Options,
) ([]string, error) {
	opts.Out.Write("- Configuring security scanners config\n")
	securityConfig, err := s.securityConfigService.ParseSecurityConfigData(moduleConfig.Security)
	if err != nil {
		return nil, fmt.Errorf("failed to parse security config data: %w", err)
	}

	err = securityConfig.ValidateBDBAImageTags(moduleConfig.Version)
	if err != nil {
		return nil, fmt.Errorf("failed to validate security config images: %w", err)
	}

	if err = s.securityConfigService.AppendSecurityScanConfig(descriptor, *securityConfig); err != nil {
		return nil, fmt.Errorf("failed to append security scan config: %w", err)
	}
	return securityConfig.BDBA, nil
}

func (s *Service) extractImagesFromManifest(manifestFilePath string, opts Options) ([]string, error) {
	opts.Out.Write("- Extracting images from raw manifest\n")
	images, err := s.manifestService.ExtractImagesFromManifest(manifestFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to extract images from manifest: %w", err)
	}
	return images, nil
}

func addImagesOciArtifactsToDescriptor(descriptor *compdesc.ComponentDescriptor,
	images []string, opts Options,
) error {
	opts.Out.Write("- Adding oci artifacts to component descriptor\n")
	if err := componentdescriptor.AddOciArtifactsToDescriptor(descriptor, images); err != nil {
		return fmt.Errorf("failed to add images to component descriptor: %w", err)
	}
	return nil
}

func (s *Service) cleanupTempFiles(opts Options) {
	if err := s.defaultCRFileResolver.CleanupTempFiles(); err != nil {
		opts.Out.Write(fmt.Sprintf("failed to cleanup temporary default CR files: %v\n", err))
	}
	if err := s.manifestFileResolver.CleanupTempFiles(); err != nil {
		opts.Out.Write(fmt.Sprintf("failed to cleanup temporary manifest files: %v\n", err))
	}
}
