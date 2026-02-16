package scaffold

import (
	"fmt"
	"path"

	commonerrors "github.com/kyma-project/modulectl/internal/common/errors"
	"github.com/kyma-project/modulectl/internal/common/types"
	"github.com/kyma-project/modulectl/internal/service/contentprovider"
	iotools "github.com/kyma-project/modulectl/tools/io"
)

type ModuleConfigService interface {
	FileGeneratorService
	ForceExplicitOverwrite(directory, moduleConfigFileName string, overwrite bool) error
}

type FileGeneratorService interface {
	GenerateFile(out iotools.Out, path string, args types.KeyValueArgs) error
}

type Service struct {
	moduleConfigService   ModuleConfigService
	manifestService       FileGeneratorService
	defaultCRService      FileGeneratorService
	securityConfigService FileGeneratorService
}

func NewService(moduleConfigService ModuleConfigService,
	manifestService FileGeneratorService,
	defaultCRService FileGeneratorService,
	securityConfigService FileGeneratorService,
) (*Service, error) {
	if moduleConfigService == nil {
		return nil, fmt.Errorf("moduleConfigService must not be nil: %w", commonerrors.ErrInvalidArg)
	}

	if manifestService == nil {
		return nil, fmt.Errorf("manifestService must not be nil: %w", commonerrors.ErrInvalidArg)
	}

	if defaultCRService == nil {
		return nil, fmt.Errorf("defaultCRService must not be nil: %w", commonerrors.ErrInvalidArg)
	}

	if securityConfigService == nil {
		return nil, fmt.Errorf("securityConfigService must not be nil: %w", commonerrors.ErrInvalidArg)
	}

	return &Service{
		moduleConfigService:   moduleConfigService,
		manifestService:       manifestService,
		defaultCRService:      defaultCRService,
		securityConfigService: securityConfigService,
	}, nil
}

func (s *Service) Run(opts Options) error {
	if err := opts.Validate(); err != nil {
		return fmt.Errorf("validation failed for options: %w", err)
	}

	if err := s.moduleConfigService.ForceExplicitOverwrite(opts.Directory, opts.ModuleConfigFileName,
		opts.ModuleConfigFileOverwrite); err != nil {
		return fmt.Errorf(
			"failed to force explicit overwrite for file %q in directory %q: %w",
			opts.ModuleConfigFileName,
			opts.Directory,
			err,
		)
	}

	manifestFilePath := path.Join(opts.Directory, opts.ManifestFileName)
	if err := s.manifestService.GenerateFile(opts.Out, manifestFilePath, nil); err != nil {
		return fmt.Errorf("failed to generate manifest file %q at %q: %w", opts.ManifestFileName, manifestFilePath, err)
	}

	var defaultCRFilePath string
	if opts.defaultCRFileNameConfigured() {
		defaultCRFilePath = path.Join(opts.Directory, opts.DefaultCRFileName)
		if err := s.defaultCRService.GenerateFile(opts.Out, defaultCRFilePath, nil); err != nil {
			return fmt.Errorf(
				"failed to generate default CR file %q at %q: %w",
				opts.DefaultCRFileName,
				defaultCRFilePath,
				err,
			)
		}
	}

	var securityConfigFilePath string
	if opts.securityConfigFileNameConfigured() {
		securityConfigFilePath = path.Join(opts.Directory, opts.SecurityConfigFileName)
		if err := s.securityConfigService.GenerateFile(
			opts.Out,
			securityConfigFilePath,
			types.KeyValueArgs{contentprovider.ArgModuleName: opts.ModuleName}); err != nil {
			return fmt.Errorf(
				"failed to generate security config file %q at %q: %w",
				opts.SecurityConfigFileName,
				securityConfigFilePath,
				err,
			)
		}
	}

	moduleConfigFilePath := path.Join(opts.Directory, opts.ModuleConfigFileName)
	if err := s.moduleConfigService.GenerateFile(
		opts.Out,
		moduleConfigFilePath,
		types.KeyValueArgs{
			contentprovider.ArgModuleName:         opts.ModuleName,
			contentprovider.ArgModuleVersion:      opts.ModuleVersion,
			contentprovider.ArgTeam:               opts.Team,
			contentprovider.ArgManifestFile:       opts.ManifestFileName,
			contentprovider.ArgDefaultCRFile:      defaultCRFilePath,
			contentprovider.ArgSecurityConfigFile: securityConfigFilePath,
		}); err != nil {
		return fmt.Errorf(
			"failed to generate module config file %q at %q: %w",
			opts.ModuleConfigFileName,
			moduleConfigFilePath,
			err,
		)
	}

	return nil
}
