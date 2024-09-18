package filesystem

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/mandelsoft/vfs/pkg/vfs"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/utils/blobaccess"
)

const tarMediaType = "application/x-tar"

type ArchiveFileSystem struct {
	MemoryFileSystem vfs.FileSystem
	OsFileSystem     vfs.FileSystem
}

func NewArchiveFileSystem(memoryFileSystem vfs.FileSystem, osFileSystem vfs.FileSystem) *ArchiveFileSystem {
	return &ArchiveFileSystem{
		MemoryFileSystem: memoryFileSystem,
		OsFileSystem:     osFileSystem,
	}
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

	inputBlob, err := s.OsFileSystem.OpenFile(filePath, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("unable to open file %q: %w", filePath, err)
	}

	data := bytes.Buffer{}
	tarWriter := tar.NewWriter(&data)
	defer tarWriter.Close()

	header, err := tar.FileInfoHeader(fileInfo, "")
	if err != nil {
		return nil, fmt.Errorf("unable to create file info header: %w", err)
	}
	header.Name = filePath

	header.AccessTime = time.Time{}
	header.ChangeTime = time.Time{}
	header.ModTime = time.Time{}

	if err := tarWriter.WriteHeader(header); err != nil {
		return nil, fmt.Errorf("unable to write header: %w", err)
	}

	if _, err := io.Copy(tarWriter, inputBlob); err != nil {
		return nil, fmt.Errorf("unable to copy file: %w", err)
	}

	return blobaccess.ForData(tarMediaType, data.Bytes()), nil
}

func (s *ArchiveFileSystem) GetArchiveFileSystem() vfs.FileSystem {
	return s.MemoryFileSystem
}
