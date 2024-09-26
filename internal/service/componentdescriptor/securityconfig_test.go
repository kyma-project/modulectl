package componentdescriptor_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"ocm.software/ocm/api/ocm/compdesc"
	ocmv1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"

	"github.com/kyma-project/modulectl/internal/service/componentdescriptor"
	"github.com/kyma-project/modulectl/internal/service/contentprovider"
)

func Test_GetImageNameAndTag(t *testing.T) {
	tests := []struct {
		name              string
		imageURL          string
		expectedImageName string
		expectedImageTag  string
		expectedError     error
	}{
		{
			name:              "valid image URL",
			imageURL:          "docker.io/template-operator/test:latest",
			expectedImageName: "test",
			expectedImageTag:  "latest",
			expectedError:     nil,
		},
		{
			name:              "invalid image URL - no tag",
			imageURL:          "docker.io/template-operator/test",
			expectedImageName: "",
			expectedImageTag:  "",
			expectedError:     errors.New("invalid image URL"),
		},
		{
			name:              "invalid image URL - multiple tags",
			imageURL:          "docker.io/template-operator/test:latest:latest",
			expectedImageName: "",
			expectedImageTag:  "",
			expectedError:     errors.New("invalid image URL"),
		},
		{
			name:              "invalid image URL - no slashes",
			imageURL:          "docker.io",
			expectedImageName: "",
			expectedImageTag:  "",
			expectedError:     errors.New("invalid image URL"),
		},
		{
			name:              "invalid image URL - empty URL",
			imageURL:          "",
			expectedImageName: "",
			expectedImageTag:  "",
			expectedError:     errors.New("invalid image URL"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			imgName, imgTag, err := componentdescriptor.GetImageNameAndTag(test.imageURL)
			if test.expectedError != nil {
				require.ErrorContains(t, err, test.expectedError.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, test.expectedImageName, imgName)
				require.Equal(t, test.expectedImageTag, imgTag)
			}
		})
	}
}

func Test_AppendProtecodeImagesLayers_ReturnCorrectResources(t *testing.T) {
	cd := &compdesc.ComponentDescriptor{}
	cd.SetName("test.io/module/test")
	cd.SetVersion("1.0.0")
	cd.Provider = ocmv1.Provider{Name: "kyma"}

	securityConfig := contentprovider.SecurityScanConfig{
		Protecode: []string{
			"europe-docker.pkg.dev/kyma-project/prod/template-operator:1.0.0",
			"europe-docker.pkg.dev/kyma-project/prod/external/ghcr.io/mymodule/anotherimage:4.5.6",
		},
	}

	err := componentdescriptor.AppendProtecodeImagesLayers(cd, securityConfig)
	require.NoError(t, err)

	require.Equal(t, "template-operator", cd.Resources[0].Name)
	require.Equal(t, "1.0.0", cd.Resources[0].Version)

	require.Equal(t, "anotherimage", cd.Resources[1].Name)
	require.Equal(t, "4.5.6", cd.Resources[1].Version)

	for _, res := range cd.Resources {
		require.Equal(t, "ociArtifact", res.Type)
		require.Equal(t, "scan.security.kyma-project.io/type", res.Labels[0].Name)
		expectedLabel := json.RawMessage(`"third-party-image"`)
		require.Equal(t, expectedLabel, res.Labels[0].Value)
	}
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
		WhiteSource: contentprovider.WhiteSourceSecConfig{
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
	securityConfigService, err := componentdescriptor.NewSecurityConfigService(&gitServiceSecurityConfigStub{})
	require.NoError(t, err)

	returned, err := securityConfigService.ParseSecurityConfigData("https://github.com/kyma-project/template-operator",
		"sec-scanners-config.yaml")
	require.NoError(t, err)

	require.Equal(t, securityConfig.RcTag, returned.RcTag)
	require.Equal(t, securityConfig.DevBranch, returned.DevBranch)
	require.Equal(t, securityConfig.WhiteSource.Exclude, returned.WhiteSource.Exclude)
	require.Equal(t, securityConfig.WhiteSource.SubProjects, returned.WhiteSource.SubProjects)
	require.Equal(t, securityConfig.WhiteSource.Language, returned.WhiteSource.Language)
}

func TestSecurityConfigService_ParseSecurityConfigData_ReturnErrOnRemoteFileReadingError(t *testing.T) {
	securityConfigService, err := componentdescriptor.NewSecurityConfigService(&gitServiceNoRemoteFile{})
	require.NoError(t, err)

	_, err = securityConfigService.ParseSecurityConfigData("testUrl", "testFile")
	require.ErrorContains(t, err, "failed to get security config content")
}

type gitServiceSecurityConfigStub struct{}

func (g *gitServiceSecurityConfigStub) GetLatestCommit(_ string) (string, error) {
	return "latestCommit", nil
}

var securityConfig = contentprovider.SecurityScanConfig{
	RcTag:     "1.0.0",
	DevBranch: "main",
	WhiteSource: contentprovider.WhiteSourceSecConfig{
		Exclude:     []string{"**/test/**", "**/*_test.go"},
		SubProjects: "false",
		Language:    "golang-mod",
	},
}

func (g *gitServiceSecurityConfigStub) GetRemoteGitFileContent(_, _, _ string) (string, error) {
	securityConfigBytes, _ := yaml.Marshal(securityConfig)
	return string(securityConfigBytes), nil
}

type gitServiceNoRemoteFile struct{}

func (*gitServiceNoRemoteFile) GetLatestCommit(_ string) (string, error) {
	return "latestCommit", nil
}

func (*gitServiceNoRemoteFile) GetRemoteGitFileContent(_, _, _ string) (string, error) {
	return "", errors.New("error")
}
