package create_test

import (
	"strconv"
	"testing"

	createcmd "github.com/kyma-project/modulectl/cmd/modulectl/create"
)

func Test_ScaffoldFlagsDefaults(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{
			name:     createcmd.ModuleConfigFileFlagName,
			value:    createcmd.ModuleConfigFileFlagDefault,
			expected: "module-config.yaml",
		},
		{name: createcmd.CredentialsFlagName, value: createcmd.CredentialsFlagDefault, expected: ""},
		{name: createcmd.GitRemoteFlagName, value: createcmd.GitRemoteFlagDefault, expected: "origin"},
		{name: createcmd.InsecureFlagName, value: strconv.FormatBool(createcmd.InsecureFlagDefault), expected: "false"},
		{name: createcmd.TemplateOutputFlagName, value: createcmd.TemplateOutputFlagDefault, expected: "template.yaml"},
		{name: createcmd.RegistryURLFlagName, value: createcmd.RegistryURLFlagDefault, expected: ""},
		{name: createcmd.RegistryCredSelectorFlagName, value: createcmd.RegistryCredSelectorFlagDefault, expected: ""},
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
