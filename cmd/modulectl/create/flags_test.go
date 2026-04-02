package create_test

import (
	"testing"

	createcmd "github.com/kyma-project/modulectl/cmd/modulectl/create"
)

func Test_CreateFlagsDefaults(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{
			name:     createcmd.ConfigFileFlagName,
			value:    createcmd.ConfigFileFlagDefault,
			expected: "module-config.yaml",
		},
		{
			name:     createcmd.TemplateOutputFlagName,
			value:    createcmd.TemplateOutputFlagDefault,
			expected: "template.yaml",
		},
		{
			name:     createcmd.ModuleSourcesGitDirectoryFlagName,
			value:    createcmd.ModuleSourcesGitDirectoryFlagDefault,
			expected: ".",
		},
		{
			name:     createcmd.OutputConstructorFileFlagName,
			value:    createcmd.OutputConstructorFileFlagDefault,
			expected: "component-constructor.yaml",
		},
	}

	for _, testcase := range tests {
		testName := "TestFlagHasCorrectDefault_" + testcase.name
		t.Run(testName, func(t *testing.T) {
			if testcase.value != testcase.expected {
				t.Errorf("Flag '%s' has different default: expected = '%s', got = '%s'",
					testcase.name, testcase.expected, testcase.value)
			}
		})
	}
}
