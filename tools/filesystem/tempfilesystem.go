package filesystem

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strings"
	"time"
)

const (
	httpGetTimeout = 20 * time.Second
	maxRedirects   = 10
)

var (
	errBadHTTPStatus    = errors.New("bad http status")
	errSSRFRedirect     = errors.New("redirect blocked: destination is a private or non-https address")
	errTooManyRedirects = errors.New("redirect blocked: too many redirects")
)

func privateRanges() []*net.IPNet {
	cidrs := []string{
		"127.0.0.0/8",    // loopback
		"::1/128",        // IPv6 loopback
		"10.0.0.0/8",     // RFC1918
		"172.16.0.0/12",  // RFC1918
		"192.168.0.0/16", // RFC1918
		"169.254.0.0/16", // link-local / cloud metadata
		"fe80::/10",      // IPv6 link-local
		"fc00::/7",       // IPv6 unique-local (fc00::/7 covers fc00:: and fd00::)
	}
	ranges := make([]*net.IPNet, 0, len(cidrs))
	for _, cidr := range cidrs {
		_, ipNet, _ := net.ParseCIDR(cidr)
		ranges = append(ranges, ipNet)
	}
	return ranges
}

func isPrivateIP(host string) bool {
	// Strip zone identifier from scoped IPv6 literals (e.g. "fe80::1%eth0") so
	// net.ParseIP can classify them; zone IDs are not part of the routable address.
	if i := strings.IndexByte(host, '%'); i != -1 {
		host = host[:i]
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}
	return slices.ContainsFunc(privateRanges(), func(r *net.IPNet) bool {
		return r.Contains(ip)
	})
}

func checkRedirect(req *http.Request, via []*http.Request) error {
	if len(via) >= maxRedirects {
		return errTooManyRedirects
	}
	// Relative redirects have no host — the client resolves them against the original
	// (already-validated) host, so they cannot change scheme or destination.
	if !req.URL.IsAbs() {
		return nil
	}
	if req.URL.Scheme != "https" {
		return errSSRFRedirect
	}
	host := req.URL.Hostname()
	if isPrivateIP(host) {
		return errSSRFRedirect
	}
	addrs, err := net.DefaultResolver.LookupHost(req.Context(), host)
	if err != nil {
		return fmt.Errorf("failed to resolve redirect host %s: %w", host, err)
	}
	if slices.ContainsFunc(addrs, isPrivateIP) {
		return errSSRFRedirect
	}
	return nil
}

type TempFileSystem struct {
	files      []*os.File
	httpClient *http.Client
}

func NewTempFileSystem() *TempFileSystem {
	return &TempFileSystem{
		files: []*os.File{},
		httpClient: &http.Client{
			Timeout:       httpGetTimeout,
			CheckRedirect: checkRedirect,
		},
	}
}

func (fs *TempFileSystem) DownloadTempFile(dir, pattern string, url *url.URL) (string, error) {
	bytes, err := fs.getBytesFromURL(url)
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

// FileExists checks if a file exists at the given filePath and is a regular file.
func (fs *TempFileSystem) FileExists(filePath string) (bool, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil // File does not exist
		}
		return false, fmt.Errorf("failed to stat file %s: %w", filePath, err)
	}
	return info.Mode().IsRegular(), nil
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

func (fs *TempFileSystem) getBytesFromURL(url *url.URL) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), httpGetTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP GET request for %s: %w", url, err)
	}

	resp, err := fs.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP GET request failed for %s: %w", url, err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			err = fmt.Errorf("failed to close response body: %w", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"received status code %d for GET request to %s: %w",
			resp.StatusCode,
			url,
			errBadHTTPStatus,
		)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body from %s: %w", url, err)
	}

	return data, nil
}
