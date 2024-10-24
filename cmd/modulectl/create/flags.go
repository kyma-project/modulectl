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

	CredentialsFlagName    = "registry-credentials" //nolint:gosec // Not hardcoded credentials, rather just flag name
	CredentialsFlagDefault = ""
	credentialsFlagUsage   = "Basic authentication credentials for the given repository in the <user:password> format."

	GitRemoteFlagName    = "git-remote"
	GitRemoteFlagDefault = ""
	gitRemoteFlagUsage   = "Specifies the URL of the module's GitHub repository. "

	InsecureFlagName    = "insecure"
	InsecureFlagDefault = false
	insecureFlagUsage   = "Uses an insecure connection to access the registry."

	TemplateOutputFlagName    = "output"
	templateOutputFlagShort   = "o"
	TemplateOutputFlagDefault = "template.yaml"
	templateOutputFlagUsage   = `Path to write the ModuleTemplate file to, if the module is uploaded to a registry (default "template.yaml").`

	RegistryURLFlagName    = "registry"
	registryFlagShort      = "r"
	RegistryURLFlagDefault = ""
	registryURLFlagUsage   = "Context URL of the repository. The repository URL will be automatically added to the repository contexts in the module descriptor."

	//nolint:gosec // Not hardcoded credentials, rather just flag name
	RegistryCredSelectorFlagName    = "registry-cred-selector"
	RegistryCredSelectorFlagDefault = ""
	//nolint:gosec // Not hardcoded credentials, rather just flag name
	registryCredSelectorFlagUsage = `Label selector to identify an externally created Secret of type "kubernetes.io/dockerconfigjson". It allows the image to be accessed in private image registries. It can be used when you push your module to a registry with authenticated access. For example, "label1=value1,label2=value2".`
)

func parseFlags(flags *pflag.FlagSet, opts *create.Options) {
	flags.StringVarP(&opts.ConfigFile, ConfigFileFlagName, configFileFlagShort, ConfigFileFlagDefault,
		configFileFlagUsage)
	flags.StringVar(&opts.Credentials, CredentialsFlagName, CredentialsFlagDefault,
		credentialsFlagUsage)
	flags.StringVar(&opts.GitRemote, GitRemoteFlagName, GitRemoteFlagDefault, gitRemoteFlagUsage)
	flags.BoolVar(&opts.Insecure, InsecureFlagName, InsecureFlagDefault, insecureFlagUsage)
	flags.StringVarP(&opts.TemplateOutput, TemplateOutputFlagName, templateOutputFlagShort, TemplateOutputFlagDefault,
		templateOutputFlagUsage)
	flags.StringVarP(&opts.RegistryURL, RegistryURLFlagName, registryFlagShort, RegistryURLFlagDefault, registryURLFlagUsage)
	flags.StringVar(&opts.RegistryCredSelector, RegistryCredSelectorFlagName, RegistryCredSelectorFlagDefault,
		registryCredSelectorFlagUsage)
}
