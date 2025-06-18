package contentprovider

import (
	"fmt"
	"net/url"

	commonerrors "github.com/kyma-project/modulectl/internal/common/errors"
)

// UrlOrLocalFile is a struct that represents either a URL or a local file path.
//
//nolint:recvcheck // This is a value type, not a pointer type.
type UrlOrLocalFile struct {
	value string
	url   *url.URL
}

// MustUrlOrLocalFile is a helper function that parses a string into a UrlOrLocalFile. Use only in tests!
func MustUrlOrLocalFile(val string) UrlOrLocalFile {
	var res UrlOrLocalFile
	if err := res.FromString(val); err != nil {
		panic(fmt.Sprintf("failed to parse UrlOrLocalFile: %s, error: %v", val, err))
	}
	return res
}

func (u UrlOrLocalFile) IsURL() bool {
	return u.url != nil && len(u.url.Scheme) > 0 && len(u.url.Host) > 0
}

func (u UrlOrLocalFile) URL() *url.URL {
	return u.url
}

func (u UrlOrLocalFile) IsEmpty() bool {
	return u.value == ""
}

func (u UrlOrLocalFile) String() string {
	return u.value
}

func (u *UrlOrLocalFile) FromString(val string) error {
	if val == "" {
		// treat the empty val as it is a local file. For optional file references an empty val is expected.
		u.value = ""
		u.url = nil
		return nil
	}

	// Check if the string is a valid URL. url.Parse is very permissive - so we just check if the scheme and host is present.
	parsedURL, err := url.Parse(val)
	if err != nil {
		return fmt.Errorf("'%s' is not a valid URL: %w", val, commonerrors.ErrInvalidOption)
	}

	if len(parsedURL.Scheme) > 0 && len(parsedURL.Host) == 0 {
		return fmt.Errorf("'%s' is not a valid URL: Missing host: %w", val, commonerrors.ErrInvalidArg)
	}
	u.value = val
	u.url = parsedURL
	return nil
}

func (u *UrlOrLocalFile) UnmarshalYAML(unmarshal func(any) error) error {
	var raw string
	if err := unmarshal(&raw); err != nil {
		return err
	}

	return u.FromString(raw)
}

func (u UrlOrLocalFile) MarshalYAML() (any, error) {
	return u.value, nil
}
