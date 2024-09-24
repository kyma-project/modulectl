package componentdescriptor_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	ocmv1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"

	"github.com/kyma-project/modulectl/internal/service/componentdescriptor"
)

func Test_InitializeComponentDescriptor_ReturnsCorrectDescriptor(t *testing.T) {
	moduleName := "github.com/test-module"
	moduleVersion := "0.0.1"
	descriptor, err := componentdescriptor.InitializeComponentDescriptor(moduleName, moduleVersion)
	expectedProviderLabel := json.RawMessage(`"modulectl"`)

	require.NoError(t, err)
	require.Equal(t, moduleName, descriptor.GetName())
	require.Equal(t, moduleVersion, descriptor.GetVersion())
	require.Equal(t, "v2", descriptor.Metadata.ConfiguredVersion)
	require.Equal(t, ocmv1.ProviderName("kyma-project.io"), descriptor.Provider.Name)
	require.Len(t, descriptor.Provider.Labels, 1)
	require.Equal(t, "kyma-project.io/built-by", descriptor.Provider.Labels[0].Name)
	require.Equal(t, expectedProviderLabel, descriptor.Provider.Labels[0].Value)
	require.Equal(t, "v1", descriptor.Provider.Labels[0].Version)
	require.Empty(t, descriptor.Resources)
}

func Test_InitializeComponentDescriptor_ReturnsErrWhenInvalidName(t *testing.T) {
	moduleName := "test-module"
	moduleVersion := "0.0.1"
	_, err := componentdescriptor.InitializeComponentDescriptor(moduleName, moduleVersion)

	expectedError := errors.New("failed to validate component descriptor")
	require.ErrorContains(t, err, expectedError.Error())
}
