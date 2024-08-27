package module_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/kyma-project/modulectl/cmd/modulectl/create/module"
)

func Test_ScaffoldFlagsDefaults(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{name: module.ModuleConfigFileFlagName, value: module.ModuleConfigFileFlagDefault, expected: "module-config.yaml"},
		{name: module.CredentialsFlagName, value: module.CredentialsFlagDefault, expected: ""},
		{name: module.GitRemoteFlagName, value: module.GitRemoteFlagDefault, expected: "origin"},
		{name: module.InsecureFlagName, value: strconv.FormatBool(module.InsecureFlagDefault), expected: "false"},
		{name: module.TemplateOutputFlagName, value: module.TemplateOutputFlagDefault, expected: "template.yaml"},
		{name: module.RegistryURLFlagName, value: module.RegistryURLFlagDefault, expected: ""},
		{name: module.RegistryCredSelectorFlagName, value: module.RegistryCredSelectorFlagDefault, expected: ""},
		{name: module.SecScannersConfigFlagName, value: module.SecScannersConfigFlagDefault, expected: "sec-scanners-config.yaml"},
	}

	for _, testcase := range tests {
		testName := fmt.Sprintf("TestFlagHasCorrectDefault_%s", testcase.name)
		t.Run(testName, func(t *testing.T) {
			if testcase.value != testcase.expected {
				t.Errorf("Flag '%s' has different default: expected = '%s', got = '%s'",
					testcase.name, testcase.expected, testcase.value)
			}
		})
	}
}
