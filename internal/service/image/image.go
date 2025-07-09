package image

import (
	"fmt"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"ocm.software/ocm/api/ocm/compdesc"
	ocmv1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/ociartifact"
	ociartifacttypes "ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/types/ociartifact"
	"regexp"
	"strings"

	"errors"
)

var errInvalidURL = errors.New("invalid image URL")

const (
	deploymentKind    = "Deployment"
	statefulSetKind   = "StatefulSet"
	moduleImageType   = "module-image"
	manifestImageType = "manifest-image"

	secScanBaseLabelKey  = "scan.security.kyma-project.io"
	typeLabelKey         = "type"
	ocmVersion           = "v1"
	imageTagSlicesLength = 2
)

var imageReferencePattern = regexp.MustCompile(`^[a-zA-Z0-9\-\.]+(/[a-zA-Z0-9\-\.\_]+)+[:\@][a-zA-Z0-9\-\.\_]+$`)

type ManifestParser interface {
	Parse(path string) ([]*unstructured.Unstructured, error)
}

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) ManifestParse(path string, parser ManifestParser) ([]*unstructured.Unstructured, error) {
	if parser == nil {
		return nil, fmt.Errorf("parser cannot be nil")
	}
	return parser.Parse(path)
}

func (s *Service) GetAllImages(manifests []*unstructured.Unstructured) []string {
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
	return kind == deploymentKind || kind == statefulSetKind
}

func (s *Service) extractImagesFromManifest(manifest *unstructured.Unstructured, imageSet map[string]struct{}) {
	containers, found, _ := unstructured.NestedSlice(manifest.Object, "spec", "template", "spec", "containers")
	if !found {
		return
	}

	for _, container := range containers {
		containerMap, ok := container.(map[string]interface{})
		if !ok {
			continue
		}

		s.extractContainerImage(containerMap, imageSet)
		s.extractEnvironmentImages(containerMap, imageSet)
	}
}

func (s *Service) extractContainerImage(container map[string]interface{}, imageSet map[string]struct{}) {
	if image, found, _ := unstructured.NestedString(container, "image"); found && image != "" {
		imageSet[image] = struct{}{}
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

func (s *Service) isImageReference(value string) bool {
	return imageReferencePattern.MatchString(value)
}

func (s *Service) setToSlice(imageSet map[string]struct{}) []string {
	images := make([]string, 0, len(imageSet))
	for image := range imageSet {
		images = append(images, image)
	}
	return images
}

// AppendManifestImages follows the same pattern as AppendBDBAImagesLayers
func (s *Service) AppendManifestImages(descriptor *compdesc.ComponentDescriptor, images []string) error {
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
	return ocmv1.NewLabel(imageTypeLabelKey, manifestImageType, ocmv1.WithVersion(ocmVersion))
}

func GetImageNameAndTag(imageURL string) (string, string, error) {
	imageTag := strings.Split(imageURL, ":")
	if len(imageTag) != imageTagSlicesLength {
		return "", "", fmt.Errorf("image URL: %s: %w", imageURL, errInvalidURL)
	}

	imageName := strings.Split(imageTag[0], "/")
	if len(imageName) == 0 {
		return "", "", fmt.Errorf("image URL: %s: %w", imageURL, errInvalidURL)
	}

	return imageName[len(imageName)-1], imageTag[len(imageTag)-1], nil
}
