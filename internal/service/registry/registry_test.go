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
	_, err := registry.NewService(nil, repo, nil)
	require.Error(t, err)

	_, err = registry.NewService(&ociRepositoryVersionExistsStub{}, repo, nil)
	require.Error(t, err)

	_, err = registry.NewService(&ociRepositoryVersionExistsStub{}, nil, nil)
	require.Error(t, err)
}

func TestServiceExistsComponentVersion_WhenCredResolverReturnsError_ReturnsErr(t *testing.T) {
	svc, _ := registry.NewService(&ociRepositoryVersionExistsStub{}, nil, errResolverFunc)

	_, err := svc.ExistsComponentVersion(&comparch.ComponentArchive{}, true, "", "ghcr.io/template-operator")

	require.ErrorContains(t, err, "failed to resolve credentials")
	require.ErrorContains(t, err, "could not get repository")
}

func TestServicePushComponentVersion_WhenCredResolverReturnsError_ReturnsErr(t *testing.T) {
	svc, _ := registry.NewService(&ociRepositoryVersionExistsStub{}, nil, errResolverFunc)

	err := svc.PushComponentVersion(&comparch.ComponentArchive{}, true, true, "creds", "ghcr.io/template-operator")

	require.ErrorContains(t, err, "failed to resolve credentials")
	require.ErrorContains(t, err, "could not get repository")
}

func TestServiceGetComponentVersion_WhenCredResolverReturnsError_ReturnsErr(t *testing.T) {
	svc, _ := registry.NewService(&ociRepositoryVersionExistsStub{}, nil, errResolverFunc)

	_, err := svc.GetComponentVersion(&comparch.ComponentArchive{}, true, "creds", "ghcr.io/template-operator")

	require.ErrorContains(t, err, "failed to resolve credentials")
	require.ErrorContains(t, err, "could not get repository")
}

func TestService_PushComponentVersion_ReturnErrorWhenSameComponentVersionExists(t *testing.T) {
	repo, err := ocireg.NewRepository(cpi.DefaultContext(), "URL")
	require.NoError(t, err)
	componentArchive := &comparch.ComponentArchive{}

	svc, _ := registry.NewService(&ociRepositoryVersionExistsStub{}, repo, defaultCredsResolverFunc)

	err = svc.PushComponentVersion(componentArchive, true, false, "", "ghcr.io/template-operator")

	require.ErrorContains(t, err, "could not push component version")
}

func TestService_PushComponentVersion_ReturnNoErrorWhenSameComponentVersionExistsWithOverwrite(t *testing.T) {
	repo, err := ocireg.NewRepository(cpi.DefaultContext(), "URL")
	require.NoError(t, err)
	componentArchive := &comparch.ComponentArchive{}

	svc, _ := registry.NewService(&ociRepositoryStub{}, repo, defaultCredsResolverFunc)

	err = svc.PushComponentVersion(componentArchive, true, true, "", "ghcr.io/template-operator")

	require.NoError(t, err)
}

func TestService_PushComponentVersion_ReturnNoErrorOnSuccess(t *testing.T) {
	repo, err := ocireg.NewRepository(cpi.DefaultContext(), "URL")
	require.NoError(t, err)
	componentArchive := &comparch.ComponentArchive{}

	svc, _ := registry.NewService(&ociRepositoryStub{}, repo, defaultCredsResolverFunc)
	err = svc.PushComponentVersion(componentArchive, true, false, "", "ghcr.io/template-operator")
	require.NoError(t, err)
}

func TestService_GetComponentVersion_ReturnCorrectData(t *testing.T) {
	repo, err := ocireg.NewRepository(cpi.DefaultContext(), "URL")
	require.NoError(t, err)
	componentArchive := &comparch.ComponentArchive{}

	svc, _ := registry.NewService(&ociRepositoryStub{}, repo, defaultCredsResolverFunc)
	componentVersion, err := svc.GetComponentVersion(componentArchive, true, "", "ghcr.io/template-operator")
	require.NoError(t, err)
	require.Equal(t, &comparch.ComponentArchive{}, componentVersion)
}

func TestService_GetComponentVersion_ReturnErrorOnComponentVersionGetError(t *testing.T) {
	repo, err := ocireg.NewRepository(cpi.DefaultContext(), "URL")
	require.NoError(t, err)
	componentArchive := &comparch.ComponentArchive{}

	svc, _ := registry.NewService(&ociRepositoryNotExistStub{}, repo, defaultCredsResolverFunc)
	_, err = svc.GetComponentVersion(componentArchive, true, "", "ghcr.io/template-operator")
	require.ErrorContains(t, err, "could not get component version")
}

func Test_ConstructRegistryUrl_ReturnsCorrectWithHTTPAndNotInsecure(t *testing.T) {
	scheme := registry.ConstructRegistryUrl("http://ghcr.io", false)

	require.Equal(t, "ghcr.io", scheme)
}

func Test_ConstructRegistryUrl_ReturnsCorrectWithHTTPSAndNotInsecure(t *testing.T) {
	scheme := registry.ConstructRegistryUrl("https://ghcr.io", false)

	require.Equal(t, "ghcr.io", scheme)
}

func Test_NoSchemeURL_ReturnsCorrectWithNoSchemeAndNotInsecure(t *testing.T) {
	scheme := registry.ConstructRegistryUrl("ghcr.io", false)

	require.Equal(t, "ghcr.io", scheme)
}

func Test_ConstructRegistryUrl_ReturnsCorrectWithHTTPAndInsecure(t *testing.T) {
	scheme := registry.ConstructRegistryUrl("http://ghcr.io", true)

	require.Equal(t, "http://ghcr.io", scheme)
}

func Test_ConstructRegistryUrl_ReturnsCorrectWithHTTPSAndInsecure(t *testing.T) {
	scheme := registry.ConstructRegistryUrl("https://ghcr.io", true)

	require.Equal(t, "http://ghcr.io", scheme)
}

func Test_NoSchemeURL_ReturnsCorrectWithNoSchemeAndInsecure(t *testing.T) {
	scheme := registry.ConstructRegistryUrl("ghcr.io", true)

	require.Equal(t, "http://ghcr.io", scheme)
}

func Test_ExistsComponentVersion_Exists(t *testing.T) {
	repo, err := ocireg.NewRepository(cpi.DefaultContext(), "URL")
	require.NoError(t, err)
	componentArchive := &comparch.ComponentArchive{}

	svc, _ := registry.NewService(&ociRepositoryVersionExistsStub{}, repo, defaultCredsResolverFunc)
	exists, err := svc.ExistsComponentVersion(componentArchive, true, "", "ghcr.io/template-operator")
	require.NoError(t, err)
	require.True(t, exists)
}

func Test_ExistsComponentVersion_NotExists(t *testing.T) {
	repo, err := ocireg.NewRepository(cpi.DefaultContext(), "URL")
	require.NoError(t, err)
	componentArchive := &comparch.ComponentArchive{}

	svc, _ := registry.NewService(&ociRepositoryNotExistStub{}, repo, defaultCredsResolverFunc)
	exists, err := svc.ExistsComponentVersion(componentArchive, true, "", "ghcr.io/template-operator")
	require.NoError(t, err)
	require.False(t, exists)
}

func Test_ExistsComponentVersion_Error(t *testing.T) {
	repo, err := ocireg.NewRepository(cpi.DefaultContext(), "URL")
	require.NoError(t, err)
	componentArchive := &comparch.ComponentArchive{}

	svc, _ := registry.NewService(&ociRepositoryStub{err: errors.New("test error")}, repo, defaultCredsResolverFunc)
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

func errResolverFunc(_ cpi.Context, _ string, _ string) (credentials.Credentials, error) {
	return nil, errors.New("nil resolver function called")
}

func defaultCredsResolverFunc(_ cpi.Context, _ string, _ string) (credentials.Credentials, error) {
	return credentials.NewCredentials(nil), nil
}
