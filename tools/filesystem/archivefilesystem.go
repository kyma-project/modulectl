package filesystem

import (
	"fmt"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"os"
)

type ArchiveFileSystem struct {
	MemoryFileSystem vfs.FileSystem
}

func NewArchiveFileSystem(memoryFileSystem vfs.FileSystem) *ArchiveFileSystem {
	return &ArchiveFileSystem{
		MemoryFileSystem: memoryFileSystem,
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

func (s *ArchiveFileSystem) GetArchiveFileSystem() vfs.FileSystem {
	return s.MemoryFileSystem
}
