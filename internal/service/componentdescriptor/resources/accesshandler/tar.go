package accesshandler

import (
	"errors"
	"fmt"

	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/utils/blobaccess"
)

const tarMediaType = "application/x-tar"

var ErrNilFileSystem = errors.New("file system must not be nil")

type TarGenerator interface {
	ArchiveFile(filePath string) ([]byte, error)
}

type Tar struct {
	generator TarGenerator
	path      string
}

func NewTar(fs TarGenerator, path string) *Tar {
	return &Tar{
		generator: fs,
		path:      path,
	}
}

func (tarAccessHandler *Tar) GetPath() string {
	return tarAccessHandler.path
}

func (tarAccessHandler *Tar) GenerateBlobAccess() (cpi.BlobAccess, error) {
	if tarAccessHandler.generator == nil {
		return nil, ErrNilFileSystem
	}

	tarData, err := tarAccessHandler.generator.ArchiveFile(tarAccessHandler.path)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tar file access, %w", err)
	}
	return blobaccess.ForData(tarMediaType, tarData), nil
}
