package git

import (
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/storage/memory"
)

const HeadRef = "HEAD"

type Service struct {
	latestCommit string
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) GetLatestCommit(repoURL string) (string, error) {
	if s.latestCommit != "" {
		return s.latestCommit, nil
	}

	repo, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
		URL:           repoURL,
		SingleBranch:  true,
		NoCheckout:    true,
		Depth:         1,
		ReferenceName: HeadRef,
	})
	if err != nil {
		return "", fmt.Errorf("failed to clone repo: %w", err)
	}

	ref, err := repo.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get head: %w", err)
	}

	s.latestCommit = ref.Hash().String()

	return s.latestCommit, nil
}
