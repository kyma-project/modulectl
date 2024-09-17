package modulectl

import (
	"fmt"

	"github.com/spf13/cobra"

	createcmd "github.com/kyma-project/modulectl/cmd/modulectl/create"
	scaffoldcmd "github.com/kyma-project/modulectl/cmd/modulectl/scaffold"
	"github.com/kyma-project/modulectl/internal/service/contentprovider"
	"github.com/kyma-project/modulectl/internal/service/create"
	"github.com/kyma-project/modulectl/internal/service/filegenerator"
	"github.com/kyma-project/modulectl/internal/service/filegenerator/reusefilegenerator"
	moduleconfiggenerator "github.com/kyma-project/modulectl/internal/service/moduleconfig/generator"
	moduleconfigreader "github.com/kyma-project/modulectl/internal/service/moduleconfig/reader"
	"github.com/kyma-project/modulectl/internal/service/scaffold"
	"github.com/kyma-project/modulectl/tools/filesystem"
	"github.com/kyma-project/modulectl/tools/yaml"

	_ "embed"
	"github.com/kyma-project/modulectl/internal/service/componentdescriptor"
	"github.com/kyma-project/modulectl/internal/service/git"
)

const (
	moduleConfigKind   = "module-config"
	manifestKind       = "manifest"
	defaultCRKind      = "defaultcr"
	securityConfigKind = "security-config"
)

//go:embed use.txt
var use string

//go:embed short.txt
var short string

//go:embed long.txt
var long string

func NewCmd() (*cobra.Command, error) {
	rootCmd := &cobra.Command{
		Use:   use,
		Short: short,
		Long:  long,
	}

	scaffoldService, err := buildScaffoldService()
	if err != nil {
		return nil, fmt.Errorf("failed to build scaffold service: %w", err)
	}

	scaffoldCmd, err := scaffoldcmd.NewCmd(scaffoldService)
	if err != nil {
		return nil, fmt.Errorf("failed to build scaffold command: %w", err)
	}

	moduleService, err := buildModuleService()
	if err != nil {
		return nil, fmt.Errorf("failed to build module service: %w", err)
	}

	createCmd, err := createcmd.NewCmd(moduleService)
	if err != nil {
		return nil, fmt.Errorf("failed to build create command: %w", err)
	}

	rootCmd.AddCommand(scaffoldCmd)
	rootCmd.AddCommand(createCmd)

	return rootCmd, nil
}

func buildModuleService() (*create.Service, error) {
	fileSystemUtil := &filesystem.Util{}
	tmpFileSystem := filesystem.NewTempFileSystem()

	moduleConfigService, err := moduleconfigreader.NewService(fileSystemUtil, tmpFileSystem)
	if err != nil {
		return nil, fmt.Errorf("failed to create module config service: %w", err)
	}

	gitService := git.NewService()
	gitSourcesService := componentdescriptor.NewGitSourcesService(gitService)
	securityConfigService := componentdescriptor.NewSecurityConfigService(gitService)

	moduleService, err := create.NewService(moduleConfigService, gitSourcesService,
		securityConfigService)
	if err != nil {
		return nil, fmt.Errorf("failed to create module service: %w", err)
	}
	return moduleService, nil
}

func buildScaffoldService() (*scaffold.Service, error) {
	fileSystemUtil := &filesystem.Util{}
	yamlConverter := &yaml.ObjectToYAMLConverter{}

	moduleConfigContentProvider, err := contentprovider.NewModuleConfigProvider(yamlConverter)
	if err != nil {
		return nil, fmt.Errorf("failed to create module config content provider: %w", err)
	}

	moduleConfigFileGenerator, err := filegenerator.NewService(moduleConfigKind, fileSystemUtil,
		moduleConfigContentProvider)
	if err != nil {
		return nil, fmt.Errorf("failed to create module config file generator: %w", err)
	}

	moduleConfigService, err := moduleconfiggenerator.NewService(fileSystemUtil, moduleConfigFileGenerator)
	if err != nil {
		return nil, fmt.Errorf("failed to create module config service: %w", err)
	}

	manifestFileGenerator, err := filegenerator.NewService(manifestKind, fileSystemUtil, contentprovider.NewManifest())
	if err != nil {
		return nil, fmt.Errorf("failed to create manifest file generator: %w", err)
	}

	manifestReuseFileGenerator, err := reusefilegenerator.NewService(manifestKind, fileSystemUtil,
		manifestFileGenerator)
	if err != nil {
		return nil, fmt.Errorf("failed to create manifest reuse file generator: %w", err)
	}

	defaultCRFileGenerator, err := filegenerator.NewService(defaultCRKind, fileSystemUtil,
		contentprovider.NewDefaultCR())
	if err != nil {
		return nil, fmt.Errorf("failed to create default CR file generator: %w", err)
	}

	defaultCRReuseFileGenerator, err := reusefilegenerator.NewService(defaultCRKind, fileSystemUtil,
		defaultCRFileGenerator)
	if err != nil {
		return nil, fmt.Errorf("failed to create default CR reuse file generator: %w", err)
	}

	securityConfigContentProvider, err := contentprovider.NewSecurityConfig(yamlConverter)
	if err != nil {
		return nil, fmt.Errorf("failed to create security config content provider: %w", err)
	}

	securityConfigFileGenerator, err := filegenerator.NewService(securityConfigKind, fileSystemUtil,
		securityConfigContentProvider)
	if err != nil {
		return nil, fmt.Errorf("failed to create security config file generator: %w", err)
	}

	securityConfigReuseFileGenerator, err := reusefilegenerator.NewService(securityConfigKind, fileSystemUtil,
		securityConfigFileGenerator)
	if err != nil {
		return nil, fmt.Errorf("failed to create security config reuse file generator: %w", err)
	}

	scaffoldService, err := scaffold.NewService(
		moduleConfigService,
		manifestReuseFileGenerator,
		defaultCRReuseFileGenerator,
		securityConfigReuseFileGenerator,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create scaffold service: %w", err)
	}

	return scaffoldService, nil
}
