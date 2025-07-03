package accesshandler_test

import (
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kyma-project/modulectl/internal/service/componentdescriptor/resources/accesshandler"
)

func TestTar_GenerateBlobAccess(t *testing.T) {
	t.Run("should generate blob access successfully", func(t *testing.T) {
		// given
		expectedBytes := []byte{1, 2, 3, 4, 5, 0, 0, 0, 0}
		mockFS := &mockArchiveFileSystem{
			generateTarFunc: func(path string) ([]byte, error) {
				assert.Equal(t, "test/path", path)
				return expectedBytes, nil
			},
		}

		tar := accesshandler.NewTar(mockFS, "test/path")

		assert.Equal(t, "test/path", tar.GetPath())

		// when
		blobAccess, err := tar.GenerateBlobAccess()

		// then
		require.NoError(t, err)
		rdr, err := blobAccess.Reader()
		require.NoError(t, err)
		defer rdr.Close()
		actualData, err := io.ReadAll(rdr)
		require.NoError(t, err)
		require.Equal(t, expectedBytes, actualData)
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
			generateTarFunc: func(path string) ([]byte, error) {
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
	generateTarFunc func(path string) ([]byte, error)
}

func (m *mockArchiveFileSystem) ArchiveFile(filePath string) ([]byte, error) {
	return m.generateTarFunc(filePath)
}
