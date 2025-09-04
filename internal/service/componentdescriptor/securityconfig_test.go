package componentdescriptor_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"ocm.software/ocm/api/ocm/compdesc"

	"github.com/kyma-project/modulectl/internal/common/types/component"
	"github.com/kyma-project/modulectl/internal/service/componentdescriptor"
	"github.com/kyma-project/modulectl/internal/service/contentprovider"
)

func Test_NewSecurityConfigService_ReturnsErrorOnNilFileReader(t *testing.T) {
	securityConfigService, err := componentdescriptor.NewSecurityConfigService(nil)
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
	securityConfigService, err := componentdescriptor.NewSecurityConfigService(&fileReaderStub{})
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

func TestSecurityConfigService_AppendSecurityScanConfigToConstructor_AddsCorrectLabels(t *testing.T) {
	securityConfigService, err := componentdescriptor.NewSecurityConfigService(&fileReaderStub{})
	require.NoError(t, err)

	constructor := component.NewConstructor("test-component", "1.0.0")
	constructor.AddGitSource("https://github.com/test/repo", "abc123")

	securityConfigService.AppendSecurityScanConfigToConstructor(constructor, securityConfig)

	comp := constructor.Components[0]
	require.Len(t, comp.Labels, 1)

	scanLabel := comp.Labels[0]
	require.Equal(t, "security.kyma-project.io/scan", scanLabel.Name)
	require.Equal(t, "enabled", scanLabel.Value)
	require.Equal(t, "v1", scanLabel.Version)

	require.Len(t, comp.Sources, 1)
	source := comp.Sources[0]
	require.Len(t, source.Labels, 6)

	expectedLabels := map[string]string{
		"scan.security.kyma-project.io/rc-tag":      "1.0.0",
		"scan.security.kyma-project.io/language":    "golang-mod",
		"scan.security.kyma-project.io/dev-branch":  "main",
		"scan.security.kyma-project.io/subprojects": "false",
		"scan.security.kyma-project.io/exclude":     "**/test/**,**/*_test.go",
	}

	for i := 1; i < len(source.Labels); i++ {
		label := source.Labels[i]
		expectedValue, exists := expectedLabels[label.Name]
		require.True(t, exists)
		require.Equal(t, expectedValue, label.Value)
		require.Equal(t, "v1", label.Version)
	}
}

func TestSecurityConfigService_AppendSecurityScanConfigToConstructor_WithEmptyExclude(t *testing.T) {
	securityConfigService, err := componentdescriptor.NewSecurityConfigService(&fileReaderStub{})
	require.NoError(t, err)

	constructor := component.NewConstructor("test-component", "1.0.0")
	constructor.AddGitSource("https://github.com/test/repo", "abc123")

	configWithEmptyExclude := contentprovider.SecurityScanConfig{
		RcTag:     "2.0.0",
		DevBranch: "develop",
		Mend: contentprovider.MendSecConfig{
			Exclude:     []string{},
			SubProjects: "true",
			Language:    "java",
		},
	}

	securityConfigService.AppendSecurityScanConfigToConstructor(constructor, configWithEmptyExclude)

	source := constructor.Components[0].Sources[0]
	var excludeLabel *component.Label
	for i := range source.Labels {
		if source.Labels[i].Name == "scan.security.kyma-project.io/exclude" {
			excludeLabel = &source.Labels[i]
			break
		}
	}
	require.NotNil(t, excludeLabel)
	require.Empty(t, excludeLabel.Value)
}

func TestSecurityConfigService_AppendSecurityScanConfigToConstructor_WithMultipleSources(t *testing.T) {
	securityConfigService, err := componentdescriptor.NewSecurityConfigService(&fileReaderStub{})
	require.NoError(t, err)

	constructor := component.NewConstructor("test-component", "1.0.0")
	constructor.AddGitSource("https://github.com/test/repo1", "abc123")
	constructor.AddGitSource("https://github.com/test/repo2", "def456")

	securityConfigService.AppendSecurityScanConfigToConstructor(constructor, securityConfig)

	require.Len(t, constructor.Components[0].Sources, 2)

	for _, source := range constructor.Components[0].Sources {
		require.Len(t, source.Labels, 6)

		var rcTagLabel *component.Label
		for i := range source.Labels {
			if source.Labels[i].Name == "scan.security.kyma-project.io/rc-tag" {
				rcTagLabel = &source.Labels[i]
				break
			}
		}
		require.NotNil(t, rcTagLabel)
		require.Equal(t, "1.0.0", rcTagLabel.Value)
	}
}

func TestSecurityConfigService_AppendSecurityScanConfigToConstructor_WithNoSources(t *testing.T) {
	securityConfigService, err := componentdescriptor.NewSecurityConfigService(&fileReaderStub{})
	require.NoError(t, err)

	constructor := component.NewConstructor("test-component", "1.0.0")

	securityConfigService.AppendSecurityScanConfigToConstructor(constructor, securityConfig)

	comp := constructor.Components[0]
	require.Len(t, comp.Labels, 1)

	scanLabel := comp.Labels[0]
	require.Equal(t, "security.kyma-project.io/scan", scanLabel.Name)
	require.Equal(t, "enabled", scanLabel.Value)

	require.Empty(t, comp.Sources)
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

type fileReaderFileExistsFalseStub struct{}

func (*fileReaderFileExistsFalseStub) FileExists(_ string) (bool, error) {
	return false, nil
}

func (*fileReaderFileExistsFalseStub) ReadFile(_ string) ([]byte, error) {
	return nil, nil
}
