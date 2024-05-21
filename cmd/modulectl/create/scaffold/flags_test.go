package scaffold_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/kyma-project/modulectl/cmd/modulectl/create/scaffold"
)

func Test_ScaffoldFlagsDefaults(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{name: scaffold.DirectoryFlagName, value: scaffold.DirectoryFlagDefault, expected: "./"},
		{name: scaffold.ModuleConfigFileFlagName, value: scaffold.ModuleConfigFileFlagDefault, expected: "scaffold-module-config.yaml"},
		{name: scaffold.ModuleConfigFileOverwriteFlagName, value: strconv.FormatBool(scaffold.ModuleConfigFileOverwriteFlagDefault), expected: "false"},
		{name: scaffold.ManifestFileFlagName, value: scaffold.ManifestFileFlagDefault, expected: "manifest.yaml"},
		{name: scaffold.DefaultCRFlagName, value: scaffold.DefaultCRFlagDefault, expected: ""},
		{name: scaffold.DefaultCRFlagName, value: scaffold.DefaultCRFlagNoOptDefault, expected: "default-cr.yaml"},
		{name: scaffold.SecurityConfigFileFlagName, value: scaffold.SecurityConfigFileFlagDefault, expected: ""},
		{name: scaffold.SecurityConfigFileFlagName, value: scaffold.SecurityConfigFileFlagNoOptDefault, expected: "sec-scanners-config.yaml"},
		{name: scaffold.ModuleNameFlagName, value: scaffold.ModuleNameFlagDefault, expected: "kyma-project.io/module/mymodule"},
		{name: scaffold.ModuleVersionFlagName, value: scaffold.ModuleVersionFlagDefault, expected: "0.0.1"},
		{name: scaffold.ModuleChannelFlagName, value: scaffold.ModuleChannelFlagDefault, expected: "regular"},
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
