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

	invalidConfigs             = testdataDir + "invalid/"
	duplicateIcons             = invalidConfigs + "duplicate-icons.yaml"
	duplicateResources         = invalidConfigs + "duplicate-resources.yaml"
	missingNameConfig          = invalidConfigs + "missing-name.yaml"
	missingVersionConfig       = invalidConfigs + "missing-version.yaml"
	missingManifestConfig      = invalidConfigs + "missing-manifest.yaml"
	missingDocumentationConfig = invalidConfigs + "missing-documentation.yaml"
	missingRepositoryConfig    = invalidConfigs + "missing-repository.yaml"
	missingIconsConfig         = invalidConfigs + "missing-icons.yaml"
	nonHttpsRepository         = invalidConfigs + "non-https-repository.yaml"
	nonHttpsDocumentation      = invalidConfigs + "non-https-documentation.yaml"
	nonHttpsResource           = invalidConfigs + "non-https-resource.yaml"
	resourceWithoutLink        = invalidConfigs + "resource-without-link.yaml"
	resourceWithoutName        = invalidConfigs + "resource-without-name.yaml"
	iconsWithoutLink           = invalidConfigs + "icons-without-link.yaml"
	iconsWithoutName           = invalidConfigs + "icons-without-name.yaml"
	invalidSecurityConfig      = invalidConfigs + "not-existing-security.yaml"

	validConfigs                  = testdataDir + "valid/"
	minimalConfig                 = validConfigs + "minimal.yaml"
	withAnnotationsConfig         = validConfigs + "with-annotations.yaml"
	withDefaultCrConfig           = validConfigs + "with-defaultcr.yaml"
	withSecurityConfig            = validConfigs + "with-security.yaml"
	withMandatoryConfig           = validConfigs + "with-mandatory.yaml"
	withAssociatedResourcesConfig = validConfigs + "with-associated-resources.yaml"
	withResources                 = validConfigs + "with-resources.yaml"
	withResourcesOverwrite        = validConfigs + "with-resources-overwrite.yaml"
	withManagerConfig             = validConfigs + "with-manager.yaml"
	withNoNamespaceManagerConfig  = validConfigs + "with-manager-no-namespace.yaml"
	withRequiresDowntimeConfig    = validConfigs + "with-requiresDowntime.yaml"
	withInternalConfig            = validConfigs + "with-internal.yaml"
	withBetaConfig                = validConfigs + "with-beta.yaml"
	manifestFileref               = validConfigs + "with-manifest-fileref.yaml"
	defaultCRFileref              = validConfigs + "with-defaultcr-fileref.yaml"

	ociRegistry        = "http://k3d-oci.localhost:5001"
	templateOutputPath = "/tmp/template.yaml"
)

// Command wrapper for `modulectl create`

type createCmd struct {
	registry         string
	output           string
	moduleConfigFile string
	insecure         bool
	overwrite        bool
	dryRun           bool
}

func (cmd *createCmd) execute() error {
	var command *exec.Cmd

	args := []string{"create"}

	if cmd.moduleConfigFile != "" {
		args = append(args, "--config-file="+cmd.moduleConfigFile)
	}

	if cmd.registry != "" {
		args = append(args, "--registry="+cmd.registry)
	}

	if cmd.output != "" {
		args = append(args, "--output="+cmd.output)
	}

	if cmd.insecure {
		args = append(args, "--insecure")
	}

	if cmd.overwrite {
		args = append(args, "--overwrite")
	}

	if cmd.dryRun {
		args = append(args, "--dry-run")
	}

	println(" >>> Executing command: modulectl", strings.Join(args, " "))

	command = exec.Command("modulectl", args...)
	cmdOut, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("create command failed with output: %s and error: %w", cmdOut, err)
	}
	return nil
}
