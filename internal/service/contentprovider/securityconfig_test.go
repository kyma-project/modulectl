package contentprovider_test

import (
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	commonerrors "github.com/kyma-project/modulectl/internal/common/errors"
	"github.com/kyma-project/modulectl/internal/common/types"
	"github.com/kyma-project/modulectl/internal/service/contentprovider"
)

func Test_SecurityConfig_NewSecurityConfig_ReturnsError_WhenYamlConverterIsNil(t *testing.T) {
	_, err := contentprovider.NewSecurityConfig(nil)

	require.ErrorIs(t, err, commonerrors.ErrInvalidArg)
	assert.Contains(t, err.Error(), "yamlConverter")
}

func Test_SecurityConfig_GetDefaultContent_ReturnsError_WhenArgsIsNil(t *testing.T) {
	svc, _ := contentprovider.NewSecurityConfig(&objectToYAMLConverterStub{})

	result, err := svc.GetDefaultContent(nil)

	require.ErrorIs(t, err, contentprovider.ErrInvalidArg)
	assert.Empty(t, result)
	assert.Contains(t, err.Error(), "args")
}

func Test_SecurityConfig_GetDefaultContent_ReturnsError_WhenModuleNameArgMissing(t *testing.T) {
	svc, _ := contentprovider.NewSecurityConfig(&objectToYAMLConverterStub{})

	result, err := svc.GetDefaultContent(types.KeyValueArgs{})

	require.ErrorIs(t, err, contentprovider.ErrMissingArg)
	assert.Empty(t, result)
	assert.Contains(t, err.Error(), "moduleName")
}

func Test_SecurityConfig_GetDefaultContent_ReturnsError_WhenModuleNameArgIsEmpty(t *testing.T) {
	svc, _ := contentprovider.NewSecurityConfig(&objectToYAMLConverterStub{})

	result, err := svc.GetDefaultContent(types.KeyValueArgs{contentprovider.ArgModuleName: ""})

	require.ErrorIs(t, err, contentprovider.ErrInvalidArg)
	assert.Empty(t, result)
	assert.Contains(t, err.Error(), "moduleName")
}

func Test_SecurityConfig_GetDefaultContent_ReturnsConvertedContent(t *testing.T) {
	svc, _ := contentprovider.NewSecurityConfig(&objectToYAMLConverterStub{})

	result, err := svc.GetDefaultContent(types.KeyValueArgs{contentprovider.ArgModuleName: "module-name"})

	require.NoError(t, err)
	assert.Equal(t, convertedContent, result)
}

func Test_SecurityScanConfig_ValidateProtecodeImageTags_IgnoresLatestTag(t *testing.T) {
	config := contentprovider.SecurityScanConfig{
		Protecode: []string{
			"europe-docker.pkg.dev/kyma-project/dev/test-image:1.2.3",
			"europe-docker.pkg.dev/kyma-project/dev/test-image:latest",
		},
	}

	err := config.ValidateProtecodeImageTags()

	require.NoError(t, err)
	assert.Len(t, config.Protecode, 1)
	assert.Equal(t, "europe-docker.pkg.dev/kyma-project/dev/test-image:1.2.3", config.Protecode[0])
}

func Test_SecurityScanConfig_ValidateProtecodeImageTags_ReturnsError_WhenNonSemVerTag(t *testing.T) {
	config := contentprovider.SecurityScanConfig{
		Protecode: []string{
			"europe-docker.pkg.dev/kyma-project/dev/test-image:1.2.3",
			"europe-docker.pkg.dev/kyma-project/dev/test-image:non-semver",
		},
	}

	err := config.ValidateProtecodeImageTags()

	require.ErrorIs(t, err, semver.ErrInvalidSemVer)
}

func Test_SecurityScanConfig_ValidateProtecodeImageTags_ReturnsNoError_WhenValidTagsProvided(t *testing.T) {
	config := contentprovider.SecurityScanConfig{
		Protecode: []string{
			"europe-docker.pkg.dev/kyma-project/dev/test-image:1.2.3",
			"europe-docker.pkg.dev/kyma-project/dev/another-image:4.5.6",
		},
	}

	err := config.ValidateProtecodeImageTags()

	require.NoError(t, err)
	assert.Len(t, config.Protecode, 2)
	assert.Equal(t, "europe-docker.pkg.dev/kyma-project/dev/test-image:1.2.3", config.Protecode[0])
	assert.Equal(t, "europe-docker.pkg.dev/kyma-project/dev/another-image:4.5.6", config.Protecode[1])
}

func TestIsWhitelistedNonSemVerTags(t *testing.T) {
	tests := []struct {
		name string
		tag  string
		want bool
	}{
		{
			name: "tag is latest",
			tag:  "latest",
			want: true,
		},
		{
			name: "tag is other text",
			tag:  "rc-1",
			want: false,
		},
		{
			name: "tag is empty",
			tag:  "",
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, contentprovider.IsWhitelistedNonSemVerTags(tt.tag),
				"IsWhitelistedNonSemVerTags(%v)", tt.tag)
		})
	}
}

// Test Stubs

type objectToYAMLConverterStub struct{}

const convertedContent = "content"

func (o *objectToYAMLConverterStub) ConvertToYaml(_ interface{}) string {
	return convertedContent
}
