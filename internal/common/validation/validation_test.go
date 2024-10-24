package validation_test

import (
	"fmt"
	"testing"

	"github.com/kyma-project/modulectl/internal/common/validation"
	"github.com/kyma-project/modulectl/internal/service/contentprovider"
)

func TestValidateModuleName(t *testing.T) {
	tests := []struct {
		name       string
		moduleName string
		wantErr    bool
	}{
		{
			name:       "valid module name",
			moduleName: "kyma-project.io/module-name",
			wantErr:    false,
		},
		{
			name:       "empty module name",
			moduleName: "",
			wantErr:    true,
		},
		{
			name:       "invalid module name - whitespaces",
			moduleName: " kyma-project.io/module-name ",
			wantErr:    true,
		},
		{
			name:       "invalid module name - no path",
			moduleName: "kyma-project.io",
			wantErr:    true,
		},
		{
			name:       "invalid module name",
			moduleName: "module-name",
			wantErr:    true,
		},
		{
			name:       "invalid module name with path",
			moduleName: "module-name/bar",
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validation.ValidateModuleName(tt.moduleName); (err != nil) != tt.wantErr {
				t.Errorf("ValidateModuleName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateModuleVersion(t *testing.T) {
	tests := []struct {
		name          string
		moduleVersion string
		wantErr       bool
	}{
		{
			name:          "valid module version",
			moduleVersion: "0.0.1",
			wantErr:       false,
		},
		{
			name:          "invalid module version with 'v' prefix",
			moduleVersion: "v0.0.1",
			wantErr:       true,
		},
		{
			name:          "empty module version",
			moduleVersion: "",
			wantErr:       true,
		},
		{
			name:          "invalid module version - no patch",
			moduleVersion: "0.0",
			wantErr:       true,
		},
		{
			name:          "invalid module version - not a semantic version",
			moduleVersion: "main",
			wantErr:       true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validation.ValidateModuleVersion(tt.moduleVersion); (err != nil) != tt.wantErr {
				t.Errorf("ValidateModuleVersion() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateModuleChannel(t *testing.T) {
	tests := []struct {
		name          string
		moduleChannel string
		wantErr       bool
	}{
		{
			name:          "valid channel",
			moduleChannel: "experimental",
			wantErr:       false,
		},
		{
			name:          "empty channel",
			moduleChannel: "",
			wantErr:       true,
		},
		{
			name:          "invalid channel - too short ",
			moduleChannel: "a",
			wantErr:       true,
		},
		{
			name:          "invalid channel - too long",
			moduleChannel: "thisstringvaluehaslengthof33chars",
			wantErr:       true,
		},
		{
			name:          "invalid channel - contains invalid characters",
			moduleChannel: "this value has spaces",
			wantErr:       true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validation.ValidateModuleChannel(tt.moduleChannel); (err != nil) != tt.wantErr {
				t.Errorf("ValidateChannel() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateModuleNamespace(t *testing.T) {
	tests := []struct {
		name            string
		moduleNamespace string
		wantErr         bool
	}{
		{
			name:            "empty module namespace",
			moduleNamespace: "",
			wantErr:         true,
		},
		{
			name:            "valid module namespace",
			moduleNamespace: "kyma-system",
			wantErr:         false,
		},
		{
			name:            "invalid module namespace - whitespaces",
			moduleNamespace: " kyma-system ",
			wantErr:         true,
		},
		{
			name:            "invalid module namespace - contains capital letters",
			moduleNamespace: "Kyma-System",
			wantErr:         true,
		},
		{
			name:            "invalid module namespace - contains special characters",
			moduleNamespace: "kyma_system",
			wantErr:         true,
		},
		{
			name:            "invalid module namespace - starts with hyphen",
			moduleNamespace: "-kyma-system",
			wantErr:         true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validation.ValidateModuleNamespace(tt.moduleNamespace); (err != nil) != tt.wantErr {
				t.Errorf("ValidateModuleNamespace() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateNamespace(t *testing.T) {
	tests := []struct {
		name            string
		moduleNamespace string
		wantErr         bool
	}{
		{
			name:            "valid module namespace",
			moduleNamespace: "kyma-system",
			wantErr:         false,
		},
		{
			name:            "empty module namespace",
			moduleNamespace: "",
			wantErr:         true,
		},
		{
			name:            "invalid module namespace - whitespaces",
			moduleNamespace: " kyma-system ",
			wantErr:         true,
		},
		{
			name:            "invalid module namespace - contains capital letters",
			moduleNamespace: "Kyma-System",
			wantErr:         true,
		},
		{
			name:            "invalid module namespace - contains special characters",
			moduleNamespace: "kyma_system",
			wantErr:         true,
		},
		{
			name:            "invalid module namespace - starts with hyphen",
			moduleNamespace: "-kyma-system",
			wantErr:         true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validation.ValidateNamespace(tt.moduleNamespace); (err != nil) != tt.wantErr {
				t.Errorf("ValidateModuleNamespace() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateResources(t *testing.T) {
	tests := []struct {
		name      string
		resources contentprovider.Resources
		wantErr   bool
	}{
		{
			name: "valid resources",
			resources: contentprovider.Resources{
				"first":  "https://github.com/kyma-project/template-operator/releases/download/1.0.1/template-operator.yaml",
				"second": "https://github.com/kyma-project/template-operator/releases/download/1.0.1/template-operator.yaml",
			},
			wantErr: false,
		},
		{
			name: "empty name",
			resources: contentprovider.Resources{
				"": "https://github.com/kyma-project/template-operator/releases/download/1.0.1/template-operator.yaml",
			},
			wantErr: true,
		},
		{
			name: "empty link",
			resources: contentprovider.Resources{
				"first": "",
			},
			wantErr: true,
		},
		{
			name: "non-https schema",
			resources: contentprovider.Resources{
				"first": "http://github.com/kyma-project/template-operator/releases/download/1.0.1/template-operator.yaml",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validation.ValidateResources(tt.resources); (err != nil) != tt.wantErr {
				fmt.Println(err.Error())
				t.Errorf("ValidateResources() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateIsValidHttpsUrl(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "valid url",
			url:     "https://github.com/kyma-project/template-operator/releases/download/1.0.1/template-operator.yaml",
			wantErr: false,
		},
		{
			name:    "invalid url - not using https",
			url:     "http://github.com/kyma-project/template-operator/releases/download/1.0.1/template-operator.yaml",
			wantErr: true,
		},
		{
			name:    "invalid url - usig file scheme",
			url:     "file:///Users/User/template-operator/releases/download/1.0.1/template-operator.yaml",
			wantErr: true,
		},
		{
			name:    "invalid url - local path",
			url:     "./1.0.1/template-operator.yaml",
			wantErr: true,
		},
		{
			name:    "invalid url",
			url:     "%% not a valid url",
			wantErr: true,
		},
		{
			name:    "empty url",
			url:     "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validation.ValidateIsValidHTTPSURL(tt.url); (err != nil) != tt.wantErr {
				fmt.Println(err.Error())
				t.Errorf("ValidateIsValidUrl() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
