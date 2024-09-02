package contentprovider

import (
	"fmt"

	commonerrors "github.com/kyma-project/modulectl/internal/common/errors"
	"github.com/kyma-project/modulectl/internal/common/types"
)

type SecurityConfig struct {
	yamlConverter ObjectToYAMLConverter
}

func NewSecurityConfig(yamlConverter ObjectToYAMLConverter) (*SecurityConfig, error) {
	if yamlConverter == nil {
		return nil, fmt.Errorf("%w: yamlConverter must not be nil", commonerrors.ErrInvalidArg)
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
		return fmt.Errorf("%w: args must not be nil", ErrInvalidArg)
	}

	value, ok := args[ArgModuleName]
	if !ok {
		return fmt.Errorf("%w: %s", ErrMissingArg, ArgModuleName)
	}

	if value == "" {
		return fmt.Errorf("%w: %s must not be empty", ErrInvalidArg, ArgModuleName)
	}

	return nil
}

func (s *SecurityConfig) getSecurityConfig(moduleName string) securityScanConfig {
	return securityScanConfig{
		ModuleName: moduleName,
		Protecode: []string{
			"europe-docker.pkg.dev/kyma-project/prod/myimage:1.2.3",
			"europe-docker.pkg.dev/kyma-project/prod/external/ghcr.io/mymodule/anotherimage:4.5.6",
		},
		WhiteSource: whiteSourceSecConfig{
			Exclude: []string{"**/test/**", "**/*_test.go"},
		},
	}
}

type securityScanConfig struct {
	ModuleName  string               `json:"module-name" yaml:"module-name" comment:"string, name of your create"`
	Protecode   []string             `json:"protecode" yaml:"protecode" comment:"list, includes the images which must be scanned by the Protecode scanner (aka. Black Duck Binary Analysis)"`
	WhiteSource whiteSourceSecConfig `json:"whitesource" yaml:"whitesource" comment:"whitesource (aka. Mend) security scanner specific configuration"`
	DevBranch   string               `json:"dev-branch" yaml:"dev-branch" comment:"string, name of the development branch"`
	RcTag       string               `json:"rc-tag" yaml:"rc-tag" comment:"string, release candidate tag"`
}

type whiteSourceSecConfig struct {
	Language    string   `json:"language" yaml:"language" comment:"string, indicating the programming language the scanner has to analyze"`
	SubProjects string   `json:"subprojects" yaml:"subprojects" comment:"string, specifying any subprojects"`
	Exclude     []string `json:"exclude" yaml:"exclude" comment:"list, directories within the repository which should not be scanned"`
}
