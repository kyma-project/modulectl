package componentdescriptor_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kyma-project/modulectl/internal/common/types/component"
	"github.com/kyma-project/modulectl/internal/service/componentdescriptor"
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
	require.Empty(t, source.Labels)
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
	require.Empty(t, source.Labels)
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
	require.Empty(t, constructor.Components[0].Sources)
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
