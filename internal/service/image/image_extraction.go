package image

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	LatestTag = "latest"
	MainTag   = "main"
)

func (s *Service) ExtractImagesFromManifest(manifestPath string) ([]string, error) {
	manifests, err := s.manifestParser.Parse(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	imageSet := make(map[string]struct{})
	for _, manifest := range manifests {
		s.extractImages(manifest, imageSet)
	}

	return setToSlice(imageSet), nil
}

func (s *Service) extractImages(manifest *unstructured.Unstructured, imageSet map[string]struct{}) {
	kind := manifest.GetKind()
	if kind != "Deployment" && kind != "StatefulSet" {
		return
	}

	s.extractFromContainers(manifest, imageSet, "spec", "template", "spec", "containers")
	s.extractFromContainers(manifest, imageSet, "spec", "template", "spec", "initContainers")
}

func (s *Service) extractFromContainers(manifest *unstructured.Unstructured, imageSet map[string]struct{}, path ...string) {
	containers, found, _ := unstructured.NestedSlice(manifest.Object, path...)
	if !found {
		return
	}

	for _, container := range containers {
		containerMap, ok := container.(map[string]interface{})
		if !ok {
			continue
		}

		// Extract from containers.image
		if image, found, _ := unstructured.NestedString(containerMap, "image"); found && s.isValidImage(image) {
			imageSet[image] = struct{}{}
		}

		// Extract from containers.env[].value
		s.extractFromEnv(containerMap, imageSet)
	}
}

func (s *Service) extractFromEnv(container map[string]interface{}, imageSet map[string]struct{}) {
	envVars, found, _ := unstructured.NestedSlice(container, "env")
	if !found {
		return
	}

	for _, envVar := range envVars {
		envMap, ok := envVar.(map[string]interface{})
		if !ok {
			continue
		}

		if value, found, _ := unstructured.NestedString(envMap, "value"); found && s.isValidImage(value) {
			imageSet[value] = struct{}{}
		}
	}
}

// isValidImage, only skip latest/main as are not allowed for prod modules manifest
func (s *Service) isValidImage(value string) bool {
	if !isValidImageFormat(value) {
		return false
	}

	// Parse image to get tag
	_, tag, err := ParseImageReference(value)
	if err != nil {
		return false
	}

	if tag == "" {
		return false
	}

	// Only skip latest and main tags
	return !isMainOrLatestTag(tag)
}

func isValidImageFormat(value string) bool {
	if value == "" || len(value) < 3 || len(value) > 256 {
		return false
	}

	hasTagOrDigest := false
	for _, char := range value {
		if char == ':' || char == '@' {
			hasTagOrDigest = true
			break
		}
		if char == ' ' || char == '\t' || char == '\n' || char == '\r' {
			return false
		}
	}

	return hasTagOrDigest
}

// isMainOrLatestTag checks if tag should be skipped - only latest and main
func isMainOrLatestTag(tag string) bool {
	if tag == "" {
		return true
	}

	lowercaseTag := strings.ToLower(tag)
	return lowercaseTag == LatestTag || lowercaseTag == MainTag
}

func setToSlice(imageSet map[string]struct{}) []string {
	images := make([]string, 0, len(imageSet))
	for image := range imageSet {
		images = append(images, image)
	}
	return images
}
