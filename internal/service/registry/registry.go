package registry

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/credentials/extensions/repositories/dockerconfig"
	"ocm.software/ocm/api/oci/extensions/repositories/ocireg"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/extensions/repositories/comparch"
	"ocm.software/ocm/api/utils/runtime"

	commonerrors "github.com/kyma-project/modulectl/internal/common/errors"
	"github.com/kyma-project/modulectl/tools/ocirepo"
)

type OCIRepository interface {
	GetComponentVersion(archive *comparch.ComponentArchive, repo cpi.Repository) (cpi.ComponentVersionAccess, error)
	PushComponentVersion(archive *comparch.ComponentArchive, repo cpi.Repository, overwrite bool) error
	ExistsComponentVersion(archive ocirepo.ComponentArchiveMeta, repo cpi.Repository) (bool, error)
}

type Service struct {
	ociRepository OCIRepository
	repo          cpi.Repository
}

func NewService(ociRepository OCIRepository, repo cpi.Repository) (*Service, error) {
	if ociRepository == nil {
		return nil, fmt.Errorf("ociRepository must not be nil: %w", commonerrors.ErrInvalidArg)
	}

	return &Service{
		ociRepository: ociRepository,
		repo:          repo,
	}, nil
}

func (s *Service) ExistsComponentVersion(archive *comparch.ComponentArchive,
	insecure bool,
	credentials string,
	registryURL string,
) (bool, error) {
	repo, err := s.getRepository(insecure, credentials, registryURL)
	if err != nil {
		return false, fmt.Errorf("could not get repository: %w", err)
	}

	exists, err := s.ociRepository.ExistsComponentVersion(archive, repo)
	if err != nil {
		return false, fmt.Errorf("could not check if component version exists: %w", err)
	}

	return exists, nil
}

func (s *Service) PushComponentVersion(archive *comparch.ComponentArchive, insecure, overwrite bool,
	credentials, registryURL string,
) error {
	repo, err := s.getRepository(insecure, credentials, registryURL)
	if err != nil {
		return fmt.Errorf("could not get repository: %w", err)
	}

	if err = s.ociRepository.PushComponentVersion(archive, repo, overwrite); err != nil {
		return fmt.Errorf("could not push component version: %w", err)
	}

	return nil
}

func (s *Service) GetComponentVersion(archive *comparch.ComponentArchive, insecure bool,
	userPasswordCreds, registryURL string,
) (cpi.ComponentVersionAccess, error) {
	repo, err := s.getRepository(insecure, userPasswordCreds, registryURL)
	if err != nil {
		return nil, fmt.Errorf("could not get repository: %w", err)
	}

	componentVersion, err := s.ociRepository.GetComponentVersion(archive, repo)
	if err != nil {
		return nil, fmt.Errorf("could not get component version: %w", err)
	}

	return componentVersion, nil
}

func (s *Service) getRepository(insecure bool, userPasswordCreds, registryURL string) (cpi.Repository, error) {
	if s.repo != nil {
		return s.repo, nil
	}

	ctx := cpi.DefaultContext()
	repoType := ocireg.Type
	registryURL = NoSchemeURL(registryURL)
	if insecure {
		registryURL = "http://" + registryURL
	}
	creds := GetCredentials(ctx, insecure, userPasswordCreds, registryURL)

	ociRepoSpec := &ocireg.RepositorySpec{
		ObjectVersionedType: runtime.NewVersionedObjectType(repoType),
		BaseURL:             registryURL,
	}

	ociRepo, err := ctx.RepositoryTypes().Convert(ociRepoSpec)
	if err != nil {
		return nil, fmt.Errorf("could not convert repository spec: %w", err)
	}

	repo, err := ctx.RepositoryForSpec(ociRepo, creds)
	if err != nil {
		return nil, fmt.Errorf("could not create repository from spec: %w", err)
	}

	s.repo = repo

	return repo, nil
}

func GetCredentials(ctx cpi.Context, insecure bool, userPasswordCreds, registryURL string) credentials.Credentials {
	if insecure {
		return credentials.NewCredentials(nil)
	}

	var creds credentials.Credentials
	user, pass := ParseUserPass(userPasswordCreds)

	if user != "" && pass != "" {
		creds = credentials.DirectCredentials{
			"username": user,
			"password": pass,
		}

		return creds
	}

	if home, err := os.UserHomeDir(); err == nil {
		path := filepath.Join(home, ".docker", "config.json")
		if repo, err := dockerconfig.NewRepository(ctx.CredentialsContext(), path, nil, true); err == nil {
			hostNameInDockerConfig := strings.Split(registryURL, "/")[0]
			if creds, err = repo.LookupCredentials(hostNameInDockerConfig); err != nil {
				creds = nil
			}
		}
	}

	return creds
}

func NoSchemeURL(url string) string {
	regex := regexp.MustCompile(`^https?://`)
	return regex.ReplaceAllString(url, "")
}

func ParseUserPass(credentials string) (string, string) {
	u, p, found := strings.Cut(credentials, ":")
	if !found {
		return "", ""
	}
	return u, p
}
