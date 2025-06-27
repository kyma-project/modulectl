//nolint:testpackage //There is nothing wrong with testing non-exported functions using tests residing in the same package: https://pkg.go.dev/testing
package filesystem

import (
	"archive/tar"
	"bytes"
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/mandelsoft/vfs/pkg/memoryfs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateTarData(t *testing.T) {
	t.Run("should generate tar data successfully", func(t *testing.T) {
		// given
		expectedData := []byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.")

		memFs := memoryfs.New()
		err := memFs.MkdirAll("test/path", 0o755)
		require.NoError(t, err)
		inputFile, err := memFs.Create("test/path/file.txt")
		require.NoError(t, err)
		_, err = inputFile.Write(expectedData)
		require.NoError(t, err)
		err = inputFile.Close()
		require.NoError(t, err)

		// when
		tarData, err := generateTarData(memFs, "test/path/file.txt")
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
