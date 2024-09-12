//go:build e2e

package scaffold_test

import (
	"fmt"
	"os/exec"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func Test_Scaffold(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "'Scaffold' Command Test Suite")
}

// Command wrapper for `modulectl scaffold`

type scaffoldCmd struct {
	moduleName                    string
	moduleVersion                 string
	moduleChannel                 string
	moduleConfigFileFlag          string
	genDefaultCRFlag              string
	genSecurityScannersConfigFlag string
	genManifestFlag               string
	overwrite                     bool
}

func (cmd *scaffoldCmd) execute() error {
	var command *exec.Cmd

	args := []string{"scaffold"}

	if cmd.moduleName != "" {
		args = append(args, "--module-name="+cmd.moduleName)
	}

	if cmd.moduleVersion != "" {
		args = append(args, "--module-version="+cmd.moduleVersion)
	}

	if cmd.moduleChannel != "" {
		args = append(args, "--module-channel="+cmd.moduleChannel)
	}

	if cmd.moduleConfigFileFlag != "" {
		args = append(args, "--module-config="+cmd.moduleConfigFileFlag)
	}

	if cmd.genDefaultCRFlag != "" {
		args = append(args, "--gen-default-cr="+cmd.genDefaultCRFlag)
	}

	if cmd.genSecurityScannersConfigFlag != "" {
		args = append(args, "--gen-security-config="+cmd.genSecurityScannersConfigFlag)
	}

	if cmd.genManifestFlag != "" {
		args = append(args, "--gen-manifest="+cmd.genManifestFlag)
	}

	if cmd.overwrite {
		args = append(args, "--overwrite=true")
	}

	command = exec.Command("modulectl", args...)
	cmdOut, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("scaffold command failed with output: %s and error: %w", cmdOut, err)
	}
	return nil
}

func (cmd *scaffoldCmd) toConfigBuilder() *moduleConfigBuilder {
	res := &moduleConfigBuilder{}
	res.defaults()
	if cmd.moduleName != "" {
		res.withName(cmd.moduleName)
	}
	if cmd.moduleVersion != "" {
		res.withVersion(cmd.moduleVersion)
	}
	if cmd.moduleChannel != "" {
		res.withChannel(cmd.moduleChannel)
	}
	if cmd.genDefaultCRFlag != "" {
		res.withDefaultCRPath(cmd.genDefaultCRFlag)
	}
	if cmd.genSecurityScannersConfigFlag != "" {
		res.withSecurityScannersPath(cmd.genSecurityScannersConfigFlag)
	}
	if cmd.genManifestFlag != "" {
		res.withManifestPath(cmd.genManifestFlag)
	}
	return res
}

// moduleConfigBuilder is used to simplify module.Config creation for testing purposes
type moduleConfigBuilder struct {
	moduleConfig
}

func (mcb *moduleConfigBuilder) get() *moduleConfig {
	res := mcb.moduleConfig
	return &res
}

func (mcb *moduleConfigBuilder) withName(val string) *moduleConfigBuilder {
	mcb.Name = val
	return mcb
}

func (mcb *moduleConfigBuilder) withVersion(val string) *moduleConfigBuilder {
	mcb.Version = val
	return mcb
}

func (mcb *moduleConfigBuilder) withChannel(val string) *moduleConfigBuilder {
	mcb.Channel = val
	return mcb
}

func (mcb *moduleConfigBuilder) withManifestPath(val string) *moduleConfigBuilder {
	mcb.ManifestPath = val
	return mcb
}

func (mcb *moduleConfigBuilder) withDefaultCRPath(val string) *moduleConfigBuilder {
	mcb.DefaultCRPath = val
	return mcb
}

func (mcb *moduleConfigBuilder) withSecurityScannersPath(val string) *moduleConfigBuilder {
	mcb.Security = val
	return mcb
}

func (mcb *moduleConfigBuilder) defaults() *moduleConfigBuilder {
	return mcb.
		withName("kyma-project.io/module/mymodule").
		withVersion("0.0.1").
		withChannel("regular").
		withManifestPath("manifest.yaml")
}

// This is a copy of the moduleConfig struct from internal/scaffold/contentprovider/moduleconfig.go
// to not make the moduleConfig public just for the sake of testing.
// It is expected that the moduleConfig struct will be made public in the future when introducing more commands.
// Once it is public, this struct should be removed.
type moduleConfig struct {
	Name          string            `yaml:"name" comment:"required, the name of the Module"`
	Version       string            `yaml:"version" comment:"required, the version of the Module"`
	Channel       string            `yaml:"channel" comment:"required, channel that should be used in the ModuleTemplate"`
	ManifestPath  string            `yaml:"manifest" comment:"required, relative path or remote URL to the manifests"`
	Mandatory     bool              `yaml:"mandatory" comment:"optional, default=false, indicates whether the module is mandatory to be installed on all clusters"`
	DefaultCRPath string            `yaml:"defaultCR" comment:"optional, relative path or remote URL to a YAML file containing the default CR for the module"`
	ResourceName  string            `yaml:"resourceName" comment:"optional, default={name}-{channel}, when channel is 'none', the default is {name}-{version}, the name for the ModuleTemplate that will be created"`
	Namespace     string            `yaml:"namespace" comment:"optional, default=kcp-system, the namespace where the ModuleTemplate will be deployed"`
	Security      string            `yaml:"security" comment:"optional, name of the security scanners config file"`
	Internal      bool              `yaml:"internal" comment:"optional, default=false, determines whether the ModuleTemplate should have the internal flag or not"`
	Beta          bool              `yaml:"beta" comment:"optional, default=false, determines whether the ModuleTemplate should have the beta flag or not"`
	Labels        map[string]string `yaml:"labels" comment:"optional, additional labels for the ModuleTemplate"`
	Annotations   map[string]string `yaml:"annotations" comment:"optional, additional annotations for the ModuleTemplate"`
}
