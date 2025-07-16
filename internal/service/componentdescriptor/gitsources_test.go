package componentdescriptor_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kyma-project/modulectl/internal/service/componentdescriptor"
	"github.com/kyma-project/modulectl/internal/testutils"
)

func TestGitSourcesService_AddGitSources_ReturnsCorrectSources(t *testing.T) {
	gitSourcesService, err := componentdescriptor.NewGitSourcesService(&gitServiceStub{})
	require.NoError(t, err)
	moduleVersion := "1.0.0"
	descriptor := testutils.CreateComponentDescriptor("test.io/module/test", moduleVersion)

	err = gitSourcesService.AddGitSources(descriptor, "repoUrl", moduleVersion)

	require.NoError(t, err)
	require.Len(t, descriptor.Sources, 1)
	source := descriptor.Sources[0]
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
	gitSourcesService, err := componentdescriptor.NewGitSourcesService(&gitServiceErrorStub{})
	require.NoError(t, err)

	moduleVersion := "1.0.0"
	descriptor := testutils.CreateComponentDescriptor("test.io/module/test", moduleVersion)

	err = gitSourcesService.AddGitSources(descriptor, "repoUrl", moduleVersion)
	require.Error(t, err)
	require.ErrorContains(t, err, "failed to get latest commit")
}

type gitServiceStub struct{}

func (*gitServiceStub) GetLatestCommit(_ string) (string, error) {
	return "latest", nil
}

type gitServiceErrorStub struct{}

func (*gitServiceErrorStub) GetLatestCommit(_ string) (string, error) {
	return "", errors.New("failed to get commit")
}
