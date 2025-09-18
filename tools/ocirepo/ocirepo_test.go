package ocirepo_test

import (
	"errors"
	"testing"

	mandelsofterrors "github.com/mandelsoft/goutils/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"ocm.software/ocm/api/ocm/cpi"

	"github.com/kyma-project/modulectl/tools/ocirepo"
)

var (
	name    = "kyma-project.io/module/template-operator"
	version = "1.0.0"
)

func Test_ExistsComponentDescriptor_Exists(t *testing.T) {
	repo := &repoStub{exists: true}
	ociRepo := &ocirepo.OCIRepo{}

	exists, err := ociRepo.ExistsComponentVersion(&archiveMeta{name, version}, repo)

	require.NoError(t, err)
	assert.True(t, exists)
}

func Test_ExistsComponentDescriptor_NotExists(t *testing.T) {
	repo := &repoStub{exists: false}
	ociRepo := &ocirepo.OCIRepo{}

	exists, err := ociRepo.ExistsComponentVersion(&archiveMeta{name, version}, repo)

	require.NoError(t, err)
	assert.False(t, exists)
}

func Test_ExistsComponentDescriptor_NotExists_NotFoundError(t *testing.T) {
	repo := &repoStub{exists: false, err: mandelsofterrors.ErrNotFound()}
	ociRepo := &ocirepo.OCIRepo{}

	exists, err := ociRepo.ExistsComponentVersion(&archiveMeta{name, version}, repo)

	require.NoError(t, err)
	assert.False(t, exists)
}

func Test_ExistsComponentDescriptor_Error_WrongName(t *testing.T) {
	repo := &repoStub{}
	ociRepo := &ocirepo.OCIRepo{}

	exists, err := ociRepo.ExistsComponentVersion(&archiveMeta{"wrong", version}, repo)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "wrong name passed")
	assert.False(t, exists)
}

func Test_ExistsComponentDescriptor_Error_WrongVersion(t *testing.T) {
	repo := &repoStub{}
	ociRepo := &ocirepo.OCIRepo{}

	exists, err := ociRepo.ExistsComponentVersion(&archiveMeta{name, "wrong"}, repo)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "wrong version passed")
	assert.False(t, exists)
}

type repoStub struct {
	cpi.Repository

	exists bool
	err    error
}

func (r *repoStub) ExistsComponentVersion(n, v string) (bool, error) {
	if n != name {
		return false, errors.New("wrong name passed")
	}

	if v != version {
		return false, errors.New("wrong version passed")
	}

	return r.exists, r.err
}

type archiveMeta struct {
	name    string
	version string
}

func (a *archiveMeta) GetName() string {
	return a.name
}

func (a *archiveMeta) GetVersion() string {
	return a.version
}
