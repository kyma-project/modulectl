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
	missingTeamConfig          = invalidConfigs + "missing-team.yaml"
	missingIconsConfig         = invalidConfigs + "missing-icons.yaml"
	nonHttpsRepository         = invalidConfigs + "non-https-repository.yaml"
	nonHttpsDocumentation      = invalidConfigs + "non-https-documentation.yaml"
	nonHttpsResource           = invalidConfigs + "non-https-resource.yaml"
	resourceWithoutLink        = invalidConfigs + "resource-without-link.yaml"
	resourceWithoutName        = invalidConfigs + "resource-without-name.yaml"
	iconsWithoutLink           = invalidConfigs + "icons-without-link.yaml"
	iconsWithoutName           = invalidConfigs + "icons-without-name.yaml"
	invalidSecurityConfig      = invalidConfigs + "not-existing-security.yaml"
	invalidSecurityConfigImage = invalidConfigs + "with-security.yaml"
	withManifestLatestMainTags = invalidConfigs + "with-manifest-image-latest-or-main-tags.yaml"

	validConfigs                  = testdataDir + "valid/"
	minimalConfig                 = validConfigs + "minimal.yaml"
	withAnnotationsConfig         = validConfigs + "with-annotations.yaml"
	withDefaultCrConfig           = validConfigs + "with-defaultcr.yaml"
	withSecurityConfig            = validConfigs + "with-security.yaml"
	withAssociatedResourcesConfig = validConfigs + "with-associated-resources.yaml"
	withResources                 = validConfigs + "with-resources.yaml"
	withResourcesOverwrite        = validConfigs + "with-resources-overwrite.yaml"
	withManagerConfig             = validConfigs + "with-manager.yaml"
	withNoNamespaceManagerConfig  = validConfigs + "with-manager-no-namespace.yaml"
	withRequiresDowntimeConfig    = validConfigs + "with-requiresDowntime.yaml"
	withInternalConfig            = validConfigs + "with-internal.yaml"
	withBetaConfig                = validConfigs + "with-beta.yaml"
	defaultCRFileref              = validConfigs + "with-defaultcr-fileref.yaml"
	manifestFileref               = validConfigs + "with-manifest-fileref.yaml"
	withManifestContainers        = validConfigs + "with-manifest-containers.yaml"
	withManifestInitContainers    = validConfigs + "with-manifest-init-containers.yaml"
	withManifestEnvVariables      = validConfigs + "with-manifest-env-variables.yaml"
	withManifestShaDigest         = validConfigs + "with-manifest-sha-digest.yaml"
	withManifestAndSecurity       = validConfigs + "with-manifest-and-security.yaml"
	withManifestNoImages          = validConfigs + "with-manifest-no-deployment-statefulset.yaml"
	withSecurityScanDisabled      = validConfigs + "with-securityScanEnabled-false.yaml"
	withSecurityScanEnabled       = validConfigs + "with-securityScanEnabled-true.yaml"

	templateOutputPath   = "/tmp/template.yaml"
	templateOperatorPath = "../../../"
)

// Command wrapper for `modulectl create`

type createCmd struct {
	output                    string
	moduleConfigFile          string
	moduleSourcesGitDirectory string
	skipVersionValidation     bool
	outputConstructorFile     string
}

func (cmd *createCmd) execute() error {
	var command *exec.Cmd

	args := []string{"create"}

	if cmd.moduleConfigFile != "" {
		args = append(args, "--config-file="+cmd.moduleConfigFile)
	}

	if cmd.output != "" {
		args = append(args, "--output="+cmd.output)
	}

	if cmd.moduleSourcesGitDirectory != "" {
		args = append(args, "--module-sources-git-directory="+cmd.moduleSourcesGitDirectory)
	}

	if cmd.skipVersionValidation {
		args = append(args, "--skip-version-validation")
	}

	if cmd.outputConstructorFile != "" {
		args = append(args, "--output-constructor-file="+cmd.outputConstructorFile)
	}

	println(" >>> Executing command: modulectl", strings.Join(args, " "))

	command = exec.Command("modulectl", args...)
	cmdOut, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("create command failed with output: %s and error: %w", cmdOut, err)
	}
	return nil
}
