package filesystem

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

const httpGetTimeout = 20 * time.Second

var errBadHTTPStatus = errors.New("bad http status")

type TempFileSystem struct {
	files []*os.File
}

func NewTempFileSystem() *TempFileSystem {
	return &TempFileSystem{files: []*os.File{}}
}

func (fs *TempFileSystem) DownloadTempFile(dir, pattern string, url *url.URL) (string, error) {
	bytes, err := getBytesFromURL(url)
	if err != nil {
		return "", fmt.Errorf("failed to download file from %s: %w", url, err)
	}

	tmpFile, err := os.CreateTemp(dir, pattern)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file with pattern %s: %w", pattern, err)
	}
	defer tmpFile.Close()
	fs.files = append(fs.files, tmpFile)
	if _, err := tmpFile.Write(bytes); err != nil {
		return "", fmt.Errorf("failed to write to temp file %s: %w", tmpFile.Name(), err)
	}
	return tmpFile.Name(), nil
}

func (fs *TempFileSystem) RemoveTempFiles() []error {
	var errs []error
	for _, file := range fs.files {
		err := os.Remove(file.Name())
		if err != nil {
			errs = append(errs, err)
		}
	}
	fs.files = []*os.File{}
	return errs
}

func getBytesFromURL(url *url.URL) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), httpGetTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("http GET request failed for %s: %w", url, err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http GET request failed for %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: bad status for GET request to %s: %q", errBadHTTPStatus, url,
			resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body from %s: %w", url, err)
	}

	return data, nil
}
