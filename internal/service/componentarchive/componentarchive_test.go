package componentarchive_test

import (
	"errors"
	"testing"

	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/stretchr/testify/require"
	"ocm.software/ocm/api/ocm/cpi"

	"github.com/kyma-project/modulectl/internal/service/componentarchive"
	"github.com/kyma-project/modulectl/internal/testutils"
)

func TestCreateComponentArchive_IfArchiveFileSystemReturnsError_ReturnsWrappedError(t *testing.T) {
	mockFS := &mockArchiveFileSystem{
		CreateArchiveFileSystemFunc: func(path string) error {
			return errors.New("some fs error")
		},
	}
	service := componentarchive.NewService(mockFS)
	descriptor := testutils.CreateComponentDescriptor("test-module", "0.0.1")

	_, err := service.CreateComponentArchive(descriptor)

	require.Error(t, err)
	require.ErrorContains(t, err, "failed to create archive file system")
}

func TestCreateComponentArchive_WriteFileError(t *testing.T) {
	mockFS := &mockArchiveFileSystem{
		CreateArchiveFileSystemFunc: func(path string) error {
			return nil
		},
		WriteFileFunc: func(data []byte, fileName string) error {
			return errors.New("some write error")
		},
	}
	service := componentarchive.NewService(mockFS)
	descriptor := testutils.CreateComponentDescriptor("test-module", "0.0.1")

	archive, err := service.CreateComponentArchive(descriptor)

	require.Error(t, err)
	require.Nil(t, archive)
	require.ErrorContains(t, err, "failed to write to component descriptor file")
}

type mockArchiveFileSystem struct {
	CreateArchiveFileSystemFunc     func(path string) error
	WriteFileFunc                   func(data []byte, fileName string) error
	GetArchiveFileSystemFunc        func() vfs.FileSystem
	GenerateTarFileSystemAccessFunc func(filePath string) (cpi.BlobAccess, error)
}

func (m *mockArchiveFileSystem) CreateArchiveFileSystem(path string) error {
	return m.CreateArchiveFileSystemFunc(path)
}

func (m *mockArchiveFileSystem) WriteFile(data []byte, fileName string) error {
	return m.WriteFileFunc(data, fileName)
}

func (m *mockArchiveFileSystem) GetArchiveFileSystem() vfs.FileSystem {
	return m.GetArchiveFileSystemFunc()
}

func (m *mockArchiveFileSystem) GenerateTarFileSystemAccess(filePath string) (cpi.BlobAccess, error) {
	return m.GenerateTarFileSystemAccessFunc(filePath)
}
