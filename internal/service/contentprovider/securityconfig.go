package contentprovider

import (
	"fmt"
	"regexp"

	"github.com/Masterminds/semver/v3"

	commonerrors "github.com/kyma-project/modulectl/internal/common/errors"
	"github.com/kyma-project/modulectl/internal/common/types"
	"github.com/kyma-project/modulectl/internal/utils"
)

type SecurityConfig struct {
	yamlConverter ObjectToYAMLConverter
}

func NewSecurityConfig(yamlConverter ObjectToYAMLConverter) (*SecurityConfig, error) {
	if yamlConverter == nil {
		return nil, fmt.Errorf("yamlConverter must not be nil: %w", commonerrors.ErrInvalidArg)
	}

	return &SecurityConfig{
		yamlConverter: yamlConverter,
	}, nil
}

func (s *SecurityConfig) GetDefaultContent(args types.KeyValueArgs) (string, error) {
	if err := s.validateArgs(args); err != nil {
		return "", err
	}

	return s.yamlConverter.ConvertToYaml(s.getSecurityConfig(args[ArgModuleName])), nil
}

func (s *SecurityConfig) validateArgs(args types.KeyValueArgs) error {
	if args == nil {
		return fmt.Errorf("args must not be nil: %w", ErrInvalidArg)
	}

	value, ok := args[ArgModuleName]
	if !ok {
		return fmt.Errorf("%s: %w", ArgModuleName, ErrMissingArg)
	}

	if value == "" {
		return fmt.Errorf("%s must not be empty: %w", ArgModuleName, ErrInvalidArg)
	}

	return nil
}

func (s *SecurityConfig) getSecurityConfig(moduleName string) SecurityScanConfig {
	return SecurityScanConfig{
		ModuleName: moduleName,
		BDBA: []string{
			"europe-docker.pkg.dev/kyma-project/prod/myimage:1.2.3",
			"europe-docker.pkg.dev/kyma-project/prod/external/ghcr.io/mymodule/anotherimage:4.5.6",
		},
		Mend: MendSecConfig{
			Exclude: []string{"**/test/**", "**/*_test.go"},
		},
	}
}

type SecurityScanConfig struct {
	ModuleName string        `json:"module-name" yaml:"module-name" comment:"string, name of your module"`
	BDBA       []string      `json:"bdba" yaml:"bdba" comment:"list, includes the images which must be scanned by the Black Duck Binary Analysis"`
	Mend       MendSecConfig `json:"mend" yaml:"mend" comment:"Mend security scanner specific configuration"`
	DevBranch  string        `json:"dev-branch" yaml:"dev-branch" comment:"string, name of the development branch"`
	RcTag      string        `json:"rc-tag" yaml:"rc-tag" comment:"string, release candidate tag"`
}

func (s *SecurityScanConfig) ValidateBDBAImageTags(moduleVersion string) error {
	foundCorrectManagerVersion := false
	filteredImages := make([]string, 0, len(s.BDBA))
	for _, image := range s.BDBA {
		_, tag, err := utils.GetImageNameAndTag(image)
		if err != nil {
			return fmt.Errorf("failed to get image name and tag: %w", err)
		}
		_, err = semver.NewVersion(tag)
		if err != nil {
			return fmt.Errorf("failed to parse image tag [%s] as semantic version: %w", tag, err)
		}
		filteredImages = append(filteredImages, image)

		if !foundCorrectManagerVersion {
			foundCorrectManagerVersion = isCorrectManagerVersion(image, moduleVersion)
		}
	}

	if !foundCorrectManagerVersion {
		return fmt.Errorf("no image with the correct manager version found in BDBA images 'europe-docker.pkg.dev/kyma-project/prod/<image-name>:%s', %w", moduleVersion, commonerrors.ErrInvalidArg)
	}

	s.BDBA = filteredImages
	return nil
}

type MendSecConfig struct {
	Language    string   `json:"language" yaml:"language" comment:"string, indicating the programming language the scanner has to analyze"`
	SubProjects string   `json:"subprojects" yaml:"subprojects" comment:"string, specifying any subprojects"`
	Exclude     []string `json:"exclude" yaml:"exclude" comment:"list, directories within the repository which should not be scanned"`
}

// TODO: revert this again with https://github.com/kyma-project/modulectl/issues/269
func isCorrectManagerVersion(image, moduleVersion string) bool {
	regex := fmt.Sprintf(`^europe-docker\.pkg\.dev/kyma-project/prod/.*:%s$`, moduleVersion)
	matched, err := regexp.MatchString(regex, image)
	if err != nil {
		return false
	}
	return matched
}
