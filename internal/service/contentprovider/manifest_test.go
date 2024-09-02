package contentprovider_test

import (
	"testing"

	"github.com/kyma-project/modulectl/internal/common/types"
	"github.com/kyma-project/modulectl/internal/service/contentprovider"
)

func Test_Manifest_GetDefaultContent_ReturnsExpectedValue(t *testing.T) {
	manifestContentProvider := contentprovider.NewManifest()

	expectedDefault := `# This file holds the Manifest of your module, encompassing all resources installed in the cluster once the module is activated.
# It should include the Custom Resource Definition for your module's default CustomResource, if it exists.

`

	manifestGeneratedDefaultContentWithNil, _ := manifestContentProvider.GetDefaultContent(nil)
	manifestGeneratedDefaultContentWithEmptyMap, _ := manifestContentProvider.GetDefaultContent(make(types.KeyValueArgs))

	t.Parallel()
	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{
			name:     "Manifest Default Content with Nil",
			value:    manifestGeneratedDefaultContentWithNil,
			expected: expectedDefault,
		}, {
			name:     "Manifest Default Content with Empty Map",
			value:    manifestGeneratedDefaultContentWithEmptyMap,
			expected: expectedDefault,
		},
	}

	for _, testcase := range tests {
		testName := "TestCorrectContentProviderFor_" + testcase.name
		t.Run(testName, func(t *testing.T) {
			t.Parallel()
			if testcase.value != testcase.expected {
				t.Errorf("ContentProvider for '%s' did not return correct default: expected = '%s', but got = '%s'",
					testcase.name, testcase.expected, testcase.value)
			}
		})
	}
}
