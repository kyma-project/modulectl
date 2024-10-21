package fileresolver

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"

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

func (r *FileResolver) Resolve(file string) (string, error) {
	if parsedURL, err := r.ParseURL(file); err == nil {
		file, err = r.tempFileSystem.DownloadTempFile("", r.filePattern, parsedURL)
		if err != nil {
			return "", fmt.Errorf("failed to download file: %w", err)
		}
		return file, nil
	}

	if !filepath.IsAbs(file) {
		// Get the current working directory
		homeDir, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get the current directory: %w", err)
		}
		// Get the relative path from the current directory
		file = filepath.Join(homeDir, file)
		file, err = filepath.Abs(file)
		if err != nil {
			return "", fmt.Errorf("failed to obtain absolute path to file: %w", err)
		}
		return file, nil
	}

	return file, nil
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
