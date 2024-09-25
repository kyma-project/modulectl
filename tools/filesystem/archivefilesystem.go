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
		return nil, fmt.Errorf("%w: memoryFileSystem must not be nil", commonerrors.ErrInvalidArg)
	}

	if osFileSystem == nil {
		return nil, fmt.Errorf("%w: osFileSystem must not be nil", commonerrors.ErrInvalidArg)
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
	fileInfo, err := s.OsFileSystem.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("unable to get file info for %q: %w", filePath, err)
	}

	inputFile, err := s.OsFileSystem.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("unable to open file %q: %w", filePath, err)
	}
	defer inputFile.Close()

	header, err := tar.FileInfoHeader(fileInfo, "")
	if err != nil {
		return nil, fmt.Errorf("unable to create header for file %q: %w", filePath, err)
	}
	header.Name = fileInfo.Name()
	data := bytes.Buffer{}
	tarWriter := tar.NewWriter(&data)
	defer tarWriter.Close()

	// Write the header to the tar
	if err := tarWriter.WriteHeader(header); err != nil {
		return nil, fmt.Errorf("unable to write header for %q: %w", filePath, err)
	}

	if _, err := inputFile.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("unable to reset input file: %w", err)
	}

	if _, err := io.Copy(tarWriter, inputFile); err != nil {
		return nil, fmt.Errorf("unable to copy file: %w", err)
	}

	return blobaccess.ForData(tarMediaType, data.Bytes()), nil
}

func (s *ArchiveFileSystem) GetArchiveFileSystem() vfs.FileSystem {
	return s.MemoryFileSystem
}
