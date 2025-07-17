package image

import (
	"errors"
	"fmt"
	"strings"

	"github.com/distribution/reference"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var (
	ErrParserNil           = errors.New("parser cannot be nil")
	ErrEmptyImageURL       = errors.New("empty image URL")
	ErrImageNameExtraction = errors.New("could not extract image name")
	ErrNoTagOrDigest       = errors.New("no tag or digest found")
)

type ManifestParser interface {
	Parse(path string) ([]*unstructured.Unstructured, error)
}

type Service struct {
	manifestParser ManifestParser
}

func NewService(manifestParser ManifestParser) (*Service, error) {
	if manifestParser == nil {
		return nil, fmt.Errorf("manifestParser must not be nil: %w", ErrParserNil)
	}
	return &Service{manifestParser: manifestParser}, nil
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
