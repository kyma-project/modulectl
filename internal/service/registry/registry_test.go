package registry_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/extensions/repositories/comparch"
	"ocm.software/ocm/api/ocm/extensions/repositories/ocireg"

	"github.com/kyma-project/modulectl/internal/service/registry"
	"github.com/kyma-project/modulectl/tools/ocirepo"
)

func TestServiceNew_WhenCalledWithNilDependency_ReturnsErr(t *testing.T) {
	repo, _ := ocireg.NewRepository(cpi.DefaultContext(), "URL")
	_, err := registry.NewService(nil, repo)
	require.Error(t, err)

	_, err = registry.NewService(&ociRepositoryVersionExistsStub{}, nil)
	require.NoError(t, err)
}

func TestService_PushComponentVersion_ReturnErrorWhenSameComponentVersionExists(t *testing.T) {
	repo, err := ocireg.NewRepository(cpi.DefaultContext(), "URL")
	require.NoError(t, err)
	componentArchive := &comparch.ComponentArchive{}

	svc, _ := registry.NewService(&ociRepositoryVersionExistsStub{}, repo)

	err = svc.PushComponentVersion(componentArchive, true, false, "", "ghcr.io/template-operator")

	require.ErrorContains(t, err, "could not push component version")
}

func TestService_PushComponentVersion_ReturnNoErrorWhenSameComponentVersionExistsWithOverwrite(t *testing.T) {
	repo, err := ocireg.NewRepository(cpi.DefaultContext(), "URL")
	require.NoError(t, err)
	componentArchive := &comparch.ComponentArchive{}

	svc, _ := registry.NewService(&ociRepositoryStub{}, repo)

	err = svc.PushComponentVersion(componentArchive, true, true, "", "ghcr.io/template-operator")

	require.NoError(t, err)
}

func TestService_PushComponentVersion_ReturnNoErrorOnSuccess(t *testing.T) {
	repo, err := ocireg.NewRepository(cpi.DefaultContext(), "URL")
	require.NoError(t, err)
	componentArchive := &comparch.ComponentArchive{}

	svc, _ := registry.NewService(&ociRepositoryStub{}, repo)
	err = svc.PushComponentVersion(componentArchive, true, false, "", "ghcr.io/template-operator")
	require.NoError(t, err)
}

func TestService_GetComponentVersion_ReturnCorrectData(t *testing.T) {
	repo, err := ocireg.NewRepository(cpi.DefaultContext(), "URL")
	require.NoError(t, err)
	componentArchive := &comparch.ComponentArchive{}

	svc, _ := registry.NewService(&ociRepositoryStub{}, repo)
	componentVersion, err := svc.GetComponentVersion(componentArchive, true, "", "ghcr.io/template-operator")
	require.NoError(t, err)
	require.Equal(t, &comparch.ComponentArchive{}, componentVersion)
}

func TestService_GetComponentVersion_ReturnErrorOnComponentVersionGetError(t *testing.T) {
	repo, err := ocireg.NewRepository(cpi.DefaultContext(), "URL")
	require.NoError(t, err)
	componentArchive := &comparch.ComponentArchive{}

	svc, _ := registry.NewService(&ociRepositoryNotExistStub{}, repo)
	_, err = svc.GetComponentVersion(componentArchive, true, "", "ghcr.io/template-operator")
	require.ErrorContains(t, err, "could not get component version")
}

func Test_GetCredentials_ReturnUserPasswordWhenGiven(t *testing.T) {
	userPasswordCreds := "user1:pass1"
	creds := registry.GetCredentials(cpi.DefaultContext(), false, userPasswordCreds, "ghcr.io")

	require.Equal(t, "user1", creds.GetProperty("username"))
	require.Equal(t, "pass1", creds.GetProperty("password"))
}

func Test_GetCredentials_ReturnNilWhenInsecure(t *testing.T) {
	creds := registry.GetCredentials(cpi.DefaultContext(), true, "", "ghcr.io")

	require.Equal(t, credentials.NewCredentials(nil), creds)
}

func Test_NoSchemeURL_ReturnsCorrectWithHTTP(t *testing.T) {
	scheme := registry.NoSchemeURL("http://ghcr.io")

	require.Equal(t, "ghcr.io", scheme)
}

func Test_NoSchemeURL_ReturnsCorrectWithHTTPS(t *testing.T) {
	scheme := registry.NoSchemeURL("https://ghcr.io")

	require.Equal(t, "ghcr.io", scheme)
}

func Test_NoSchemeURL_ReturnsCorrectWithNoScheme(t *testing.T) {
	scheme := registry.NoSchemeURL("ghcr.io")

	require.Equal(t, "ghcr.io", scheme)
}

func Test_UserPass_ReturnsCorrectUsernameAndPassword(t *testing.T) {
	user, pass := registry.ParseUserPass("user1:pass1")
	require.Equal(t, "user1", user)
	require.Equal(t, "pass1", pass)
}

func Test_UserPass_ReturnsCorrectUsername(t *testing.T) {
	user, pass := registry.ParseUserPass("user1:")
	require.Equal(t, "user1", user)
	require.Empty(t, pass)
}

func Test_UserPass_ReturnsCorrectPassword(t *testing.T) {
	user, pass := registry.ParseUserPass(":pass1")
	require.Empty(t, user)
	require.Equal(t, "pass1", pass)
}

func Test_ExistsComponentVersion_Exists(t *testing.T) {
	repo, err := ocireg.NewRepository(cpi.DefaultContext(), "URL")
	require.NoError(t, err)
	componentArchive := &comparch.ComponentArchive{}

	svc, _ := registry.NewService(&ociRepositoryVersionExistsStub{}, repo)
	exists, err := svc.ExistsComponentVersion(componentArchive, true, "", "ghcr.io/template-operator")
	require.NoError(t, err)
	require.True(t, exists)
}

func Test_ExistsComponentVersion_NotExists(t *testing.T) {
	repo, err := ocireg.NewRepository(cpi.DefaultContext(), "URL")
	require.NoError(t, err)
	componentArchive := &comparch.ComponentArchive{}

	svc, _ := registry.NewService(&ociRepositoryNotExistStub{}, repo)
	exists, err := svc.ExistsComponentVersion(componentArchive, true, "", "ghcr.io/template-operator")
	require.NoError(t, err)
	require.False(t, exists)
}

func Test_ExistsComponentVersion_Error(t *testing.T) {
	repo, err := ocireg.NewRepository(cpi.DefaultContext(), "URL")
	require.NoError(t, err)
	componentArchive := &comparch.ComponentArchive{}

	svc, _ := registry.NewService(&ociRepositoryStub{err: errors.New("test error")}, repo)
	exists, err := svc.ExistsComponentVersion(componentArchive, true, "", "ghcr.io/template-operator")
	require.Error(t, err)
	require.Equal(t, "could not check if component version exists: test error", err.Error())
	require.False(t, exists)
}

type ociRepositoryVersionExistsStub struct{}

func (*ociRepositoryVersionExistsStub) GetComponentVersion(_ *comparch.ComponentArchive,
	_ cpi.Repository,
) (cpi.ComponentVersionAccess, error) {
	componentVersion := &comparch.ComponentArchive{}
	return componentVersion, nil
}

func (*ociRepositoryVersionExistsStub) PushComponentVersion(_ *comparch.ComponentArchive,
	_ cpi.Repository, _ bool,
) error {
	return errors.New("component version already exists")
}

func (*ociRepositoryVersionExistsStub) ExistsComponentVersion(_ ocirepo.ComponentArchiveMeta,
	_ cpi.Repository,
) (bool, error) {
	return true, nil
}

type ociRepositoryStub struct {
	err error
}

func (s *ociRepositoryStub) GetComponentVersion(_ *comparch.ComponentArchive,
	_ cpi.Repository,
) (cpi.ComponentVersionAccess, error) {
	componentVersion := &comparch.ComponentArchive{}
	return componentVersion, s.err
}

func (s *ociRepositoryStub) PushComponentVersion(_ *comparch.ComponentArchive,
	_ cpi.Repository, _ bool,
) error {
	return s.err
}

func (s *ociRepositoryStub) ExistsComponentVersion(_ ocirepo.ComponentArchiveMeta,
	_ cpi.Repository,
) (bool, error) {
	return false, s.err
}

type ociRepositoryNotExistStub struct{}

func (*ociRepositoryNotExistStub) GetComponentVersion(_ *comparch.ComponentArchive,
	_ cpi.Repository,
) (cpi.ComponentVersionAccess, error) {
	return nil, errors.New("failed to get component version")
}

func (*ociRepositoryNotExistStub) PushComponentVersion(_ *comparch.ComponentArchive,
	_ cpi.Repository, _ bool,
) error {
	return nil
}

func (*ociRepositoryNotExistStub) ExistsComponentVersion(_ ocirepo.ComponentArchiveMeta,
	_ cpi.Repository,
) (bool, error) {
	return false, nil
}
