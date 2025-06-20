package fileresolver

import (
	"fmt"
	"net/url"
	"path"

	commonerrors "github.com/kyma-project/modulectl/internal/common/errors"
	"github.com/kyma-project/modulectl/internal/service/contentprovider"
)

type TempFileSystem interface {
	DownloadTempFile(dir, pattern string, url *url.URL) (string, error)
	FileExists(filePath string) (bool, error)
	RemoveTempFiles() []error
}

type FileResolver struct {
	filePattern    string
	tempFileSystem TempFileSystem
}

func NewFileResolver(filePattern string, tempFileSystem TempFileSystem) (*FileResolver, error) {
	if filePattern == "" {
		return nil, fmt.Errorf("filePattern must not be empty: %w", commonerrors.ErrInvalidArg)
	}

	if tempFileSystem == nil {
		return nil, fmt.Errorf("tempFileSystem must not be nil: %w", commonerrors.ErrInvalidArg)
	}

	return &FileResolver{
		filePattern:    filePattern,
		tempFileSystem: tempFileSystem,
	}, nil
}

func (r *FileResolver) Resolve(fileRef contentprovider.UrlOrLocalFile, basePath string) (string, error) {
	if fileRef.IsEmpty() {
		return "", fmt.Errorf("file reference is empty: %w", commonerrors.ErrInvalidArg)
	}
	if fileRef.IsURL() {
		tempFilePath, err := r.tempFileSystem.DownloadTempFile("", r.filePattern, fileRef.URL())
		if err != nil {
			return "", fmt.Errorf("failed to download file: %w", err)
		}
		return tempFilePath, nil
	} else {
		finalPath := path.Join(basePath, fileRef.String())
		exists, err := r.tempFileSystem.FileExists(finalPath)
		if err != nil {
			return "", fmt.Errorf("failed to check if file exists %s: %w", finalPath, err)
		}
		if !exists {
			return "", fmt.Errorf("file does not exist: %s: %w", finalPath, commonerrors.ErrInvalidArg)
		}
		return finalPath, nil
	}
}

func (r *FileResolver) CleanupTempFiles() []error {
	return r.tempFileSystem.RemoveTempFiles()
}
