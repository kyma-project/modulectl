package filesystem_test

import (
	"archive/tar"
	"bytes"
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/mandelsoft/vfs/pkg/memoryfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kyma-project/modulectl/tools/filesystem"
)

func TestGenerateTarArchive(t *testing.T) {
	t.Run("should generate tar data successfully", func(t *testing.T) {
		// given
		expectedData := []byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.")

		mockFs := memoryfs.New()
		err := mockFs.MkdirAll("test/path", 0o755)
		require.NoError(t, err)
		inputFile, err := mockFs.Create("test/path/file.txt")
		require.NoError(t, err)
		_, err = inputFile.Write(expectedData)
		require.NoError(t, err)
		err = inputFile.Close()
		require.NoError(t, err)

		afs, err := filesystem.NewArchiveFileSystem(memoryfs.New(), mockFs)
		require.NoError(t, err)

		// when
		tarData, err := afs.ArchiveFile("test/path/file.txt")
		require.NoError(t, err)

		// then verify the tar archive is created correctly, including the padding etc.
		err = verifyTar(tarData)
		require.NoError(t, err)

		// and verify the contents of the tar archive is as expected
		buf := bytes.NewBuffer(tarData)
		tr := tar.NewReader(buf)
		header, err := tr.Next()
		require.NoError(t, err)
		assert.Equal(t, "file.txt", header.Name)
		assert.Equal(t, int64(len(expectedData)), header.Size)
		data, err := io.ReadAll(tr)
		require.NoError(t, err)
		assert.Equal(t, expectedData, data)
	})
	t.Run("should return an error on file not found", func(t *testing.T) {
		// given
		mockFs := memoryfs.New()
		err := mockFs.MkdirAll("test/path", 0o755)
		require.NoError(t, err)

		afs, err := filesystem.NewArchiveFileSystem(memoryfs.New(), mockFs)
		require.NoError(t, err)

		// when
		tarData, err := afs.ArchiveFile("test/path/file.txt")
		require.Error(t, err, "expected error when file does not exist")
		require.ErrorContains(t, err, "unable to get file info for", "error should be specific enough to identify it's origin")
		require.ErrorContains(t, err, "file does not exist", "error should contain additional details from the underlying file system")
		assert.Nil(t, tarData, "tarData should be nil when file does not exist")
	})
	t.Run("should return an error when file can't be open", func(t *testing.T) {
		// given
		memFs := memoryfs.New()
		err := memFs.MkdirAll("test/path", 0o755)
		require.NoError(t, err)
		_, err = memFs.Create("test/path/file.txt")
		require.NoError(t, err)

		errorFs := &mockFileSystem{
			FileSystem: memFs,
			openError:  errors.New("hard drive joined the resistance"),
		}

		afs, err := filesystem.NewArchiveFileSystem(memoryfs.New(), errorFs)
		require.NoError(t, err)

		// when
		tarData, err := afs.ArchiveFile("test/path/file.txt")
		require.Error(t, err, "expected error when file does not exist")
		require.ErrorContains(t, err, "unable to open file", "error should be specific enough to identify it's origin")
		require.ErrorContains(t, err, "hard drive joined the resistance", "error should contain additional details from the underlying file system")
		assert.Nil(t, tarData, "tarData should be nil when file does not exist")
	})

	t.Run("should return an error when error occurs during input file reading", func(t *testing.T) {
		// given
		memFs := memoryfs.New()
		err := memFs.MkdirAll("test/path", 0o755)
		require.NoError(t, err)
		_, err = memFs.Create("test/path/file.txt")
		require.NoError(t, err)

		errorFile := &mockFile{
			readError: errors.New("file decided to take a vacation"),
		}

		errorFs := &mockFileSystem{
			FileSystem: memFs,
			mockFile:   errorFile,
		}

		afs, err := filesystem.NewArchiveFileSystem(memoryfs.New(), errorFs)
		require.NoError(t, err)

		// when
		tarData, err := afs.ArchiveFile("test/path/file.txt")
		require.Error(t, err, "expected error when file does not exist")
		require.ErrorContains(t, err, "unable to copy file data", "error should be specific enough to identify it's origin")
		require.ErrorContains(t, err, "file decided to take a vacation", "error should contain additional details from the underlying file system")
		assert.Nil(t, tarData, "tarData should be nil when file does not exist")
	})
}

// VerifyTar inspects the tar archive in data[] and performs basic checks for GNU tar compliance:
//  1. Valid structure (can be fully read by archive/tar)
//  2. Proper 1024-byte trailer of zeroes at the end: "[...] an archive
//     consists of a series of file entries terminated by an end-of-archive entry, which consists of two 512 blocks of zero bytes."
func verifyTar(data []byte) error {
	// Structural check via tar.Reader
	r := bytes.NewReader(data)
	tr := tar.NewReader(r)

	for {
		_, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break // normal end of tar stream
		}
		if err != nil {
			return fmt.Errorf("tar parse error: %w", err)
		}
		// read contents to ensure full consumption
		if _, err := io.Copy(io.Discard, io.LimitReader(tr, 200)); err != nil {
			return fmt.Errorf("tar content read error: %w", err)
		}
	}

	// Trailer check: last 1024 bytes must be zero
	if len(data) < 1024 {
		return errors.New("tar file too short to contain proper trailer")
	}
	trailer := data[len(data)-1024:]
	for i, b := range trailer {
		if b != 0 {
			return fmt.Errorf("invalid trailer: byte %d is %d, expected 0", i, b)
		}
	}
	return nil
}

// A wrapper around vfs.File that allows to simulate read errors when a value of mockFile is used as a source in io.Copy().
type mockFile struct {
	vfs.File
	readError error
}

// This method will be used by io.Copy if a mockFile value is used as a source.
func (m *mockFile) WriteTo(w io.Writer) (int64, error) {
	if m.readError != nil {
		return 0, m.readError
	}
	panic("mockFile is intended to be used only when readError is set")
}

// If Close() is overridden like this it works even if the mockFile.File is nil
func (m *mockFile) Close() error {
	return nil
}

type mockFileSystem struct {
	vfs.FileSystem
	openError error
	mockFile  *mockFile
}

func (m *mockFileSystem) Open(name string) (vfs.File, error) {
	if m.openError != nil {
		return nil, m.openError
	}
	if m.mockFile != nil {
		return m.mockFile, nil
	}
	return m.FileSystem.Open(name)
}
