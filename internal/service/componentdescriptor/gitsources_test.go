package componentdescriptor_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/kyma-project/modulectl/internal/common"
	"github.com/kyma-project/modulectl/internal/common/types/component"
	"github.com/kyma-project/modulectl/internal/service/componentdescriptor"
	"github.com/kyma-project/modulectl/internal/service/git"

	"github.com/stretchr/testify/require"

	"github.com/kyma-project/modulectl/internal/testutils"
)

func TestGitSourcesService_AddGitSources_ReturnsCorrectSources(t *testing.T) {
	gitSourcesService, err := componentdescriptor.NewGitSourcesService(&gitServiceStub{latestCommit: "latest"})
	require.NoError(t, err)
	moduleVersion := "1.0.0"
	descriptor := testutils.CreateComponentDescriptor("test.io/module/test", moduleVersion)

	err = gitSourcesService.AddGitSources(descriptor, "gitRepoPath", "gitRepoUrl", moduleVersion)

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

	err = gitSourcesService.AddGitSources(descriptor, "gitRepoPath", "gitRepoUrl", moduleVersion)
	require.Error(t, err)
	require.ErrorContains(t, err, "failed to get latest commit")
}

func TestGitSourcesService_AddGitSourcesToConstructor_AddsCorrectSource(t *testing.T) {
	gitSourcesService, err := componentdescriptor.NewGitSourcesService(&gitServiceStub{latestCommit: "abcdefg"})
	require.NoError(t, err)

	constructor := component.NewConstructor("test.io/module/test", "1.0.0")

	err = gitSourcesService.AddGitSourcesToConstructor(constructor, "gitRepoPath", "gitRepoUrl")

	require.NoError(t, err)
	require.Len(t, constructor.Components, 1)
	require.Len(t, constructor.Components[0].Sources, 1)
	source := constructor.Components[0].Sources[0]
	require.Equal(t, component.GithubSourceType, source.Type)
	require.Equal(t, "module-sources", source.Name)
	require.Equal(t, "1.0.0", source.Version)
	require.Len(t, source.Labels, 1)
	require.Equal(t, common.RefLabel, source.Labels[0].Name)
	require.Equal(t, git.HeadRef, source.Labels[0].Value)
	require.Equal(t, common.OCMVersion, source.Labels[0].Version)
	require.NotNil(t, source.Access)
	require.Equal(t, component.GithubAccessType, source.Access.Type)
	require.Equal(t, "gitRepoUrl", source.Access.RepoUrl)
	require.Equal(t, "abcdefg", source.Access.Commit)
}

func TestGitSourcesService_AddGitSourcesToConstructor_ReturnsErrorOnCommitRetrievalError(t *testing.T) {
	gitSourcesService, err := componentdescriptor.NewGitSourcesService(&gitServiceErrorStub{})
	require.NoError(t, err)

	constructor := component.NewConstructor("test.io/module/test", "1.0.0")

	err = gitSourcesService.AddGitSourcesToConstructor(constructor, "gitRepoPath", "gitRepoUrl")

	require.Error(t, err)
	require.ErrorContains(t, err, "failed to get latest commit")
	require.Len(t, constructor.Components[0].Sources, 0)
}

type gitServiceStub struct {
	latestCommit string
}

func (gs *gitServiceStub) GetLatestCommit(_ string) (string, error) {
	return gs.latestCommit, nil
}

type gitServiceErrorStub struct{}

func (*gitServiceErrorStub) GetLatestCommit(_ string) (string, error) {
	return "", errors.New("failed to get commit")
}
