package componentdescriptor_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/kyma-project/modulectl/internal/service/componentdescriptor"
	"github.com/kyma-project/modulectl/internal/service/contentprovider"
)

func Test_NewSecurityConfigService_ReturnsErrorOnNilFileReader(t *testing.T) {
	securityConfigService, err := componentdescriptor.NewSecurityConfigService(nil)
	require.ErrorContains(t, err, "fileReader must not be nil")
	require.Nil(t, securityConfigService)
}

func TestSecurityConfigService_ParseSecurityConfigData_ReturnsCorrectData(t *testing.T) {
	securityConfigService, err := componentdescriptor.NewSecurityConfigService(&fileReaderStub{})
	require.NoError(t, err)

	returned, err := securityConfigService.ParseSecurityConfigData("sec-scanners-config.yaml")
	require.NoError(t, err)

	require.Equal(t, securityConfig.ModuleName, returned.ModuleName)
	for i, image := range securityConfig.BDBA {
		require.Equal(t, image, returned.BDBA[i])
	}
}

func TestSecurityConfigService_ParseSecurityConfigData_ReturnErrOnFileExistenceCheckError(t *testing.T) {
	securityConfigService, err := componentdescriptor.NewSecurityConfigService(&fileReaderFileExistsErrorStub{})
	require.NoError(t, err)

	_, err = securityConfigService.ParseSecurityConfigData("testFile")
	require.ErrorContains(t, err, "failed to check if security config file exists")
}

func TestSecurityConfigService_ParseSecurityConfigData_ReturnErrOnFileReadingError(t *testing.T) {
	securityConfigService, err := componentdescriptor.NewSecurityConfigService(&fileReaderReadFileErrorStub{})
	require.NoError(t, err)

	_, err = securityConfigService.ParseSecurityConfigData("testFile")
	require.ErrorContains(t, err, "failed to read security config file")
}

func TestSecurityConfigService_ParseSecurityConfigData_ReturnErrOnFileDoesNotExist(t *testing.T) {
	securityConfigService, err := componentdescriptor.NewSecurityConfigService(&fileReaderFileExistsFalseStub{})
	require.NoError(t, err)

	_, err = securityConfigService.ParseSecurityConfigData("testFile")
	require.ErrorContains(t, err, "security config file does not exist")
}

type fileReaderStub struct{}

func (*fileReaderStub) FileExists(_ string) (bool, error) {
	return true, nil
}

func (*fileReaderStub) ReadFile(_ string) ([]byte, error) {
	securityConfigBytes, _ := yaml.Marshal(securityConfig)
	return securityConfigBytes, nil
}

var securityConfig = contentprovider.SecurityScanConfig{
	ModuleName: "test-module",
	BDBA:       []string{"image1", "image2"},
}

type fileReaderFileExistsErrorStub struct{}

func (*fileReaderFileExistsErrorStub) FileExists(_ string) (bool, error) {
	return false, errors.New("error while checking file existence")
}

func (*fileReaderFileExistsErrorStub) ReadFile(_ string) ([]byte, error) {
	return nil, errors.New("error while reading file")
}

type fileReaderReadFileErrorStub struct{}

func (*fileReaderReadFileErrorStub) FileExists(_ string) (bool, error) {
	return true, nil
}

func (*fileReaderReadFileErrorStub) ReadFile(_ string) ([]byte, error) {
	return nil, errors.New("error while reading file")
}

type fileReaderFileExistsFalseStub struct{}

func (*fileReaderFileExistsFalseStub) FileExists(_ string) (bool, error) {
	return false, nil
}

func (*fileReaderFileExistsFalseStub) ReadFile(_ string) ([]byte, error) {
	return nil, nil
}
