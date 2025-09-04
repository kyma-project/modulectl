package componentdescriptor

import (
	"fmt"

	"github.com/kyma-project/modulectl/internal/common"
	"github.com/kyma-project/modulectl/internal/common/types/component"
	"ocm.software/ocm/api/ocm/compdesc"
	ocmv1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/github"
	"ocm.software/ocm/api/tech/github/identity"

	commonerrors "github.com/kyma-project/modulectl/internal/common/errors"
	"github.com/kyma-project/modulectl/internal/service/git"
)

type GitService interface {
	GetLatestCommit(gitRepoPath string) (string, error)
}

type GitSourcesService struct {
	gitService GitService
}

func NewGitSourcesService(gitService GitService) (*GitSourcesService, error) {
	if gitService == nil {
		return nil, fmt.Errorf("gitService must not be nil: %w", commonerrors.ErrInvalidArg)
	}

	return &GitSourcesService{
		gitService: gitService,
	}, nil
}

func (s *GitSourcesService) AddGitSources(componentDescriptor *compdesc.ComponentDescriptor,
	gitRepoPath, gitRepoURL, moduleVersion string,
) error {
	label, err := ocmv1.NewLabel(common.RefLabel, git.HeadRef, ocmv1.WithVersion(common.OCMVersion))
	if err != nil {
		return fmt.Errorf("failed to create label: %w", err)
	}

	sourceMeta := compdesc.SourceMeta{
		Type: identity.CONSUMER_TYPE,
		ElementMeta: compdesc.ElementMeta{
			Name:    common.OCMIdentityName,
			Version: moduleVersion,
			Labels:  ocmv1.Labels{*label},
		},
	}

	latestCommit, err := s.gitService.GetLatestCommit(gitRepoPath)
	if err != nil {
		return fmt.Errorf("failed to get latest commit: %w", err)
	}

	access := github.New(gitRepoURL, "", latestCommit)

	componentDescriptor.Sources = append(componentDescriptor.Sources, compdesc.Source{
		SourceMeta: sourceMeta,
		Access:     access,
	})

	return nil
}

func (s *GitSourcesService) AddGitSourcesToConstructor(constructor *component.Constructor,
	gitRepoPath, gitRepoURL string) error {
	latestCommit, err := s.gitService.GetLatestCommit(gitRepoPath)
	if err != nil {
		return fmt.Errorf("failed to get latest commit: %w", err)
	}

	constructor.AddGitSource(gitRepoURL, latestCommit)
	return nil
}
