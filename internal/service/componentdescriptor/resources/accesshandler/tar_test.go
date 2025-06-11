package accesshandler_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"ocm.software/ocm/api/ocm/cpi"

	"github.com/kyma-project/modulectl/internal/service/componentdescriptor/resources/accesshandler"
)

func TestTar_GenerateBlobAccess(t *testing.T) {
	t.Run("should generate blob access successfully", func(t *testing.T) {
		// given
		expectedBlobAccess := &mockBlobAccess{}
		mockFS := &mockArchiveFileSystem{
			generateTarFunc: func(path string) (cpi.BlobAccess, error) {
				assert.Equal(t, "test/path", path)
				return expectedBlobAccess, nil
			},
		}

		tar := &accesshandler.Tar{
			FileSystem: mockFS,
			Path:       "test/path",
		}

		// when
		blobAccess, err := tar.GenerateBlobAccess()

		// then
		require.NoError(t, err)
		require.Equal(t, expectedBlobAccess, blobAccess)
	})

	t.Run("should return error when file system is nil", func(t *testing.T) {
		// given
		tar := accesshandler.NewTar(nil, "test/path")

		// when
		blobAccess, err := tar.GenerateBlobAccess()

		// then
		require.Error(t, err)
		require.Nil(t, blobAccess)
		require.ErrorIs(t, err, accesshandler.ErrNilFileSystem)
	})

	t.Run("should return error when generate tar fails", func(t *testing.T) {
		// given
		expectedError := errors.New("generation failed")
		mockFS := &mockArchiveFileSystem{
			generateTarFunc: func(path string) (cpi.BlobAccess, error) {
				return nil, expectedError
			},
		}

		tar := accesshandler.NewTar(mockFS, "test/path")

		// when
		blobAccess, err := tar.GenerateBlobAccess()

		// then
		require.Error(t, err)
		require.Nil(t, blobAccess)
		require.ErrorContains(t, err, "failed to generate tar file access")
		require.ErrorIs(t, err, expectedError)
	})
}

type mockArchiveFileSystem struct {
	generateTarFunc func(path string) (cpi.BlobAccess, error)
}

func (m *mockArchiveFileSystem) GenerateTarFileSystemAccess(filePath string) (cpi.BlobAccess, error) {
	return m.generateTarFunc(filePath)
}

type mockBlobAccess struct {
	cpi.BlobAccess
}
