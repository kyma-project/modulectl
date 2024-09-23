package componentdescriptor_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/kyma-project/modulectl/internal/service/componentdescriptor"
	"github.com/stretchr/testify/require"
	"ocm.software/ocm/api/ocm/compdesc"
)

func TestGitSourcesService_AddGitSources_ReturnsCorrectSources(t *testing.T) {
	gitSourcesService := componentdescriptor.NewGitSourcesService(&gitServiceStub{})

	cd := &compdesc.ComponentDescriptor{}
	cd.SetName("test.io/module/test")
	moduleVersion := "1.0.0"
	cd.SetVersion(moduleVersion)

	err := gitSourcesService.AddGitSources(cd, "repoUrl", moduleVersion)
	require.NoError(t, err)
	require.Len(t, cd.Sources, 1)

	source := cd.Sources[0]
	require.Equal(t, "Github", source.Type)
	require.Equal(t, "module-sources", source.Name)
	require.Equal(t, moduleVersion, source.Version)
	require.Len(t, source.Labels, 1)
	require.Equal(t, "git.kyma-project.io/ref", source.Labels[0].Name)
	require.Equal(t, "v1", source.Labels[0].Version)
	expectedLabel := json.RawMessage(`"HEAD"`)
	require.Equal(t, expectedLabel, source.Labels[0].Value)
}

func TestGitSourcesService_AddGitSources_ReturnsErrorOnCommitRetrievalError(t *testing.T) {
	gitSourcesService := componentdescriptor.NewGitSourcesService(&gitServiceErrorStub{})

	cd := &compdesc.ComponentDescriptor{}
	cd.SetName("test.io/module/test")
	moduleVersion := "1.0.0"
	cd.SetVersion(moduleVersion)

	err := gitSourcesService.AddGitSources(cd, "repoUrl", moduleVersion)
	require.Error(t, err)
	require.ErrorContains(t, err, "failed to get latest commit")
}

type gitServiceStub struct {
}

func (*gitServiceStub) GetLatestCommit(_ string) (string, error) {
	return "latest", nil
}

func (*gitServiceStub) GetRemoteGitFileContent(_, _, _ string) (string, error) {
	return "test", nil
}

type gitServiceErrorStub struct {
}

func (*gitServiceErrorStub) GetLatestCommit(_ string) (string, error) {
	return "", fmt.Errorf("failed to get commit")
}

func (*gitServiceErrorStub) GetRemoteGitFileContent(_, _, _ string) (string, error) {
	return "", fmt.Errorf("failed to get file content")
}
