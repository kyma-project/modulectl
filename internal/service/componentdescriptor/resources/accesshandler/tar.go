package accesshandler

import (
	"errors"
	"fmt"

	"ocm.software/ocm/api/ocm/cpi"
)

var ErrNilFileSystem = errors.New("file system must not be nil")

type ArchiveFileSystem interface {
	GenerateTarFileSystemAccess(filePath string) (cpi.BlobAccess, error)
}

type Tar struct {
	FileSystem ArchiveFileSystem
	Path       string
}

func (tarAccessHandler *Tar) GenerateBlobAccess() (cpi.BlobAccess, error) {
	if tarAccessHandler.FileSystem == nil {
		return nil, ErrNilFileSystem
	}

	access, err := tarAccessHandler.FileSystem.GenerateTarFileSystemAccess(tarAccessHandler.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tar file access, %w", err)
	}

	return access, nil
}
