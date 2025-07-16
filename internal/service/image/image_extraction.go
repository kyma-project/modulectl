package image

import (
	"errors"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	LatestTag       = "latest"
	MainTag         = "main"
	KindDeploymnent = "Deployment"
	KindStatefulSet = "StatefulSet"
)

var (
	ErrMissingImageTag = errors.New("image is missing a tag")
	ErrDisallowedTag   = errors.New("image tag is disallowed (latest/main)")
)

func (s *Service) ExtractImagesFromManifest(manifestPath string) ([]string, error) {
	manifests, err := s.manifestParser.Parse(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse manifest at %q: %w", manifestPath, err)
	}

	imageSet := make(map[string]struct{})
	for _, manifest := range manifests {
		if err := s.extractImages(manifest, imageSet); err != nil {
			return nil, fmt.Errorf("failed to extract images from %q kind: %w", manifest.GetKind(), err)
		}
	}

	return setToSlice(imageSet), nil
}

func (s *Service) extractImages(manifest *unstructured.Unstructured, imageSet map[string]struct{}) error {
	kind := manifest.GetKind()
	if kind != KindDeploymnent && kind != KindStatefulSet {
		return nil
	}

	if err := s.extractFromContainers(manifest, imageSet, "spec", "template", "spec", "containers"); err != nil {
		return fmt.Errorf("failed to extract from containers: %w", err)
	}

	if err := s.extractFromContainers(manifest, imageSet, "spec", "template", "spec", "initContainers"); err != nil {
		return fmt.Errorf("failed to extract from initContainers: %w", err)
	}

	return nil
}

func (s *Service) extractFromContainers(manifest *unstructured.Unstructured, imageSet map[string]struct{}, path ...string) error {
	containers, found, _ := unstructured.NestedSlice(manifest.Object, path...)
	if !found {
		return nil
	}

	for _, container := range containers {
		containerMap, ok := container.(map[string]interface{})
		if !ok {
			continue
		}

		if image, found, _ := unstructured.NestedString(containerMap, "image"); found {
			valid, err := s.isValidImage(image)
			if err != nil {
				return fmt.Errorf("invalid image %q in %v: %w", image, path, err)
			}
			if valid {
				imageSet[image] = struct{}{}
			}
		}

		if err := s.extractFromEnv(containerMap, imageSet); err != nil {
			return fmt.Errorf("extracting env images failed: %w", err)
		}
	}

	return nil
}

func (s *Service) extractFromEnv(container map[string]interface{}, imageSet map[string]struct{}) error {
	envVars, found, _ := unstructured.NestedSlice(container, "env")
	if !found {
		return nil
	}

	for _, envVar := range envVars {
		envMap, ok := envVar.(map[string]interface{})
		if !ok {
			continue
		}

		if value, found, _ := unstructured.NestedString(envMap, "value"); found {
			valid, err := s.isValidImage(value)
			if err != nil {
				return fmt.Errorf("invalid image %q in env var: %w", value, err)
			}
			if valid {
				imageSet[value] = struct{}{}
			}
		}
	}

	return nil
}

func (s *Service) isValidImage(value string) (bool, error) {
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

func setToSlice(imageSet map[string]struct{}) []string {
	images := make([]string, 0, len(imageSet))
	for image := range imageSet {
		images = append(images, image)
	}
	return images
}
