package registry

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"fmt"
	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/credentials/extensions/repositories/dockerconfig"
	"ocm.software/ocm/api/oci/extensions/repositories/ocireg"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/extensions/repositories/comparch"
	"ocm.software/ocm/api/utils/runtime"
)

type OCIRepository interface {
	GetComponentVersion(archive *comparch.ComponentArchive, repo cpi.Repository) (cpi.ComponentVersionAccess, error)
	PushComponentVersionIfNotExist(archive *comparch.ComponentArchive, repo cpi.Repository) error
}

type Service struct {
	ociRepository OCIRepository
	repo          cpi.Repository
}

func NewService(ociRepository OCIRepository) *Service {
	return &Service{
		ociRepository: ociRepository,
	}
}

func (s *Service) PushComponentVersion(archive *comparch.ComponentArchive, insecure bool,
	credentials, registryURL string) error {
	repo, err := s.getRepository(insecure, credentials, registryURL)
	if err != nil {
		return fmt.Errorf("could not get repository: %w", err)
	}

	return s.ociRepository.PushComponentVersionIfNotExist(archive, repo)
}

func (s *Service) GetComponentVersion(archive *comparch.ComponentArchive, insecure bool,
	userPasswordCreds, registryURL string) (cpi.ComponentVersionAccess, error) {
	repo, err := s.getRepository(insecure, userPasswordCreds, registryURL)
	if err != nil {
		return nil, fmt.Errorf("could not get repository: %w", err)
	}

	return s.ociRepository.GetComponentVersion(archive, repo)
}

func (s *Service) getRepository(insecure bool, userPasswordCreds, registryURL string) (cpi.Repository, error) {
	if s.repo != nil {
		return s.repo, nil
	}

	ctx := cpi.DefaultContext()
	repoType := ocireg.Type
	registryURL = noSchemeURL(registryURL)
	if insecure {
		registryURL = "http://" + registryURL
	}
	creds := getCredentials(ctx, insecure, userPasswordCreds, registryURL)

	ociRepoSpec := &ocireg.RepositorySpec{
		ObjectVersionedType: runtime.NewVersionedObjectType(repoType),
		BaseURL:             registryURL,
	}

	ociRepo, err := ctx.RepositoryTypes().Convert(ociRepoSpec)
	if err != nil {
		return nil, fmt.Errorf("could not convert repository spec: %w", err)
	}

	repo, err := ctx.RepositoryForSpec(ociRepo, creds)
	s.repo = repo

	return repo, nil
}

func getCredentials(ctx cpi.Context, insecure bool, userPasswordCreds, registryURL string) credentials.Credentials {
	if insecure {
		return credentials.NewCredentials(nil)
	}

	var creds credentials.Credentials
	if home, err := os.UserHomeDir(); err == nil {
		path := filepath.Join(home, ".docker", "config.json")
		if repo, err := dockerconfig.NewRepository(ctx.CredentialsContext(), path, nil, true); err == nil {
			hostNameInDockerConfig := strings.Split(registryURL, "/")[0]
			if creds, err = repo.LookupCredentials(hostNameInDockerConfig); err != nil {
				creds = nil
			}
		}
	}

	if creds == nil || isEmptyAuth(creds) {
		user, pass := userPass(userPasswordCreds)
		creds = credentials.DirectCredentials{
			"username": user,
			"password": pass,
		}
	}

	return creds
}

func noSchemeURL(url string) string {
	regex := regexp.MustCompile(`^https?://`)
	return regex.ReplaceAllString(url, "")
}

func isEmptyAuth(creds credentials.Credentials) bool {
	if len(creds.GetProperty("auth")) != 0 {
		return false
	}
	if len(creds.GetProperty("username")) != 0 {
		return false
	}

	return true
}
func userPass(credentials string) (string, string) {
	u, p, found := strings.Cut(credentials, ":")
	if !found {
		return "", ""
	}
	return u, p
}
