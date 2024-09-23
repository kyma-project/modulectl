package componentdescriptor_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/kyma-project/modulectl/internal/service/componentdescriptor"
	"github.com/stretchr/testify/require"
	v1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
)

func Test_InitializeComponentDescriptor_ReturnsCorrectDescriptor(t *testing.T) {
	moduleName := "github.com/test-module"
	moduleVersion := "0.0.1"
	cd, err := componentdescriptor.InitializeComponentDescriptor(moduleName, moduleVersion)
	expectedProviderLabel := json.RawMessage(`"modulectl"`)

	require.NoError(t, err)
	require.Equal(t, moduleName, cd.GetName())
	require.Equal(t, moduleVersion, cd.GetVersion())
	require.Equal(t, "v2", cd.Metadata.ConfiguredVersion)
	require.Equal(t, v1.ProviderName("kyma-project.io"), cd.Provider.Name)
	require.Len(t, cd.Provider.Labels, 1)
	require.Equal(t, "kyma-project.io/built-by", cd.Provider.Labels[0].Name)
	require.Equal(t, expectedProviderLabel, cd.Provider.Labels[0].Value)
	require.Equal(t, "v1", cd.Provider.Labels[0].Version)
	require.Len(t, cd.Resources, 0)
}

func Test_InitializeComponentDescriptor_ReturnsErrWhenInvalidName(t *testing.T) {
	moduleName := "test-module"
	moduleVersion := "0.0.1"
	_, err := componentdescriptor.InitializeComponentDescriptor(moduleName, moduleVersion)

	expectedError := errors.New("failed to validate component descriptor")
	require.ErrorContains(t, err, expectedError.Error())
}
