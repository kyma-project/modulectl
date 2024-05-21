package create

import (
	_ "embed"
	"fmt"

	scaffoldcmd "github.com/kyma-project/modulectl/cmd/modulectl/create/scaffold"
	"github.com/kyma-project/modulectl/internal/service/contentprovider"
	"github.com/kyma-project/modulectl/internal/service/filegenerator"
	"github.com/kyma-project/modulectl/internal/service/filegenerator/reusefilegenerator"
	"github.com/kyma-project/modulectl/internal/service/moduleconfig"
	scaffoldsvc "github.com/kyma-project/modulectl/internal/service/scaffold"
	"github.com/kyma-project/modulectl/tools/filesystem"
	"github.com/kyma-project/modulectl/tools/yaml"
	"github.com/spf13/cobra"
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

	svc, err := buildScaffoldService()
	if err != nil {
		return nil, fmt.Errorf("failed to build scaffold service: %w", err)
	}

	cmd, err := scaffoldcmd.NewCmd(svc)
	if err != nil {
		return nil, fmt.Errorf("failed to build scaffold command: %w", err)
	}

	rootCmd.AddCommand(cmd)
	return rootCmd, nil
}

func buildScaffoldService() (*scaffoldsvc.Service, error) {
	fileSystemUtil := &filesystem.FileSystemUtil{}
	yamlConverter := &yaml.ObjectToYAMLConverter{}

	moduleConfigContentProvider, err := contentprovider.NewModuleConfig(yamlConverter)
	if err != nil {
		return nil, err
	}

	moduleConfigFileGenerator, err := filegenerator.NewService(moduleConfigKind, fileSystemUtil, moduleConfigContentProvider)
	if err != nil {
		return nil, err
	}

	moduleConfigService, err := moduleconfig.NewService(fileSystemUtil, moduleConfigFileGenerator)
	if err != nil {
		return nil, err
	}

	manifestFileGenerator, err := filegenerator.NewService(manifestKind, fileSystemUtil, contentprovider.NewManifest())
	if err != nil {
		return nil, err
	}

	manifestReuseFileGenerator, err := reusefilegenerator.NewService(manifestKind, fileSystemUtil, manifestFileGenerator)
	if err != nil {
		return nil, err
	}

	defaultCRFileGenerator, err := filegenerator.NewService(defaultCRKind, fileSystemUtil, contentprovider.NewDefaultCR())
	if err != nil {
		return nil, err
	}

	defaultCRReuseFileGenerator, err := reusefilegenerator.NewService(defaultCRKind, fileSystemUtil, defaultCRFileGenerator)
	if err != nil {
		return nil, err
	}

	securitConfigContentProvider, err := contentprovider.NewSecurityConfig(yamlConverter)
	if err != nil {
		return nil, err
	}

	securityConfigFileGenerator, err := filegenerator.NewService(securityConfigKind, fileSystemUtil, securitConfigContentProvider)
	if err != nil {
		return nil, err
	}

	securityConfigReuseFileGenerator, err := reusefilegenerator.NewService(securityConfigKind, fileSystemUtil, securityConfigFileGenerator)
	if err != nil {
		return nil, err
	}

	scaffoldService, err := scaffoldsvc.NewService(
		moduleConfigService,
		manifestReuseFileGenerator,
		defaultCRReuseFileGenerator,
		securityConfigReuseFileGenerator,
	)
	if err != nil {
		return nil, err
	}

	return scaffoldService, nil
}
