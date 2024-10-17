//go:build e2e

package create_test

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func Test_Create(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "'Create' Command Test Suite")
}

const (
	testdataDir = "./testdata/moduleconfig/"

	invalidConfigs        = testdataDir + "invalid/"
	missingNameConfig     = invalidConfigs + "missing-name.yaml"
	missingChannelConfig  = invalidConfigs + "missing-channel.yaml"
	missingVersionConfig  = invalidConfigs + "missing-version.yaml"
	missingManifestConfig = invalidConfigs + "missing-manifest.yaml"
	missingInfoConfig     = invalidConfigs + "missing-info.yaml"

	validConfigs          = testdataDir + "valid/"
	minimalConfig         = validConfigs + "minimal.yaml"
	withAnnotationsConfig = validConfigs + "with-annotations.yaml"
	withDefaultCrConfig   = validConfigs + "with-defaultcr.yaml"
	withSecurityConfig    = validConfigs + "with-security.yaml"
	withMandatoryConfig   = validConfigs + "with-mandatory.yaml"

	ociRegistry        = "http://k3d-oci.localhost:5001"
	templateOutputPath = "/tmp/template.yaml"
	gitRemote          = "https://github.com/kyma-project/template-operator"
)

// Command wrapper for `modulectl create`

type createCmd struct {
	registry         string
	output           string
	moduleConfigFile string
	gitRemote        string
	insecure         bool
}

func (cmd *createCmd) execute() error {
	var command *exec.Cmd

	args := []string{"create"}

	if cmd.moduleConfigFile != "" {
		args = append(args, "--module-config-file="+cmd.moduleConfigFile)
	}

	if cmd.registry != "" {
		args = append(args, "--registry="+cmd.registry)
	}

	if cmd.output != "" {
		args = append(args, "--output="+cmd.output)
	}

	if cmd.gitRemote != "" {
		args = append(args, "--git-remote="+cmd.gitRemote)
	}

	if cmd.insecure {
		args = append(args, "--insecure")
	}

	println(" >>> Executing command: modulectl", strings.Join(args, " "))

	command = exec.Command("modulectl", args...)
	cmdOut, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("create command failed with output: %s and error: %w", cmdOut, err)
	}
	return nil
}
