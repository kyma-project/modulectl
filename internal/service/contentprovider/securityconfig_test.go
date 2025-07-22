package contentprovider_test

import (
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/stretchr/testify/require"

	commonerrors "github.com/kyma-project/modulectl/internal/common/errors"
	"github.com/kyma-project/modulectl/internal/common/types"
	"github.com/kyma-project/modulectl/internal/service/contentprovider"
)

func Test_SecurityConfig_NewSecurityConfig_ReturnsError_WhenYamlConverterIsNil(t *testing.T) {
	_, err := contentprovider.NewSecurityConfig(nil)

	require.ErrorIs(t, err, commonerrors.ErrInvalidArg)
	require.Contains(t, err.Error(), "yamlConverter")
}

func Test_SecurityConfig_GetDefaultContent_ReturnsError_WhenArgsIsNil(t *testing.T) {
	svc, _ := contentprovider.NewSecurityConfig(&objectToYAMLConverterStub{})

	result, err := svc.GetDefaultContent(nil)

	require.ErrorIs(t, err, contentprovider.ErrInvalidArg)
	require.Empty(t, result)
	require.Contains(t, err.Error(), "args")
}

func Test_SecurityConfig_GetDefaultContent_ReturnsError_WhenModuleNameArgMissing(t *testing.T) {
	svc, _ := contentprovider.NewSecurityConfig(&objectToYAMLConverterStub{})

	result, err := svc.GetDefaultContent(types.KeyValueArgs{})

	require.ErrorIs(t, err, contentprovider.ErrMissingArg)
	require.Empty(t, result)
	require.Contains(t, err.Error(), "moduleName")
}

func Test_SecurityConfig_GetDefaultContent_ReturnsError_WhenModuleNameArgIsEmpty(t *testing.T) {
	svc, _ := contentprovider.NewSecurityConfig(&objectToYAMLConverterStub{})

	result, err := svc.GetDefaultContent(types.KeyValueArgs{contentprovider.ArgModuleName: ""})

	require.ErrorIs(t, err, contentprovider.ErrInvalidArg)
	require.Empty(t, result)
	require.Contains(t, err.Error(), "moduleName")
}

func Test_SecurityConfig_GetDefaultContent_ReturnsConvertedContent(t *testing.T) {
	svc, _ := contentprovider.NewSecurityConfig(&objectToYAMLConverterStub{})

	result, err := svc.GetDefaultContent(types.KeyValueArgs{contentprovider.ArgModuleName: "module-name"})

	require.NoError(t, err)
	require.Equal(t, convertedContent, result)
}

func Test_SecurityScanConfig_ValidateBDBAImageTags_ReturnsError_WhenImageNameAndTagInvalid(t *testing.T) {
	config := contentprovider.SecurityScanConfig{
		BDBA: []string{
			"invalid-image-format",
		},
	}

	err := config.ValidateBDBAImageTags("1.2.3")

	require.Error(t, err)
	require.Contains(t, err.Error(), "no tag or digest found")
}

func Test_SecurityScanConfig_ValidateBDBAImageTags_ReturnsError_WhenLatestTag(t *testing.T) {
	config := contentprovider.SecurityScanConfig{
		BDBA: []string{
			"europe-docker.pkg.dev/kyma-project/prod/test-image:latest",
		},
	}
	err := config.ValidateBDBAImageTags("1.2.3")

	require.ErrorIs(t, err, semver.ErrInvalidSemVer)
}

func Test_SecurityScanConfig_ValidateBDBAImageTags_ReturnsError_WhenNonSemVerTag(t *testing.T) {
	config := contentprovider.SecurityScanConfig{
		BDBA: []string{
			"europe-docker.pkg.dev/kyma-project/prod/test-image:1.2.3",
			"europe-docker.pkg.dev/kyma-project/prod/test-image:non-semver",
		},
	}

	err := config.ValidateBDBAImageTags("1.2.3")

	require.ErrorIs(t, err, semver.ErrInvalidSemVer)
}

func Test_SecurityScanConfig_ValidateBDBAImageTags_ReturnsError_WhenInvalidManagerImageProvided(t *testing.T) {
	config := contentprovider.SecurityScanConfig{
		BDBA: []string{
			"europe-docker.pkg.dev/kyma-project/prod/test-image:1.2.4",
			"europe-docker.pkg.dev/kyma-project/prod/another-image:4.5.6",
		},
	}

	err := config.ValidateBDBAImageTags("1.2.3")

	require.Error(t, err)
	require.Contains(t, err.Error(), "no image with the correct manager version found in BDBA images 'europe-docker.pkg.dev/kyma-project/prod/<image-name>:1.2.3'")
}

func Test_SecurityScanConfig_ValidateBDBAImageTags_ReturnsNoError_WhenValidTagsProvided(t *testing.T) {
	config := contentprovider.SecurityScanConfig{
		BDBA: []string{
			"europe-docker.pkg.dev/kyma-project/prod/test-image:1.2.3",
			"europe-docker.pkg.dev/kyma-project/prod/another-image:4.5.6",
		},
	}

	err := config.ValidateBDBAImageTags("1.2.3")

	require.NoError(t, err)
	require.Len(t, config.BDBA, 2)
	require.Equal(t, "europe-docker.pkg.dev/kyma-project/prod/test-image:1.2.3", config.BDBA[0])
	require.Equal(t, "europe-docker.pkg.dev/kyma-project/prod/another-image:4.5.6", config.BDBA[1])
}

// Test Stubs

type objectToYAMLConverterStub struct{}

const convertedContent = "content"

func (o *objectToYAMLConverterStub) ConvertToYaml(_ interface{}) string {
	return convertedContent
}
