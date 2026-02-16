package contentprovider_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	commonerrors "github.com/kyma-project/modulectl/internal/common/errors"
	"github.com/kyma-project/modulectl/internal/common/types"
	"github.com/kyma-project/modulectl/internal/service/contentprovider"
)

func Test_ModuleConfig_NewModuleConfig_ReturnsError_WhenYamlConverterIsNil(t *testing.T) {
	_, err := contentprovider.NewModuleConfigProvider(nil)

	require.ErrorIs(t, err, commonerrors.ErrInvalidArg)
	assert.Contains(t, err.Error(), "yamlConverter")
}

func Test_ModuleConfig_GetDefaultContent_ReturnsError_WhenArgsIsNil(t *testing.T) {
	svc, _ := contentprovider.NewModuleConfigProvider(&mcObjectToYAMLConverterStub{})

	result, err := svc.GetDefaultContent(nil)

	require.ErrorIs(t, err, contentprovider.ErrInvalidArg)
	assert.Empty(t, result)
	assert.Contains(t, err.Error(), "args")
}

func Test_ModuleConfig_GetDefaultContent_ReturnsError_WhenRequiredArgsMissing(t *testing.T) {
	t.Parallel()
	tests := []struct {
		argName string
		args    types.KeyValueArgs
	}{
		{
			argName: contentprovider.ArgModuleName,
			args: types.KeyValueArgs{
				contentprovider.ArgModuleVersion: "0.0.1",
			},
		},
		{
			argName: contentprovider.ArgModuleVersion,
			args: types.KeyValueArgs{
				contentprovider.ArgModuleName: "module-name",
			},
		},
	}

	svc, _ := contentprovider.NewModuleConfigProvider(&mcObjectToYAMLConverterStub{})

	for _, testcase := range tests {
		testName := "TestArgumentRequired_" + testcase.argName
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			result, err := svc.GetDefaultContent(testcase.args)

			require.ErrorIs(t, err, contentprovider.ErrMissingArg)
			assert.Empty(t, result)
			assert.Contains(t, err.Error(), testcase.argName)
		})
	}
}

func Test_ModuleConfig_GetDefaultContent_ReturnsError_WhenRequiredArgIsEmpty(t *testing.T) {
	t.Parallel()
	tests := []struct {
		argName string
		args    types.KeyValueArgs
	}{
		{
			argName: contentprovider.ArgModuleName,
			args: types.KeyValueArgs{
				contentprovider.ArgModuleName:    "",
				contentprovider.ArgModuleVersion: "0.0.1",
			},
		},
		{
			argName: contentprovider.ArgModuleVersion,
			args: types.KeyValueArgs{
				contentprovider.ArgModuleName:    "module-name",
				contentprovider.ArgModuleVersion: "",
			},
		},
	}

	svc, _ := contentprovider.NewModuleConfigProvider(&mcObjectToYAMLConverterStub{})

	for _, testcase := range tests {
		testName := "TestArgumentRequired_" + testcase.argName
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			result, err := svc.GetDefaultContent(testcase.args)

			require.ErrorIs(t, err, contentprovider.ErrInvalidArg)
			assert.Empty(t, result)
			assert.Contains(t, err.Error(), testcase.argName)
		})
	}
}

func Test_ModuleConfig_GetDefaultContent_ReturnsConvertedContent(t *testing.T) {
	svc, _ := contentprovider.NewModuleConfigProvider(&mcObjectToYAMLConverterStub{})

	result, err := svc.GetDefaultContent(types.KeyValueArgs{
		contentprovider.ArgModuleName:    "module-name",
		contentprovider.ArgModuleVersion: "0.0.1",
	})

	require.NoError(t, err)
	assert.Equal(t, mcConvertedContent, result)
}

func Test_ModuleConfig_Unmarshal_Icons_Success(t *testing.T) {
	moduleConfigData := `
icons:
  - name: icon1
    link: https://example.com/icon1
  - name: icon2
    link: https://example.com/icon2
`

	moduleConfig := &contentprovider.ModuleConfig{}
	err := yaml.Unmarshal([]byte(moduleConfigData), moduleConfig)

	require.NoError(t, err)
	assert.Len(t, moduleConfig.Icons, 2)
	assert.Equal(t, "https://example.com/icon1", moduleConfig.Icons["icon1"])
	assert.Equal(t, "https://example.com/icon2", moduleConfig.Icons["icon2"])
}

func Test_ModuleConfig_Unmarshal_Icons_Success_Ignoring_Unknown_Fields(t *testing.T) {
	moduleConfigData := `
icons:
  - name: icon
    link: https://example.com/icon
    unknown: something
`

	moduleConfig := &contentprovider.ModuleConfig{}
	err := yaml.Unmarshal([]byte(moduleConfigData), moduleConfig)

	require.NoError(t, err)
	assert.Len(t, moduleConfig.Icons, 1)
	assert.Equal(t, "https://example.com/icon", moduleConfig.Icons["icon"])
}

func Test_ModuleConfig_Unmarshal_Icons_FailOnDuplicateNames(t *testing.T) {
	moduleConfigData := `
icons:
  - name: icon1
    link: https://example.com/icon1
  - name: icon1
    link: https://example.com/icon2
`

	moduleConfig := &contentprovider.ModuleConfig{}
	err := yaml.Unmarshal([]byte(moduleConfigData), moduleConfig)

	require.Error(t, err)
	assert.Equal(t, "failed to unmarshal Icons: map contains duplicate entries", err.Error())
}

func Test_ModuleConfig_Marshal_Icons_Success(t *testing.T) {
	// parse the expected config
	expectedModuleConfigData := `
icons:
  - name: icon1
    link: https://example.com/icon1
  - name: icon2
    link: https://example.com/icon2
`
	expectedModuleConfig := &contentprovider.ModuleConfig{}
	err := yaml.Unmarshal([]byte(expectedModuleConfigData), expectedModuleConfig)
	require.NoError(t, err)

	moduleConfig := &contentprovider.ModuleConfig{
		Icons: contentprovider.Icons{
			"icon1": "https://example.com/icon1",
			"icon2": "https://example.com/icon2",
		},
	}
	marshalledModuleConfigData, err := yaml.Marshal(moduleConfig)
	require.NoError(t, err)

	marshalledModuleConfig := &contentprovider.ModuleConfig{}
	err = yaml.Unmarshal(marshalledModuleConfigData, marshalledModuleConfig)

	require.NoError(t, err)
	assert.Equal(t, expectedModuleConfig.Icons, marshalledModuleConfig.Icons)
}

func Test_ModuleConfig_Unmarshall_Resources_Success(t *testing.T) {
	moduleConfigData := `
resources:
  - name: resource1
    link: https://example.com/resource1
  - name: resource2
    link: https://example.com/resource2
`

	moduleConfig := &contentprovider.ModuleConfig{}
	err := yaml.Unmarshal([]byte(moduleConfigData), moduleConfig)

	require.NoError(t, err)
	assert.Len(t, moduleConfig.Resources, 2)
	assert.Equal(t, "https://example.com/resource1", moduleConfig.Resources["resource1"])
	assert.Equal(t, "https://example.com/resource2", moduleConfig.Resources["resource2"])
}

func Test_ModuleConfig_Unmarshall_Resources_Success_Ignoring_Unknown_Fields(t *testing.T) {
	moduleConfigData := `
resources:
  - name: resource1
    link: https://example.com/resource1
    unknown: something
`

	moduleConfig := &contentprovider.ModuleConfig{}
	err := yaml.Unmarshal([]byte(moduleConfigData), moduleConfig)

	require.NoError(t, err)
	assert.Len(t, moduleConfig.Resources, 1)
	assert.Equal(t, "https://example.com/resource1", moduleConfig.Resources["resource1"])
}

func Test_ModuleConfig_Unmarshall_Resources_FailOnDuplicateNames(t *testing.T) {
	moduleConfigData := `
resources:
  - name: resource1
    link: https://example.com/resource1
  - name: resource1
    link: https://example.com/resource1
`

	moduleConfig := &contentprovider.ModuleConfig{}
	err := yaml.Unmarshal([]byte(moduleConfigData), moduleConfig)

	require.Error(t, err)
	assert.Equal(t, "failed to unmarshal Resources: map contains duplicate entries", err.Error())
}

func Test_ModuleConfig_Marshall_Resources_Success(t *testing.T) {
	// parse the expected config
	expectedModuleConfigData := `
resources:
  - name: resource1
    link: https://example.com/resource1
  - name: resource2
    link: https://example.com/resource2
`
	expectedModuleConfig := &contentprovider.ModuleConfig{}
	err := yaml.Unmarshal([]byte(expectedModuleConfigData), expectedModuleConfig)
	require.NoError(t, err)

	// round trip a module config (marshal and unmarshal)
	moduleConfig := &contentprovider.ModuleConfig{
		Resources: contentprovider.Resources{
			"resource1": "https://example.com/resource1",
			"resource2": "https://example.com/resource2",
		},
	}
	marshalledModuleConfigData, err := yaml.Marshal(moduleConfig)
	require.NoError(t, err)

	roudTrippedModuleConfig := &contentprovider.ModuleConfig{}
	err = yaml.Unmarshal(marshalledModuleConfigData, roudTrippedModuleConfig)

	require.NoError(t, err)
	assert.Equal(t, expectedModuleConfig.Resources, roudTrippedModuleConfig.Resources)
}

// Test Stubs

type mcObjectToYAMLConverterStub struct{}

const mcConvertedContent = "content"

func (o *mcObjectToYAMLConverterStub) ConvertToYaml(_ any) string {
	return mcConvertedContent
}
