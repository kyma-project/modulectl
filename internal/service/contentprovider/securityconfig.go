package contentprovider

import (
	"fmt"
	"regexp"

	"github.com/Masterminds/semver/v3"

	commonerrors "github.com/kyma-project/modulectl/internal/common/errors"
	"github.com/kyma-project/modulectl/internal/common/types"
	"github.com/kyma-project/modulectl/internal/service/image"
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
	ModuleName string        `comment:"string, name of your module"                                                       json:"module-name" yaml:"module-name"`
	BDBA       []string      `comment:"list, includes the images which must be scanned by the Black Duck Binary Analysis" json:"bdba"        yaml:"bdba"`
	Mend       MendSecConfig `comment:"Mend security scanner specific configuration"                                      json:"mend"        yaml:"mend"`
	DevBranch  string        `comment:"string, name of the development branch"                                            json:"dev-branch"  yaml:"dev-branch"`
	RcTag      string        `comment:"string, release candidate tag"                                                     json:"rc-tag"      yaml:"rc-tag"`
}

func (s *SecurityScanConfig) ValidateBDBAImageTags(moduleVersion string) error {
	foundCorrectManagerVersion := false
	filteredImages := make([]string, 0, len(s.BDBA))
	for _, img := range s.BDBA {
		imageInfo, err := image.ParseImageInfo(img)
		if err != nil {
			return fmt.Errorf("failed to parse image reference: %w", err)
		}
		_, err = semver.NewVersion(imageInfo.Tag)
		if err != nil {
			return fmt.Errorf("failed to parse image tag [%s] as semantic version: %w", imageInfo.Tag, err)
		}

		filteredImages = append(filteredImages, img)

		if !foundCorrectManagerVersion {
			foundCorrectManagerVersion = isCorrectManagerVersion(img, moduleVersion)
		}
	}

	if !foundCorrectManagerVersion {
		return fmt.Errorf("no image with the correct manager version found in BDBA images 'europe-docker.pkg.dev/kyma-project/prod/<image-name>:%s', %w", moduleVersion, commonerrors.ErrInvalidArg)
	}

	s.BDBA = filteredImages
	return nil
}

type MendSecConfig struct {
	Language    string   `comment:"string, indicating the programming language the scanner has to analyze" json:"language"    yaml:"language"`
	SubProjects string   `comment:"string, specifying any subprojects"                                     json:"subprojects" yaml:"subprojects"`
	Exclude     []string `comment:"list, directories within the repository which should not be scanned"    json:"exclude"     yaml:"exclude"`
}

// revert this again with https://github.com/kyma-project/modulectl/issues/269
// isCorrectManagerVersion checks if the image matches the expected registry and version for the manager
// the exact image name is unknown.
func isCorrectManagerVersion(image, moduleVersion string) bool {
	regex := fmt.Sprintf(`^europe-docker\.pkg\.dev/kyma-project/prod/.*:%s$`, moduleVersion)
	matched, err := regexp.MatchString(regex, image)
	if err != nil {
		return false
	}
	return matched
}
