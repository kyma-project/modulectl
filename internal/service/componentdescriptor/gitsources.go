package componentdescriptor

import (
	"fmt"

	"github.com/kyma-project/modulectl/internal/service/git"
	ocm "ocm.software/ocm/api/ocm/compdesc"
	ocmv1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/github"
	"ocm.software/ocm/api/tech/github/identity"
)

type GitService interface {
	GetLatestCommit(repoURL string) (string, error)
	GetRemoteGitFileContent(repoURL, commit, filePath string) (string, error)
}

type GitSourcesService struct {
	gitService GitService
}

func NewGitSourcesService(gitService GitService) *GitSourcesService {
	return &GitSourcesService{
		gitService: gitService,
	}
}

func (s *GitSourcesService) AddGitSources(componentDescriptor *ocm.ComponentDescriptor,
	gitRepoURL, moduleVersion string) error {
	label, err := ocmv1.NewLabel(refLabel, git.HeadRef, ocmv1.WithVersion(ocmVersion))
	if err != nil {
		return fmt.Errorf("failed to create label: %w", err)
	}

	sourceMeta := ocm.SourceMeta{
		Type: identity.CONSUMER_TYPE,
		ElementMeta: ocm.ElementMeta{
			Name:    ocmIdentityName,
			Version: moduleVersion,
			Labels:  ocmv1.Labels{*label},
		},
	}

	latestCommit, err := s.gitService.GetLatestCommit(gitRepoURL)
	if err != nil {
		return fmt.Errorf("failed to get latest commit: %w", err)
	}
	access := github.New(gitRepoURL, "", latestCommit)

	componentDescriptor.Sources = append(componentDescriptor.Sources, ocm.Source{
		SourceMeta: sourceMeta,
		Access:     access,
	})

	return nil
}
