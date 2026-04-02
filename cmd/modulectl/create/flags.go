package create

import (
	"github.com/spf13/pflag"

	"github.com/kyma-project/modulectl/internal/service/create"
)

const (
	ConfigFileFlagName    = "config-file"
	configFileFlagShort   = "c"
	ConfigFileFlagDefault = "module-config.yaml"
	configFileFlagUsage   = "Specifies the path to the module configuration file."

	TemplateOutputFlagName    = "output"
	templateOutputFlagShort   = "o"
	TemplateOutputFlagDefault = "template.yaml"
	templateOutputFlagUsage   = `Path to write the ModuleTemplate file to (default "template.yaml").`

	ModuleSourcesGitDirectoryFlagName    = "module-sources-git-directory"
	ModuleSourcesGitDirectoryFlagDefault = "."
	ModuleSourcesGitDirectoryFlagUsage   = "Path to the directory containing the module sources. If not set, the current directory is used. The directory must contain a valid Git repository."

	OutputConstructorFileFlagName    = "output-constructor-file"
	OutputConstructorFileFlagDefault = "component-constructor.yaml"
	OutputConstructorFileFlagUsage   = "Path to write the component constructor file to (default \"component-constructor.yaml\")."
)

func parseFlags(flags *pflag.FlagSet, opts *create.Options) {
	flags.StringVarP(&opts.ConfigFile,
		ConfigFileFlagName,
		configFileFlagShort,
		ConfigFileFlagDefault,
		configFileFlagUsage)
	flags.StringVarP(&opts.TemplateOutput,
		TemplateOutputFlagName,
		templateOutputFlagShort,
		TemplateOutputFlagDefault,
		templateOutputFlagUsage)
	flags.StringVar(&opts.ModuleSourcesGitDirectory,
		ModuleSourcesGitDirectoryFlagName,
		ModuleSourcesGitDirectoryFlagDefault,
		ModuleSourcesGitDirectoryFlagUsage)

	// Feature toggle flag for skipping version validation, should be removed once all module confirmed in the internal backlog issue: 7573
	flags.BoolVar(&opts.SkipVersionValidation,
		"skip-version-validation",
		true,
		"Skipping image and ocm version validation")

	flags.StringVar(&opts.OutputConstructorFile,
		OutputConstructorFileFlagName,
		OutputConstructorFileFlagDefault,
		OutputConstructorFileFlagUsage)
}
