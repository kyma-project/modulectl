package create

import (
	"github.com/spf13/pflag"

	"github.com/kyma-project/modulectl/internal/service/create"
)

const (
	ModuleConfigFileFlagName    = "module-config-file"
	ModuleConfigFileFlagDefault = "module-config.yaml"
	moduleConfigFileFlagUsage   = "Specifies the module configuration file."

	CredentialsFlagName    = "credentials"
	credentialsFlagShort   = "c"
	CredentialsFlagDefault = ""
	credentialsFlagUsage   = "Basic authentication credentials for the given repository in the <user:password> format."

	GitRemoteFlagName    = "git-remote"
	GitRemoteFlagDefault = "origin"
	gitRemoteFlagUsage   = `Specifies the remote name of the wanted GitHub repository. For example "origin" or "upstream".`

	InsecureFlagName    = "insecure"
	InsecureFlagDefault = false
	insecureFlagUsage   = "Uses an insecure connection to access the registry."

	TemplateOutputFlagName    = "output"
	templateOutputFlagShort   = "o"
	TemplateOutputFlagDefault = "template.yaml"
	templateOutputFlagUsage   = `File to write the module template if the module is uploaded to a registry.`

	RegistryURLFlagName    = "registry"
	RegistryURLFlagDefault = ""
	registryURLFlagUsage   = "Context URL of the repository. The repository URL will be automatically added to the repository contexts in the module descriptor."

	//nolint:gosec // Not hardcoded credentials, rather just flag name
	RegistryCredSelectorFlagName    = "registry-cred-selector"
	RegistryCredSelectorFlagDefault = ""
	//nolint:gosec // Not hardcoded credentials, rather just flag name
	registryCredSelectorFlagUsage = `Label selector to identify an externally created Secret of type "kubernetes.io/dockerconfigjson". It allows the image to be accessed in private image registries. It can be used when you push your module to a registry with authenticated access. For example, "label1=value1,label2=value2".`

	SecScannersConfigFlagName    = "sec-scanners-config"
	SecScannersConfigFlagDefault = "sec-scanners-config.yaml"
	secScannersConfigFlagUsage   = `Path to the file holding the security scan configuration.`
)

func parseFlags(flags *pflag.FlagSet, opts *create.Options) {
	flags.StringVar(&opts.ModuleConfigFile, ModuleConfigFileFlagName, ModuleConfigFileFlagDefault, moduleConfigFileFlagUsage)
	flags.StringVarP(&opts.Credentials, CredentialsFlagName, credentialsFlagShort, CredentialsFlagDefault, credentialsFlagUsage)
	flags.StringVar(&opts.GitRemote, GitRemoteFlagName, GitRemoteFlagDefault, gitRemoteFlagUsage)
	flags.BoolVar(&opts.Insecure, InsecureFlagName, InsecureFlagDefault, insecureFlagUsage)
	flags.StringVarP(&opts.TemplateOutput, TemplateOutputFlagName, templateOutputFlagShort, TemplateOutputFlagDefault, templateOutputFlagUsage)
	flags.StringVar(&opts.RegistryURL, RegistryURLFlagName, RegistryURLFlagDefault, registryURLFlagUsage)
	flags.StringVar(&opts.RegistryCredSelector, RegistryCredSelectorFlagName, RegistryCredSelectorFlagDefault, registryCredSelectorFlagUsage)
	flags.StringVar(&opts.SecScannerConfig, SecScannersConfigFlagName, SecScannersConfigFlagDefault, secScannersConfigFlagUsage)
}
