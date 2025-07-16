package create_test

import (
	"strconv"
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
		{name: createcmd.CredentialsFlagName, value: createcmd.CredentialsFlagDefault, expected: ""},
		{name: createcmd.InsecureFlagName, value: strconv.FormatBool(createcmd.InsecureFlagDefault), expected: "false"},
		{name: createcmd.TemplateOutputFlagName, value: createcmd.TemplateOutputFlagDefault, expected: "template.yaml"},
		{name: createcmd.RegistryURLFlagName, value: createcmd.RegistryURLFlagDefault, expected: ""},
		{
			name:     createcmd.OverwriteComponentVersionFlagName,
			value:    strconv.FormatBool(createcmd.OverwriteComponentVersionFlagDefault),
			expected: "false",
		},
		{name: createcmd.DryRunFlagName, value: strconv.FormatBool(createcmd.DryRunFlagDefault), expected: "false"},
		{
			name:     createcmd.ModuleSourcesGitDirectoryFlagName,
			value:    createcmd.ModuleSourcesGitDirectoryFlagDefault,
			expected: ".",
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
