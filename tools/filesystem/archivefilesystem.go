package filesystem

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/mandelsoft/vfs/pkg/vfs"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/utils/blobaccess"

	commonerrors "github.com/kyma-project/modulectl/internal/common/errors"
)

const tarMediaType = "application/x-tar"

type ArchiveFileSystem struct {
	MemoryFileSystem vfs.FileSystem
	OsFileSystem     vfs.FileSystem
}

func NewArchiveFileSystem(memoryFileSystem vfs.FileSystem, osFileSystem vfs.FileSystem) (*ArchiveFileSystem, error) {
	if memoryFileSystem == nil {
		return nil, fmt.Errorf("memoryFileSystem must not be nil: %w", commonerrors.ErrInvalidArg)
	}

	if osFileSystem == nil {
		return nil, fmt.Errorf("osFileSystem must not be nil: %w", commonerrors.ErrInvalidArg)
	}

	return &ArchiveFileSystem{
		MemoryFileSystem: memoryFileSystem,
		OsFileSystem:     osFileSystem,
	}, nil
}

func (s *ArchiveFileSystem) CreateArchiveFileSystem(path string) error {
	if err := s.MemoryFileSystem.MkdirAll(path, os.ModePerm); err != nil {
		return fmt.Errorf("unable to create directory %q: %w", path, err)
	}

	return nil
}

func (s *ArchiveFileSystem) WriteFile(data []byte, fileName string) error {
	file, err := s.MemoryFileSystem.Create(fileName)
	if err != nil {
		return fmt.Errorf("unable to create file %q: %w", fileName, err)
	}

	defer file.Close()
	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("unable to write data to file %q: %w", fileName, err)
	}

	return nil
}

func (s *ArchiveFileSystem) GenerateTarFileSystemAccess(filePath string) (cpi.BlobAccess, error) {
	tarData, err := generateTarData(s.OsFileSystem, filePath)
	if err != nil {
		return nil, err
	}
	return blobaccess.ForData(tarMediaType, tarData), nil
}

func generateTarData(filesystem vfs.FileSystem, filePath string) ([]byte, error) {
	fileInfo, err := filesystem.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("unable to get file info for %q: %w", filePath, err)
	}

	inputFile, err := filesystem.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("unable to open file %q: %w", filePath, err)
	}
	defer inputFile.Close()

	header, err := tar.FileInfoHeader(fileInfo, "")
	if err != nil {
		return nil, fmt.Errorf("unable to create header for file %q: %w", filePath, err)
	}
	outputBuffer := bytes.Buffer{}
	tarWriter := tar.NewWriter(&outputBuffer)

	if err := tarWriter.WriteHeader(header); err != nil {
		return nil, fmt.Errorf("unable to write header for %q: %w", filePath, err)
	}

	if _, err = io.Copy(tarWriter, inputFile); err != nil {
		return nil, fmt.Errorf("unable to copy file data: %w", err)
	}

	// Close the tar writer to flush the data.
	// I am not using defer for closing, because Close() on tarWriter appends a final padding to the tar archive, which is then not directly visible for the caller.
	// To make it visible to the caller the function return value must be re-assigned in the deferred code, which requires usage of named returns. The technique works, but the code is harder to understand.
	if err := tarWriter.Close(); err != nil {
		return nil, fmt.Errorf("unable to close tar writer: %w", err)
	}
	return outputBuffer.Bytes(), nil
}

func (s *ArchiveFileSystem) GetArchiveFileSystem() vfs.FileSystem {
	return s.MemoryFileSystem
}
