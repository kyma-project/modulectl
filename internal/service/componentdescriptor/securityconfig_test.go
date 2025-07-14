package componentdescriptor_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"ocm.software/ocm/api/ocm/compdesc"

	"github.com/kyma-project/modulectl/internal/service/componentdescriptor"
	"github.com/kyma-project/modulectl/internal/service/contentprovider"
)

func Test_NewSecurityConfigService_ReturnsErrorOnNilImageService(t *testing.T) {
	securityConfigService, err := componentdescriptor.NewSecurityConfigService(&fileReaderStub{}, nil)
	require.ErrorContains(t, err, "imageService must not be nil")
	require.Nil(t, securityConfigService)
}

func Test_NewSecurityConfigService_ReturnsErrorOnNilFileReader(t *testing.T) {
	securityConfigService, err := componentdescriptor.NewSecurityConfigService(nil, &imageServiceStub{})
	require.ErrorContains(t, err, "fileReader must not be nil")
	require.Nil(t, securityConfigService)
}

func Test_AppendSecurityLabelsToSources_ReturnCorrectLabels(t *testing.T) {
	sources := compdesc.Sources{
		{
			SourceMeta: compdesc.SourceMeta{
				Type: "Github",
				ElementMeta: compdesc.ElementMeta{
					Name:    "module-sources",
					Version: "1.0.0",
				},
			},
		},
	}

	securityConfig := contentprovider.SecurityScanConfig{
		RcTag:     "1.0.0",
		DevBranch: "main",
		Mend: contentprovider.MendSecConfig{
			Exclude:     []string{"**/test/**", "**/*_test.go"},
			SubProjects: "false",
			Language:    "golang-mod",
		},
	}

	err := componentdescriptor.AppendSecurityLabelsToSources(securityConfig, sources)
	require.NoError(t, err)

	require.Len(t, sources[0].Labels, 5)

	require.Equal(t, "scan.security.kyma-project.io/rc-tag", sources[0].Labels[0].Name)
	expectedLabel := json.RawMessage(`"1.0.0"`)
	require.Equal(t, expectedLabel, sources[0].Labels[0].Value)

	require.Equal(t, "scan.security.kyma-project.io/language", sources[0].Labels[1].Name)
	expectedLabel = json.RawMessage(`"golang-mod"`)
	require.Equal(t, expectedLabel, sources[0].Labels[1].Value)

	require.Equal(t, "scan.security.kyma-project.io/dev-branch", sources[0].Labels[2].Name)
	expectedLabel = json.RawMessage(`"main"`)
	require.Equal(t, expectedLabel, sources[0].Labels[2].Value)

	require.Equal(t, "scan.security.kyma-project.io/subprojects", sources[0].Labels[3].Name)
	expectedLabel = json.RawMessage(`"false"`)
	require.Equal(t, expectedLabel, sources[0].Labels[3].Value)

	require.Equal(t, "scan.security.kyma-project.io/exclude", sources[0].Labels[4].Name)
	expectedLabel = json.RawMessage(`"**/test/**,**/*_test.go"`)
	require.Equal(t, expectedLabel, sources[0].Labels[4].Value)
}

func TestSecurityConfigService_ParseSecurityConfigData_ReturnsCorrectData(t *testing.T) {
	securityConfigService, err := componentdescriptor.NewSecurityConfigService(&fileReaderStub{}, &imageServiceStub{})
	require.NoError(t, err)

	returned, err := securityConfigService.ParseSecurityConfigData("sec-scanners-config.yaml")
	require.NoError(t, err)

	require.Equal(t, securityConfig.RcTag, returned.RcTag)
	require.Equal(t, securityConfig.DevBranch, returned.DevBranch)
	require.Equal(t, securityConfig.Mend.Exclude, returned.Mend.Exclude)
	require.Equal(t, securityConfig.Mend.SubProjects, returned.Mend.SubProjects)
	require.Equal(t, securityConfig.Mend.Language, returned.Mend.Language)
}

func TestSecurityConfigService_ParseSecurityConfigData_ReturnErrOnFileExistenceCheckError(t *testing.T) {
	securityConfigService, err := componentdescriptor.NewSecurityConfigService(&fileReaderFileExistsErrorStub{}, &imageServiceStub{})
	require.NoError(t, err)

	_, err = securityConfigService.ParseSecurityConfigData("testFile")
	require.ErrorContains(t, err, "failed to check if security config file exists")
}

func TestSecurityConfigService_ParseSecurityConfigData_ReturnErrOnFileReadingError(t *testing.T) {
	securityConfigService, err := componentdescriptor.NewSecurityConfigService(&fileReaderReadFileErrorStub{}, &imageServiceStub{})
	require.NoError(t, err)

	_, err = securityConfigService.ParseSecurityConfigData("testFile")
	require.ErrorContains(t, err, "failed to read security config file")
}

func TestSecurityConfigService_ParseSecurityConfigData_ReturnErrOnFileDoesNotExist(t *testing.T) {
	securityConfigService, err := componentdescriptor.NewSecurityConfigService(&fileReaderFileExistsFalseStub{}, &imageServiceStub{})
	require.NoError(t, err)

	_, err = securityConfigService.ParseSecurityConfigData("testFile")
	require.ErrorContains(t, err, "security config file does not exist")
}

func TestSecurityConfigService_ParseSecurityConfigData_ReturnErrOnInvalidYAML(t *testing.T) {
	securityConfigService, err := componentdescriptor.NewSecurityConfigService(&fileReaderInvalidYAMLStub{}, &imageServiceStub{})
	require.NoError(t, err)

	_, err = securityConfigService.ParseSecurityConfigData("testFile")
	require.ErrorContains(t, err, "failed to unmarshal security config file")
}

func TestSecurityConfigService_AppendSecurityScanConfig_FailsOnImageExtraction(t *testing.T) {
	securityConfigService, err := componentdescriptor.NewSecurityConfigService(&fileReaderStub{}, &imageServiceExtractErrorStub{})
	require.NoError(t, err)

	descriptor := &compdesc.ComponentDescriptor{}

	err = securityConfigService.AppendSecurityScanConfig(descriptor, securityConfig, "manifest.yaml")
	require.ErrorContains(t, err, "failed to extract images from manifest")
}

func TestSecurityConfigService_AppendSecurityScanConfig_FailsOnAddImages(t *testing.T) {
	securityConfigService, err := componentdescriptor.NewSecurityConfigService(&fileReaderStub{}, &imageServiceAddErrorStub{})
	require.NoError(t, err)

	descriptor := &compdesc.ComponentDescriptor{}

	err = securityConfigService.AppendSecurityScanConfig(descriptor, securityConfig, "manifest.yaml")
	require.ErrorContains(t, err, "failed to add images to component descriptor")
}

func TestSecurityConfigService_MergeAndDeduplicateImages_OnlySecurityImages(t *testing.T) {
	securityConfigService, err := componentdescriptor.NewSecurityConfigService(&fileReaderStub{}, &imageServiceStub{})
	require.NoError(t, err)

	securityImages := []string{"image1:v1", "image2:v1"}
	result := securityConfigService.MergeAndDeduplicateImages(securityImages, []string{})

	require.Len(t, result, 2)
	require.Contains(t, result, "image1:v1")
	require.Contains(t, result, "image2:v1")
}

func TestSecurityConfigService_MergeAndDeduplicateImages_IdenticalImages(t *testing.T) {
	securityConfigService, err := componentdescriptor.NewSecurityConfigService(&fileReaderStub{}, &imageServiceStub{})
	require.NoError(t, err)

	images := []string{"image1:v1", "image2:v1"}
	result := securityConfigService.MergeAndDeduplicateImages(images, images)

	require.Len(t, result, 2)
	require.Contains(t, result, "image1:v1")
	require.Contains(t, result, "image2:v1")
}

func TestSecurityConfigService_MergeAndDeduplicateImages_EmptySlices(t *testing.T) {
	securityConfigService, err := componentdescriptor.NewSecurityConfigService(&fileReaderStub{}, &imageServiceStub{})
	require.NoError(t, err)

	result := securityConfigService.MergeAndDeduplicateImages([]string{}, []string{})
	require.Empty(t, result)
}

func TestSecurityConfigService_MergeAndDeduplicateImages(t *testing.T) {
	securityConfigService, err := componentdescriptor.NewSecurityConfigService(&fileReaderStub{}, &imageServiceStub{})
	require.NoError(t, err)

	// Test with overlapping images
	securityImages := []string{"image1:v1", "image2:v1", ""}
	manifestImages := []string{"image2:v1", "image3:v1", ""}

	result := securityConfigService.MergeAndDeduplicateImages(securityImages, manifestImages)

	require.Len(t, result, 3)
	require.Contains(t, result, "image1:v1")
	require.Contains(t, result, "image2:v1")
	require.Contains(t, result, "image3:v1")
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
	RcTag:     "1.0.0",
	DevBranch: "main",
	Mend: contentprovider.MendSecConfig{
		Exclude:     []string{"**/test/**", "**/*_test.go"},
		SubProjects: "false",
		Language:    "golang-mod",
	},
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

type imageServiceStub struct{}

func (*imageServiceStub) ExtractImagesFromManifest(_ string) ([]string, error) {
	return []string{
		"europe-docker.pkg.dev/kyma-project/prod/telemetry-manager:v1.2.0",
		"gcr.io/istio-release/proxyv2:1.19.0",
	}, nil
}

func (*imageServiceStub) AddImagesToOcmDescriptor(_ *compdesc.ComponentDescriptor, _ []string) error {
	return nil
}

type fileReaderFileExistsFalseStub struct{}

func (*fileReaderFileExistsFalseStub) FileExists(_ string) (bool, error) {
	return false, nil
}

func (*fileReaderFileExistsFalseStub) ReadFile(_ string) ([]byte, error) {
	return nil, nil
}

type imageServiceAddErrorStub struct{}

func (*imageServiceAddErrorStub) ExtractImagesFromManifest(_ string) ([]string, error) {
	return []string{"image1:v1"}, nil
}

func (*imageServiceAddErrorStub) AddImagesToOcmDescriptor(_ *compdesc.ComponentDescriptor, _ []string) error {
	return errors.New("add images error")
}

type fileReaderInvalidYAMLStub struct{}

func (*fileReaderInvalidYAMLStub) FileExists(_ string) (bool, error) {
	return true, nil
}

func (*fileReaderInvalidYAMLStub) ReadFile(_ string) ([]byte, error) {
	return []byte("invalid: yaml: content: ["), nil
}

type imageServiceExtractErrorStub struct{}

func (*imageServiceExtractErrorStub) ExtractImagesFromManifest(_ string) ([]string, error) {
	return nil, errors.New("extraction error")
}

func (*imageServiceExtractErrorStub) AddImagesToOcmDescriptor(_ *compdesc.ComponentDescriptor, _ []string) error {
	return nil
}
