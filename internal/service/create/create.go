package create

import (
	"fmt"
	"path"
	"strconv"

	"github.com/kyma-project/lifecycle-manager/api/shared"

	"github.com/kyma-project/modulectl/internal/common"
	commonerrors "github.com/kyma-project/modulectl/internal/common/errors"
	"github.com/kyma-project/modulectl/internal/common/types"
	"github.com/kyma-project/modulectl/internal/common/types/component"
	"github.com/kyma-project/modulectl/internal/service/contentprovider"
)

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

type GitSourcesService interface {
	AddGitSourcesToConstructor(constructor *component.Constructor, gitRepoPath, gitRepoURL string) error
}

type ComponentConstructorService interface {
	AddImagesToConstructor(componentConstructor *component.Constructor,
		images []string,
	) error
	AddResources(componentConstructor *component.Constructor,
		resourcePaths *types.ResourcePaths,
	) error
	CreateConstructorFile(componentConstructor *component.Constructor,
		outputFile string,
	) error
	SetComponentLabel(componentConstructor *component.Constructor, name, value string)
	SetResponsiblesLabel(componentConstructor *component.Constructor, team string)
}

type ModuleTemplateService interface {
	GenerateModuleTemplate(moduleConfig *contentprovider.ModuleConfig,
		data []byte,
		isCrdClusterScoped bool,
		templateOutput string,
	) error
}

type CRDParserService interface {
	IsCRDClusterScoped(paths *types.ResourcePaths) (bool, error)
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
	componentConstructorService ComponentConstructorService
	moduleTemplateService       ModuleTemplateService
	crdParserService            CRDParserService
	imageVersionVerifierService ImageVersionVerifierService
	manifestService             ManifestService
	manifestFileResolver        FileResolver
	defaultCRFileResolver       FileResolver
	fileSystem                  FileSystem
}

func NewService(moduleConfigService ModuleConfigService,
	gitSourcesService GitSourcesService,
	componentConstructorService ComponentConstructorService,
	moduleTemplateService ModuleTemplateService,
	crdParserService CRDParserService,
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

	if componentConstructorService == nil {
		return nil, fmt.Errorf("componentConstructorService must not be nil: %w", commonerrors.ErrInvalidArg)
	}

	if moduleTemplateService == nil {
		return nil, fmt.Errorf("moduleTemplateService must not be nil: %w", commonerrors.ErrInvalidArg)
	}

	if crdParserService == nil {
		return nil, fmt.Errorf("crdParserService must not be nil: %w", commonerrors.ErrInvalidArg)
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
		componentConstructorService: componentConstructorService,
		moduleTemplateService:       moduleTemplateService,
		crdParserService:            crdParserService,
		imageVersionVerifierService: imageVersionVerifierService,
		manifestService:             manifestService,
		manifestFileResolver:        manifestFileResolver,
		defaultCRFileResolver:       defaultCRFileResolver,
		fileSystem:                  fileSystem,
	}, nil
}

func (s *Service) Run(opts Options) (rErr error) { //nolint:nonamedreturns // named return to detect error in defer
	if err := opts.Validate(); err != nil {
		return err
	}

	moduleConfig, err := s.moduleConfigService.ParseAndValidateModuleConfig(opts.ConfigFile)
	if err != nil {
		return fmt.Errorf("failed to parse module config: %w", err)
	}

	defer func() {
		if rErr != nil { // only clean up if an error occurs
			s.cleanupTempFiles(opts)
		}
	}()

	configFilePath := path.Dir(opts.ConfigFile)
	// If the manifest is a local file reference, it's entry in the module config file will be relative to the module
	// config file location (usually the same directory).
	manifestFilePath, err := s.manifestFileResolver.Resolve(moduleConfig.Manifest, configFilePath)
	if err != nil {
		return fmt.Errorf("failed to resolve manifest file: %w", err)
	}

	var defaultCRFilePath string
	if !moduleConfig.DefaultCR.IsEmpty() {
		// If the defaultCR is a local file reference, it's entry in the module config file will be relative to the
		// module config file location (usually the same directory).
		defaultCRFilePath, err = s.defaultCRFileResolver.Resolve(moduleConfig.DefaultCR, configFilePath)
		if err != nil {
			return fmt.Errorf("failed to resolve default CR file: %w", err)
		}
	}

	resourcePaths := types.NewResourcePaths(defaultCRFilePath, manifestFilePath, opts.TemplateOutput)

	if err = s.createComponentConstructor(moduleConfig, resourcePaths, opts); err != nil {
		return fmt.Errorf("failed to process component: %w", err)
	}
	return nil
}

func (s *Service) createComponentConstructor(moduleConfig *contentprovider.ModuleConfig,
	resourcePaths *types.ResourcePaths,
	opts Options,
) error {
	constructor := component.NewConstructor(moduleConfig.Name, moduleConfig.Version)

	if err := s.gitSourcesService.AddGitSourcesToConstructor(constructor, opts.ModuleSourcesGitDirectory,
		moduleConfig.Repository); err != nil {
		return fmt.Errorf("failed to add git sources to constructor: %w", err)
	}

	images, err := s.extractImagesFromManifest(resourcePaths.RawManifest, opts)
	if err != nil {
		return fmt.Errorf("failed to extract images from manifest: %w", err)
	}

	if !opts.SkipVersionValidation {
		if err := s.imageVersionVerifierService.VerifyModuleResources(moduleConfig,
			resourcePaths.RawManifest); err != nil {
			return fmt.Errorf("failed to verify module resources: %w", err)
		}
	}

	opts.Out.Write("- Adding oci artifacts to component descriptor\n")
	if err := s.componentConstructorService.AddImagesToConstructor(constructor, images); err != nil {
		return fmt.Errorf("failed to add images to component constructor: %w", err)
	}

	opts.Out.Write("- Creating module template\n")
	err = s.createModuleTemplate(moduleConfig, resourcePaths)
	if err != nil {
		return fmt.Errorf("failed to create module template: %w", err)
	}

	opts.Out.Write("- Generating module resources\n")
	if err = s.componentConstructorService.AddResources(constructor, resourcePaths); err != nil {
		return fmt.Errorf("failed to add resources to component constructor: %w", err)
	}

	opts.Out.Write("- Setting OCM Component labels\n")
	s.componentConstructorService.SetComponentLabel(constructor,
		shared.BetaLabel, strconv.FormatBool(moduleConfig.Beta))

	s.componentConstructorService.SetComponentLabel(constructor,
		shared.InternalLabel, strconv.FormatBool(moduleConfig.Internal))

	s.componentConstructorService.SetComponentLabel(constructor,
		common.RequiresDowntimeLabelKey,
		strconv.FormatBool(moduleConfig.RequiresDowntime))

	isCRDClusterScoped, err := s.crdParserService.IsCRDClusterScoped(resourcePaths)
	if err != nil {
		return fmt.Errorf("failed to determine if CRD is cluster scoped: %w", err)
	}
	s.componentConstructorService.SetComponentLabel(constructor,
		shared.IsClusterScopedAnnotation,
		strconv.FormatBool(isCRDClusterScoped))

	// Add security scan and responsibles labels if security scan is enabled
	securityScanEnabled := getSecurityScanEnabled(moduleConfig)
	if securityScanEnabled {
		s.componentConstructorService.SetComponentLabel(constructor,
			common.SecurityScanLabelKey, common.SecurityScanEnabledValue)
		// Add responsibles label with team information
		s.componentConstructorService.SetResponsiblesLabel(constructor, moduleConfig.Team)
	}

	opts.Out.Write("- Creating component constructor file\n")
	if err = s.componentConstructorService.CreateConstructorFile(constructor,
		opts.OutputConstructorFile); err != nil {
		return fmt.Errorf("failed to create constructor file: %w", err)
	}
	return nil
}

func (s *Service) extractImagesFromManifest(manifestFilePath string, opts Options) ([]string, error) {
	opts.Out.Write("- Extracting images from raw manifest\n")
	images, err := s.manifestService.ExtractImagesFromManifest(manifestFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to extract images from manifest: %w", err)
	}
	return images, nil
}

// getSecurityScanEnabled returns true if securityScanEnabled is nil or true, false if explicitly set to false.
func getSecurityScanEnabled(moduleConfig *contentprovider.ModuleConfig) bool {
	if moduleConfig.SecurityScanEnabled == nil {
		return true // default is enabled
	}
	return *moduleConfig.SecurityScanEnabled
}

func (s *Service) cleanupTempFiles(opts Options) {
	if err := s.defaultCRFileResolver.CleanupTempFiles(); err != nil {
		opts.Out.Write(fmt.Sprintf("failed to cleanup temporary default CR files: %v\n", err))
	}
	if err := s.manifestFileResolver.CleanupTempFiles(); err != nil {
		opts.Out.Write(fmt.Sprintf("failed to cleanup temporary manifest files: %v\n", err))
	}
}

func (s *Service) createModuleTemplate(
	moduleConfig *contentprovider.ModuleConfig,
	resourcePaths *types.ResourcePaths,
) error {
	isCRDClusterScoped, err := s.crdParserService.IsCRDClusterScoped(resourcePaths)
	if err != nil {
		return fmt.Errorf("failed to determine if CRD is cluster scoped: %w", err)
	}

	var crData []byte
	if resourcePaths.DefaultCR != "" {
		crData, err = s.fileSystem.ReadFile(resourcePaths.DefaultCR)
		if err != nil {
			return fmt.Errorf("failed to get default CR data: %w", err)
		}
	}

	if err := s.moduleTemplateService.GenerateModuleTemplate(moduleConfig,
		crData,
		isCRDClusterScoped,
		resourcePaths.ModuleTemplate); err != nil {
		return fmt.Errorf("failed to generate module template: %w", err)
	}

	return nil
}
