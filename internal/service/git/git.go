package git

import (
	"fmt"

	"github.com/go-git/go-git/v5"
)

type Service struct {
	latestCommit string
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) GetLatestCommit(gitRepoPath string) (string, error) {
	if s.latestCommit != "" {
		return s.latestCommit, nil
	}

	repo, err := git.PlainOpen(gitRepoPath)
	if err != nil {
		return "", fmt.Errorf("failed to open repo: %w", err)
	}

	ref, err := repo.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get head: %w", err)
	}

	s.latestCommit = ref.Hash().String()

	return s.latestCommit, nil
}
