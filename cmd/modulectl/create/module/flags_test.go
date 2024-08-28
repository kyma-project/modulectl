package module_test

import (
	"strconv"
	"testing"

	modulecmd "github.com/kyma-project/modulectl/cmd/modulectl/create/module"
)

func Test_ScaffoldFlagsDefaults(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{name: modulecmd.ModuleConfigFileFlagName, value: modulecmd.ModuleConfigFileFlagDefault, expected: "module-config.yaml"},
		{name: modulecmd.CredentialsFlagName, value: modulecmd.CredentialsFlagDefault, expected: ""},
		{name: modulecmd.GitRemoteFlagName, value: modulecmd.GitRemoteFlagDefault, expected: "origin"},
		{name: modulecmd.InsecureFlagName, value: strconv.FormatBool(modulecmd.InsecureFlagDefault), expected: "false"},
		{name: modulecmd.TemplateOutputFlagName, value: modulecmd.TemplateOutputFlagDefault, expected: "template.yaml"},
		{name: modulecmd.RegistryURLFlagName, value: modulecmd.RegistryURLFlagDefault, expected: ""},
		{name: modulecmd.RegistryCredSelectorFlagName, value: modulecmd.RegistryCredSelectorFlagDefault, expected: ""},
		{name: modulecmd.SecScannersConfigFlagName, value: modulecmd.SecScannersConfigFlagDefault, expected: "sec-scanners-config.yaml"},
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
