package module

import "github.com/kyma-project/modulectl/tools/io"

// Options defines available options for the create module command
type Options struct {
	Out                  io.Out // TODO: Can be extracted one level above
	ModuleConfigFile     string
	GitRemote            string
	RegistryURL          string
	Credentials          string
	TemplateOutput       string
	Insecure             bool
	RegistryCredSelector string
	SecurityScanConfig   string
}
