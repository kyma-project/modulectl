package image

import (
	"errors"
	"fmt"
	"strings"

	"github.com/distribution/reference"
)

const (
	LatestTag = "latest"
	MainTag   = "main"
)

var (
	ErrEmptyImageURL       = errors.New("empty image URL")
	ErrImageNameExtraction = errors.New("could not extract image name")
	ErrNoTagOrDigest       = errors.New("no tag or digest found")
	ErrMissingImageTag     = errors.New("image is missing a tag")
	ErrDisallowedTag       = errors.New("image tag is disallowed (latest/main)")
)

func IsValidImage(value string) (bool, error) {
	if !isValidImageFormat(value) {
		return false, nil
	}

	_, tag, err := ParseImageReference(value)
	if err != nil {
		return false, fmt.Errorf("invalid image reference %q: %w", value, err)
	}

	if tag == "" {
		return false, fmt.Errorf("%w: %q", ErrMissingImageTag, value)
	}

	if isMainOrLatestTag(tag) {
		return false, fmt.Errorf("%w: %q", ErrDisallowedTag, tag)
	}

	return true, nil
}

func ParseImageReference(imageURL string) (string, string, error) {
	if imageURL == "" {
		return "", "", fmt.Errorf("failed to parse image reference: %w", ErrEmptyImageURL)
	}

	ref, err := reference.ParseAnyReference(imageURL)
	if err != nil {
		return "", "", fmt.Errorf("invalid image reference: %w", err)
	}

	var imageName string
	if named, ok := ref.(reference.Named); ok {
		parts := strings.Split(named.Name(), "/")
		imageName = parts[len(parts)-1]
	} else {
		return "", "", fmt.Errorf("failed to extract image name from %s: %w", imageURL, ErrImageNameExtraction)
	}

	switch t := ref.(type) {
	case reference.Tagged:
		return imageName, t.Tag(), nil
	case reference.Digested:
		return imageName, t.Digest().String(), nil
	default:
		return "", "", fmt.Errorf("no tag or digest found in %s: %w", imageURL, ErrNoTagOrDigest)
	}
}

func isValidImageFormat(value string) bool {
	if len(value) < 3 || len(value) > 256 {
		return false
	}

	hasTagOrDigest := false
	for _, c := range value {
		switch c {
		case ':', '@':
			hasTagOrDigest = true
		case ' ', '\t', '\n', '\r':
			return false
		}
	}

	return hasTagOrDigest
}

func isMainOrLatestTag(tag string) bool {
	switch strings.ToLower(tag) {
	case LatestTag, MainTag:
		return true
	default:
		return false
	}
}
