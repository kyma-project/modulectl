package moduleconfigreader_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	commonerrors "github.com/kyma-project/modulectl/internal/common/errors"
	"github.com/kyma-project/modulectl/internal/service/contentprovider"
	moduleconfigreader "github.com/kyma-project/modulectl/internal/service/moduleconfig/reader"
)

const (
	moduleConfigFile = "config.yaml"
)

func Test_ParseModuleConfig_ReturnsError_WhenFileReaderReturnsError(t *testing.T) {
	result, err := moduleconfigreader.ParseModuleConfig(moduleConfigFile, &fileDoesNotExistStub{})

	require.ErrorIs(t, err, errReadingFile)
	require.Nil(t, result)
}

func Test_ParseModuleConfig_Returns_CorrectModuleConfig(t *testing.T) {
	result, err := moduleconfigreader.ParseModuleConfig(moduleConfigFile, &fileExistsStub{})

	require.NoError(t, err)
	require.Equal(t, "github.com/module-name", result.Name)
	require.Equal(t, "0.0.1", result.Version)
	require.Equal(t, "regular", result.Channel)
	require.Equal(t, "path/to/manifests", result.Manifest)
	require.Equal(t, "path/to/defaultCR", result.DefaultCR)
	require.Equal(t, "module-name-0.0.1", result.ResourceName)
	require.False(t, result.Mandatory)
	require.Equal(t, "kcp-system", result.Namespace)
	require.Equal(t, "path/to/securityConfig", result.Security)
	require.False(t, result.Internal)
	require.False(t, result.Beta)
	require.Equal(t, map[string]string{"label1": "value1"}, result.Labels)
	require.Equal(t, map[string]string{"annotation1": "value1"}, result.Annotations)
	require.Equal(t, contentprovider.ResourcesMap{
		"rawManifest": "https://github.com/kyma-project/template-operator/releases/download/1.0.1/template-operator.yaml",
	}, result.Resources)
	require.Equal(t, "manager-name", result.Manager.Name)
	require.Equal(t, "manager-namespace", result.Manager.Namespace)
	require.Equal(t, "apps", result.Manager.GroupVersionKind.Group)
	require.Equal(t, "v1", result.Manager.GroupVersionKind.Version)
	require.Equal(t, "Deployment", result.Manager.GroupVersionKind.Kind)
}

func TestNew_CalledWithNilDependencies_ReturnsErr(t *testing.T) {
	_, err := moduleconfigreader.NewService(nil)
	require.Error(t, err)
}

func Test_ValidateModuleConfig(t *testing.T) {
	tests := []struct {
		name          string
		moduleConfig  *contentprovider.ModuleConfig
		expectedError error
	}{
		{
			name:          "valid module config",
			moduleConfig:  &expectedReturnedModuleConfig,
			expectedError: nil,
		},
		{
			name: "invalid module name",
			moduleConfig: &contentprovider.ModuleConfig{
				Name:      "invalid name",
				Version:   "0.0.1",
				Channel:   "regular",
				Namespace: "kcp-system",
				Manifest:  "test",
			},
			expectedError: fmt.Errorf("failed to validate module name: %w", commonerrors.ErrInvalidOption),
		},
		{
			name: "invalid module version",
			moduleConfig: &contentprovider.ModuleConfig{
				Name:      "github.com/module-name",
				Version:   "invalid version",
				Channel:   "regular",
				Namespace: "kcp-system",
				Manifest:  "test",
			},
			expectedError: fmt.Errorf("failed to validate module version: %w", commonerrors.ErrInvalidOption),
		},
		{
			name: "invalid module channel",
			moduleConfig: &contentprovider.ModuleConfig{
				Name:      "github.com/module-name",
				Version:   "0.0.1",
				Channel:   "invalid channel",
				Namespace: "kcp-system",
				Manifest:  "test",
			},
			expectedError: fmt.Errorf("failed to validate module channel: %w", commonerrors.ErrInvalidOption),
		},
		{
			name: "invalid module namespace",
			moduleConfig: &contentprovider.ModuleConfig{
				Name:      "github.com/module-name",
				Version:   "0.0.1",
				Channel:   "regular",
				Namespace: "invalid namespace",
				Manifest:  "test",
			},
			expectedError: fmt.Errorf("failed to validate module namespace: %w", commonerrors.ErrInvalidOption),
		},
		{
			name: "empty manifest path",
			moduleConfig: &contentprovider.ModuleConfig{
				Name:      "github.com/module-name",
				Version:   "0.0.1",
				Channel:   "regular",
				Namespace: "kcp-system",
				Manifest:  "",
			},
			expectedError: fmt.Errorf("manifest must not be empty: %w", commonerrors.ErrInvalidOption),
		},
		{
			name: "invalid module resources - not a URL",
			moduleConfig: &contentprovider.ModuleConfig{
				Name:      "github.com/module-name",
				Version:   "0.0.1",
				Channel:   "regular",
				Namespace: "kcp-system",
				Manifest:  "test",
				Resources: contentprovider.ResourcesMap{
					"key": "%% not a URL",
				},
			},
			expectedError: fmt.Errorf("failed to validate resources: %w: link %%%% not a URL is not a valid URL", commonerrors.ErrInvalidOption),
		},
		{
			name: "invalid module resources - empty name",
			moduleConfig: &contentprovider.ModuleConfig{
				Name:      "github.com/module-name",
				Version:   "0.0.1",
				Channel:   "regular",
				Namespace: "kcp-system",
				Manifest:  "test",
				Resources: contentprovider.ResourcesMap{
					"": "https://github.com/kyma-project/template-operator/releases/download/1.0.1/template-operator.yaml",
				},
			},
			expectedError: fmt.Errorf("failed to validate resources: %w: name must not be empty", commonerrors.ErrInvalidOption),
		},
		{
			name: "invalid module resources - empty link",
			moduleConfig: &contentprovider.ModuleConfig{
				Name:      "github.com/module-name",
				Version:   "0.0.1",
				Channel:   "regular",
				Namespace: "kcp-system",
				Manifest:  "test",
				Resources: contentprovider.ResourcesMap{
					"name": "",
				},
			},
			expectedError: fmt.Errorf("failed to validate resources: %w: link must not be empty", commonerrors.ErrInvalidOption),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := moduleconfigreader.ValidateModuleConfig(test.moduleConfig)
			if test.expectedError != nil {
				require.ErrorContains(t, err, test.expectedError.Error())
				return
			}
			require.NoError(t, err)
		})
	}
}

func Test_ValidateManager(t *testing.T) {
	tests := []struct {
		name          string
		manager       *contentprovider.Manager
		expectedError error
	}{
		{
			name:          "nil manager",
			manager:       nil,
			expectedError: nil,
		},
		{
			name: "valid manager",
			manager: &contentprovider.Manager{
				Name: "manager-name",
				GroupVersionKind: metav1.GroupVersionKind{
					Group:   "apps",
					Version: "v1",
					Kind:    "Deployment",
				},
				Namespace: "manager-namespace",
			},
			expectedError: nil,
		},
		{
			name: "valid manager - empty namespace",
			manager: &contentprovider.Manager{
				Name: "manager-name",
				GroupVersionKind: metav1.GroupVersionKind{
					Group:   "apps",
					Version: "v1",
					Kind:    "Deployment",
				},
			},
			expectedError: nil,
		},
		{
			name: "invalid manager - empty name",
			manager: &contentprovider.Manager{
				Name:      "",
				Namespace: "manager-namespace",
				GroupVersionKind: metav1.GroupVersionKind{
					Group:   "apps",
					Version: "v1",
					Kind:    "Deployment",
				},
			},
			expectedError: fmt.Errorf("name must not be empty: %w", commonerrors.ErrInvalidOption),
		},
		{
			name: "invalid manager - empty kind",
			manager: &contentprovider.Manager{
				Name:      "manager-name",
				Namespace: "manager-namespace",
				GroupVersionKind: metav1.GroupVersionKind{
					Group:   "apps",
					Version: "v1",
				},
			},
			expectedError: fmt.Errorf("kind must not be empty: %w", commonerrors.ErrInvalidOption),
		},
		{
			name: "invalid manager - empty group",
			manager: &contentprovider.Manager{
				Name:      "manager-name",
				Namespace: "manager-namespace",
				GroupVersionKind: metav1.GroupVersionKind{
					Version: "v1",
					Kind:    "Deployment",
				},
			},
			expectedError: fmt.Errorf("group must not be empty: %w", commonerrors.ErrInvalidOption),
		},
		{
			name: "invalid manager - empty version",
			manager: &contentprovider.Manager{
				Name:      "manager-name",
				Namespace: "manager-namespace",
				GroupVersionKind: metav1.GroupVersionKind{
					Kind:  "Deployment",
					Group: "apps",
				},
			},
			expectedError: fmt.Errorf("version must not be empty: %w", commonerrors.ErrInvalidOption),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := moduleconfigreader.ValidateManager(test.manager)
			if test.expectedError != nil {
				require.ErrorContains(t, err, test.expectedError.Error())
				return
			}
			require.NoError(t, err)
		})
	}
}

// Test Stubs

type fileExistsStub struct{}

func (*fileExistsStub) FileExists(_ string) (bool, error) {
	return true, nil
}

var expectedReturnedModuleConfig = contentprovider.ModuleConfig{
	Name:         "github.com/module-name",
	Version:      "0.0.1",
	Channel:      "regular",
	Manifest:     "path/to/manifests",
	Mandatory:    false,
	DefaultCR:    "path/to/defaultCR",
	ResourceName: "module-name-0.0.1",
	Namespace:    "kcp-system",
	Security:     "path/to/securityConfig",
	Internal:     false,
	Beta:         false,
	Labels:       map[string]string{"label1": "value1"},
	Annotations:  map[string]string{"annotation1": "value1"},
	Resources: contentprovider.ResourcesMap{
		"rawManifest": "https://github.com/kyma-project/template-operator/releases/download/1.0.1/template-operator.yaml",
	},
	Manager: &contentprovider.Manager{
		Name:      "manager-name",
		Namespace: "manager-namespace",
		GroupVersionKind: metav1.GroupVersionKind{
			Group:   "apps",
			Version: "v1",
			Kind:    "Deployment",
		},
	},
}

func (*fileExistsStub) ReadFile(_ string) ([]byte, error) {
	return yaml.Marshal(expectedReturnedModuleConfig)
}

type fileDoesNotExistStub struct{}

func (*fileDoesNotExistStub) FileExists(_ string) (bool, error) {
	return false, nil
}

var errReadingFile = errors.New("some error reading file")

func (*fileDoesNotExistStub) ReadFile(_ string) ([]byte, error) {
	return nil, errReadingFile
}
