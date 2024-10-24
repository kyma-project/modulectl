package fileresolver

import (
	"fmt"
	"net/url"

	commonerrors "github.com/kyma-project/modulectl/internal/common/errors"
)

type TempFileSystem interface {
	DownloadTempFile(dir, pattern string, url *url.URL) (string, error)
	RemoveTempFiles() []error
}

type FileResolver struct {
	filePattern    string
	tempFileSystem TempFileSystem
}

func NewFileResolver(filePattern string, tempFileSystem TempFileSystem) (*FileResolver, error) {
	if filePattern == "" {
		return nil, fmt.Errorf("%w: filePattern must not be empty", commonerrors.ErrInvalidArg)
	}

	if tempFileSystem == nil {
		return nil, fmt.Errorf("%w: tempFileSystem must not be nil", commonerrors.ErrInvalidArg)
	}

	return &FileResolver{
		filePattern:    filePattern,
		tempFileSystem: tempFileSystem,
	}, nil
}

func (r *FileResolver) Resolve(fileName string) (string, error) {
	parsedURL, err := r.ParseURL(fileName)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL: %w", err)
	}

	tempFilePath, err := r.tempFileSystem.DownloadTempFile("", r.filePattern, parsedURL)
	if err != nil {
		return "", fmt.Errorf("failed to download file: %w", err)
	}
	return tempFilePath, nil
}

func (r *FileResolver) ParseURL(urlString string) (*url.URL, error) {
	urlParsed, err := url.Parse(urlString)
	if err == nil && urlParsed.Scheme != "" && urlParsed.Host != "" {
		return urlParsed, nil
	}
	return nil, fmt.Errorf("failed to parse url %s: %w", urlString, commonerrors.ErrInvalidArg)
}

func (r *FileResolver) CleanupTempFiles() []error {
	return r.tempFileSystem.RemoveTempFiles()
}
