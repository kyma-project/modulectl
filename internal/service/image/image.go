package image

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"ocm.software/ocm/api/ocm/compdesc"
	ocmv1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/ociartifact"
	ociartifacttypes "ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/types/ociartifact"
)

var (
	ErrParserNil   = errors.New("parser cannot be nil")
	ErrInvalidPath = errors.New("invalid path provided")
	ErrInvalidURL  = errors.New("invalid image URL")
)

const (
	KindDeployment    = "Deployment"
	KindStatefulSet   = "StatefulSet"
	TypeManifestImage = "manifest-image"

	secScanBaseLabelKey  = "scan.security.kyma-project.io"
	typeLabelKey         = "type"
	ocmVersion           = "v1"
	imageTagSlicesLength = 2
)

var (
	// Full image reference with registry
	fullImagePattern = regexp.MustCompile(`^[a-zA-Z0-9\-\.]+\.[a-zA-Z]{2,}(/[a-zA-Z0-9\-\.\_]+)+[:\@][a-zA-Z0-9\-\.\_]+$`)

	// Docker Hub short name (namespace/image:tag)
	dockerHubPattern = regexp.MustCompile(`^[a-zA-Z0-9\-\.\_]+/[a-zA-Z0-9\-\.\_]+[:\@][a-zA-Z0-9\-\.\_]+$`)

	// Simple image name with tag (image:tag)
	simpleImagePattern = regexp.MustCompile(`^[a-zA-Z0-9\-\.\_]+:[a-zA-Z0-9\-\.\_]+$`)
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
	return &Service{
		manifestParser: manifestParser,
	}, nil
}

func (s *Service) AddImagesToOcmDescriptor(descriptor *compdesc.ComponentDescriptor, images []string) error {
	for _, img := range images {
		if err := s.appendImageResource(descriptor, img); err != nil {
			return fmt.Errorf("failed to append image %s: %w", img, err)
		}
	}

	compdesc.DefaultResources(descriptor)
	if err := compdesc.Validate(descriptor); err != nil {
		return fmt.Errorf("failed to validate component descriptor: %w", err)
	}

	return nil
}

func (s *Service) ExtractImagesFromManifest(manifestPath string) ([]string, error) {
	manifests, err := s.manifestParse(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	return s.extractImages(manifests), nil
}

func (s *Service) manifestParse(path string) ([]*unstructured.Unstructured, error) {
	if path == "" {
		return nil, fmt.Errorf("path cannot be empty: %w", ErrInvalidPath)
	}

	manifest, err := s.manifestParser.Parse(path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse manifest at %s: %w", path, err)
	}

	return manifest, nil
}

func (s *Service) extractImages(manifests []*unstructured.Unstructured) []string {
	imageSet := make(map[string]struct{})

	for _, manifest := range manifests {
		if !s.isSupportedWorkload(manifest) {
			continue
		}

		s.extractImagesFromManifest(manifest, imageSet)
	}

	return s.setToSlice(imageSet)
}

func (s *Service) isSupportedWorkload(manifest *unstructured.Unstructured) bool {
	kind := manifest.GetKind()
	return kind == KindDeployment || kind == KindStatefulSet
}

func (s *Service) extractImagesFromManifest(manifest *unstructured.Unstructured, imageSet map[string]struct{}) {
	containers, found, _ := unstructured.NestedSlice(manifest.Object, "spec", "template", "spec", "containers")
	if found {
		s.scanContainersForImages(containers, imageSet)
	}

	// Extract from init containers as well
	initContainers, found, _ := unstructured.NestedSlice(manifest.Object, "spec", "template", "spec", "initContainers")
	if found {
		s.scanContainersForImages(initContainers, imageSet)
	}
}

func (s *Service) extractContainerImage(container map[string]interface{}, imageSet map[string]struct{}) {
	if image, found, _ := unstructured.NestedString(container, "image"); found && image != "" && s.isImageReference(image) {
		imageSet[image] = struct{}{}
	}
}

func (s *Service) scanContainersForImages(containers []interface{}, imageSet map[string]struct{}) {
	for _, container := range containers {
		containerMap, ok := container.(map[string]interface{})
		if !ok {
			continue
		}

		s.extractContainerImage(containerMap, imageSet)
		s.extractEnvironmentImages(containerMap, imageSet)
	}
}

func (s *Service) extractEnvironmentImages(container map[string]interface{}, imageSet map[string]struct{}) {
	envVars, found, _ := unstructured.NestedSlice(container, "env")
	if !found {
		return
	}

	for _, envVar := range envVars {
		envMap, ok := envVar.(map[string]interface{})
		if !ok {
			continue
		}

		if value, found, _ := unstructured.NestedString(envMap, "value"); found && s.isImageReference(value) {
			imageSet[value] = struct{}{}
		}
	}
}

func (s *Service) isImageReference(imageValue string) bool {
	// Early exit optimizations
	vLen := len(imageValue)
	if vLen < 3 || vLen > 256 {
		return false
	}

	var (
		hasColon    bool
		hasDot      bool
		hasSlash    bool
		dotCount    int
		firstDotPos = -1
	)

	for index, char := range imageValue {
		switch char {
		case ':':
			hasColon = true
		case '.':
			if !hasDot {
				firstDotPos = index
				hasDot = true
			}
			dotCount++
		case '/':
			hasSlash = true
		case ' ', '\t', '\n', '\r':
			return false // Early reject on invalid chars
		}
	}

	if !hasColon && !hasDot {
		return false
	}

	switch {
	case hasDot && hasSlash:
		if firstDotPos > 0 && firstDotPos < vLen-3 && dotCount >= 2 {
			return fullImagePattern.MatchString(imageValue)
		}
		return dockerHubPattern.MatchString(imageValue)

	case hasSlash && hasColon:
		return dockerHubPattern.MatchString(imageValue)

	case hasDot && !hasSlash:
		return simpleImagePattern.MatchString(imageValue)

	case hasColon && !hasSlash:
		return simpleImagePattern.MatchString(imageValue)
	}

	return false
}

func (s *Service) setToSlice(imageSet map[string]struct{}) []string {
	images := make([]string, 0, len(imageSet))
	for image := range imageSet {
		images = append(images, image)
	}
	return images
}

func (s *Service) appendImageResource(descriptor *compdesc.ComponentDescriptor, img string) error {
	imgName, imgTag, err := GetImageNameAndTag(img)
	if err != nil {
		return fmt.Errorf("failed to get image name and tag: %w", err)
	}

	typeLabel, err := s.createImageTypeLabel()
	if err != nil {
		return fmt.Errorf("failed to create image type label: %w", err)
	}

	access := ociartifact.New(img)
	access.SetType(ociartifact.Type)

	resource := compdesc.Resource{
		ResourceMeta: compdesc.ResourceMeta{
			Type:     ociartifacttypes.TYPE,
			Relation: ocmv1.ExternalRelation,
			ElementMeta: compdesc.ElementMeta{
				Name:    imgName,
				Labels:  []ocmv1.Label{*typeLabel},
				Version: imgTag,
			},
		},
		Access: access,
	}

	descriptor.Resources = append(descriptor.Resources, resource)
	return nil
}

func (s *Service) createImageTypeLabel() (*ocmv1.Label, error) {
	imageTypeLabelKey := fmt.Sprintf("%s/%s", secScanBaseLabelKey, typeLabelKey)
	label, err := ocmv1.NewLabel(imageTypeLabelKey, TypeManifestImage, ocmv1.WithVersion(ocmVersion))
	if err != nil {
		return nil, fmt.Errorf("failed to create OCM label: %w", err)
	}
	return label, nil
}

func GetImageNameAndTag(imageURL string) (string, string, error) {
	// Find last colon to separate tag
	lastColonIndex := strings.LastIndex(imageURL, ":")
	if lastColonIndex == -1 {
		return "", "", fmt.Errorf("image URL: %s: %w", imageURL, ErrInvalidURL)
	}

	imagePart := imageURL[:lastColonIndex]
	tag := imageURL[lastColonIndex+1:]

	// Extract image name from path
	imageName := imagePart
	if lastSlash := strings.LastIndex(imagePart, "/"); lastSlash != -1 {
		imageName = imagePart[lastSlash+1:]
	}

	return imageName, tag, nil
}
