package contentprovider_test

import (
	"fmt"
	"testing"

	"github.com/kyma-project/modulectl/internal/scaffold/common/types"
	"github.com/kyma-project/modulectl/internal/scaffold/contentprovider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_ModuleConfigContentProvider_GetDefaultContent_ReturnsError_WhenArgsIsNil(t *testing.T) {
	svc := contentprovider.NewModuleConfigContentProvider(&mcObjectToYAMLConverterStub{})

	result, err := svc.GetDefaultContent(nil)

	require.ErrorIs(t, err, contentprovider.ErrInvalidArg)
	assert.Empty(t, result)
	assert.Contains(t, err.Error(), "args")
}

func Test_ModuleConfigContentProvider_ReturnsError_WhenRequiredArgsMissing(t *testing.T) {
	t.Parallel()
	tests := []struct {
		argName string
		args    types.KeyValueArgs
	}{
		{
			argName: contentprovider.ArgModuleName,
			args: types.KeyValueArgs{
				contentprovider.ArgModuleChannel: "experimental",
				contentprovider.ArgModuleVersion: "0.0.1",
			},
		},
		{
			argName: contentprovider.ArgModuleVersion,
			args: types.KeyValueArgs{
				contentprovider.ArgModuleName:    "module-name",
				contentprovider.ArgModuleChannel: "experimental",
			},
		},
		{
			argName: contentprovider.ArgModuleChannel,
			args: types.KeyValueArgs{
				contentprovider.ArgModuleName:    "module-name",
				contentprovider.ArgModuleVersion: "0.0.1",
			},
		},
	}

	svc := contentprovider.NewModuleConfigContentProvider(&mcObjectToYAMLConverterStub{})

	for _, testcase := range tests {
		testcase := testcase
		testName := fmt.Sprintf("TestArgumentRequired_%s", testcase.argName)
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			result, err := svc.GetDefaultContent(testcase.args)

			require.ErrorIs(t, err, contentprovider.ErrMissingArg)
			assert.Empty(t, result)
			assert.Contains(t, err.Error(), testcase.argName)
		})
	}
}

func Test_ModuleConfigContentProvider_ReturnsError_WhenRequiredArgIsEmpty(t *testing.T) {
	t.Parallel()
	tests := []struct {
		argName string
		args    types.KeyValueArgs
	}{
		{
			argName: contentprovider.ArgModuleName,
			args: types.KeyValueArgs{
				contentprovider.ArgModuleName:    "",
				contentprovider.ArgModuleChannel: "experimental",
				contentprovider.ArgModuleVersion: "0.0.1",
			},
		},
		{
			argName: contentprovider.ArgModuleVersion,
			args: types.KeyValueArgs{
				contentprovider.ArgModuleName:    "module-name",
				contentprovider.ArgModuleChannel: "experimental",
				contentprovider.ArgModuleVersion: "",
			},
		},
		{
			argName: contentprovider.ArgModuleChannel,
			args: types.KeyValueArgs{
				contentprovider.ArgModuleName:    "module-name",
				contentprovider.ArgModuleChannel: "",
				contentprovider.ArgModuleVersion: "0.0.1",
			},
		},
	}

	svc := contentprovider.NewModuleConfigContentProvider(&mcObjectToYAMLConverterStub{})

	for _, testcase := range tests {
		testcase := testcase
		testName := fmt.Sprintf("TestArgumentRequired_%s", testcase.argName)
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			result, err := svc.GetDefaultContent(testcase.args)

			require.ErrorIs(t, err, contentprovider.ErrInvalidArg)
			assert.Empty(t, result)
			assert.Contains(t, err.Error(), testcase.argName)
		})
	}
}

func Test_ModuleConfigContentProvider_ReturnsConvertedContent(t *testing.T) {
	svc := contentprovider.NewModuleConfigContentProvider(&mcObjectToYAMLConverterStub{})

	result, err := svc.GetDefaultContent(types.KeyValueArgs{
		contentprovider.ArgModuleName:    "module-name",
		contentprovider.ArgModuleChannel: "regular",
		contentprovider.ArgModuleVersion: "0.0.1",
	})

	require.NoError(t, err)
	assert.Equal(t, mcConvertedContent, result)
}

// ***************
// Test Stubs
// ***************

type mcObjectToYAMLConverterStub struct{}

const mcConvertedContent = "content"

func (o *mcObjectToYAMLConverterStub) ConvertToYaml(obj interface{}) string {
	return mcConvertedContent
}