package scaffold

import (
	"github.com/spf13/pflag"

	"github.com/kyma-project/modulectl/internal/service/scaffold"
)

const (
	DirectoryFlagName    = "directory"
	directoryFlagShort   = "d"
	DirectoryFlagDefault = "./"
	directoryFlagUsage   = `Specifies the target directory where the scaffolding shall be generated (default "./").`

	ModuleConfigFileFlagName    = "module-config"
	ModuleConfigFileFlagDefault = "scaffold-module-config.yaml"
	moduleConfigFileFlagUsage   = `Specifies the name of the generated module configuration file (default "scaffold-module-config.yaml").`

	ModuleConfigFileOverwriteFlagName    = "overwrite"
	moduleConfigFileOverwriteFlagShort   = "o"
	ModuleConfigFileOverwriteFlagDefault = false
	moduleConfigFileOverwriteFlagUsage   = "Specifies if the command overwrites an existing module configuration file."

	ManifestFileFlagName    = "gen-manifest"
	ManifestFileFlagDefault = "manifest.yaml"
	manifestFileFlagUsage   = `Specifies the manifest in the generated module config. A blank manifest file is generated if it doesn't exist (default "manifest.yaml").`

	DefaultCRFlagName         = "gen-default-cr"
	DefaultCRFlagDefault      = ""
	DefaultCRFlagNoOptDefault = "default-cr.yaml"
	defaultCRFlagUsage        = `Specifies the default CR in the generated module config. A blank default CR file is generated if it doesn't exist (default "default-cr.yaml").`

	SecurityConfigFileFlagName         = "gen-security-config"
	SecurityConfigFileFlagDefault      = ""
	SecurityConfigFileFlagNoOptDefault = "sec-scanners-config.yaml"
	securityConfigFileFlagUsage        = `Specifies the security file in the generated module config. A scaffold security config file is generated if it doesn't exist (default "sec-scanners-config.yaml").`

	ModuleNameFlagName    = "module-name"
	ModuleNameFlagDefault = "kyma-project.io/module/mymodule"
	moduleNameFlagUsage   = `Specifies the module name in the generated config file (default "kyma-project.io/module/mymodule").`

	ModuleVersionFlagName    = "module-version"
	ModuleVersionFlagDefault = "0.0.1"
	moduleVersionFlagUsage   = `Specifies the module version in the generated module config file (default "0.0.1").`

	ModuleChannelFlagName    = "module-channel"
	ModuleChannelFlagDefault = "regular"
	moduleChannelFlagUsage   = `Specifies the module channel in the generated module config file (default "regular").`
)

func parseFlags(flags *pflag.FlagSet, opts *scaffold.Options) {
	flags.StringVarP(&opts.Directory, DirectoryFlagName, directoryFlagShort, DirectoryFlagDefault, directoryFlagUsage)
	flags.StringVar(&opts.ModuleConfigFileName, ModuleConfigFileFlagName, ModuleConfigFileFlagDefault, moduleConfigFileFlagUsage)
	flags.BoolVarP(&opts.ModuleConfigFileOverwrite, ModuleConfigFileOverwriteFlagName, moduleConfigFileOverwriteFlagShort, ModuleConfigFileOverwriteFlagDefault, moduleConfigFileOverwriteFlagUsage)
	flags.StringVar(&opts.ManifestFileName, ManifestFileFlagName, ManifestFileFlagDefault, manifestFileFlagUsage)
	flags.StringVar(&opts.DefaultCRFileName, DefaultCRFlagName, DefaultCRFlagDefault, defaultCRFlagUsage)
	flags.StringVar(&opts.SecurityConfigFileName, SecurityConfigFileFlagName, SecurityConfigFileFlagDefault, securityConfigFileFlagUsage)
	flags.StringVar(&opts.ModuleName, ModuleNameFlagName, ModuleNameFlagDefault, moduleNameFlagUsage)
	flags.StringVar(&opts.ModuleVersion, ModuleVersionFlagName, ModuleVersionFlagDefault, moduleVersionFlagUsage)
	flags.StringVar(&opts.ModuleChannel, ModuleChannelFlagName, ModuleChannelFlagDefault, moduleChannelFlagUsage)

	flags.Lookup(SecurityConfigFileFlagName).NoOptDefVal = SecurityConfigFileFlagNoOptDefault
	flags.Lookup(DefaultCRFlagName).NoOptDefVal = DefaultCRFlagNoOptDefault
}
