package scaffold_test

import (
	"strconv"
	"testing"

	scaffoldcmd "github.com/kyma-project/modulectl/cmd/modulectl/scaffold"
)

func Test_ScaffoldFlagsDefaults(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{name: scaffoldcmd.DirectoryFlagName, value: scaffoldcmd.DirectoryFlagDefault, expected: "./"},
		{
			name:     scaffoldcmd.ModuleConfigFileFlagName,
			value:    scaffoldcmd.ModuleConfigFileFlagDefault,
			expected: "scaffold-module-config.yaml",
		},
		{
			name:     scaffoldcmd.ModuleConfigFileOverwriteFlagName,
			value:    strconv.FormatBool(scaffoldcmd.ModuleConfigFileOverwriteFlagDefault),
			expected: "false",
		},
		{name: scaffoldcmd.ManifestFileFlagName, value: scaffoldcmd.ManifestFileFlagDefault, expected: "manifest.yaml"},
		{name: scaffoldcmd.DefaultCRFlagName, value: scaffoldcmd.DefaultCRFlagDefault, expected: ""},
		{
			name:     scaffoldcmd.DefaultCRFlagName,
			value:    scaffoldcmd.DefaultCRFlagNoOptDefault,
			expected: "default-cr.yaml",
		},
		{name: scaffoldcmd.SecurityConfigFileFlagName, value: scaffoldcmd.SecurityConfigFileFlagDefault, expected: ""},
		{
			name:     scaffoldcmd.SecurityConfigFileFlagName,
			value:    scaffoldcmd.SecurityConfigFileFlagNoOptDefault,
			expected: "sec-scanners-config.yaml",
		},
		{
			name:     scaffoldcmd.ModuleNameFlagName,
			value:    scaffoldcmd.ModuleNameFlagDefault,
			expected: "kyma-project.io/module/mymodule",
		},
		{name: scaffoldcmd.ModuleVersionFlagName, value: scaffoldcmd.ModuleVersionFlagDefault, expected: "0.0.1"},
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
