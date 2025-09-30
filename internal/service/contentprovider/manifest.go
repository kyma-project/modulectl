package contentprovider

import (
	"errors"
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/kyma-project/modulectl/internal/common/types"
	"github.com/kyma-project/modulectl/internal/common/utils/slices"
	"github.com/kyma-project/modulectl/internal/service/image"
)

const (
	KindDeployment  = "Deployment"
	KindStatefulSet = "StatefulSet"
)

var ErrParserNil = errors.New("parser cannot be nil")

type Manifest struct {
	manifestParser types.RawManifestParser
}

func NewManifest(manifestParser types.RawManifestParser) (*Manifest, error) {
	if manifestParser == nil {
		return nil, fmt.Errorf("manifestParser must not be nil: %w", ErrParserNil)
	}
	return &Manifest{manifestParser: manifestParser}, nil
}

func (m *Manifest) GetDefaultContent(_ types.KeyValueArgs) (string, error) {
	return `# This file holds the Manifest of your module, encompassing all resources ` +
		`installed in the cluster once the module is activated.
# It should include the Custom Resource Definition for your module's default CustomResource, if it exists.

`, nil
}

func (m *Manifest) ExtractImagesFromManifest(manifestPath string) ([]string, error) {
	manifests, err := m.manifestParser.Parse(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse manifest at %q: %w", manifestPath, err)
	}

	imageSet := make(map[string]struct{})
	for _, manifest := range manifests {
		if err = extractImages(manifest, imageSet); err != nil {
			return nil, fmt.Errorf("failed to extract images from %q kind: %w", manifest.GetKind(), err)
		}
	}

	return slices.SetToSlice(imageSet), nil
}

func extractImages(manifest *unstructured.Unstructured, imageSet map[string]struct{}) error {
	kind := manifest.GetKind()
	if kind != KindDeployment && kind != KindStatefulSet {
		return nil
	}

	if err := extractFromContainers(manifest, imageSet, "spec", "template", "spec", "containers"); err != nil {
		return fmt.Errorf("failed to extract from containers: %w", err)
	}

	if err := extractFromContainers(manifest, imageSet, "spec", "template", "spec", "initContainers"); err != nil {
		return fmt.Errorf("failed to extract from initContainers: %w", err)
	}

	return nil
}

func extractFromContainers(manifest *unstructured.Unstructured, imageSet map[string]struct{}, path ...string) error {
	containers, found, _ := unstructured.NestedSlice(manifest.Object, path...)
	if !found {
		return nil
	}

	for _, container := range containers {
		containerMap, ok := container.(map[string]interface{})
		if !ok {
			continue
		}

		if img, found, _ := unstructured.NestedString(containerMap, "image"); found {
			if image.IsImageReferenceCandidate(img) {
				if _, err := image.ValidateAndParseImageInfo(img); err != nil {
					return fmt.Errorf("invalid img %q in %v: %w", img, path, err)
				}
				imageSet[img] = struct{}{}
			}
		}

		if err := extractFromEnv(containerMap, imageSet); err != nil {
			return fmt.Errorf("extracting env images failed: %w", err)
		}
	}

	return nil
}

func extractFromEnv(container map[string]interface{}, imageSet map[string]struct{}) error {
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
			if image.IsImageReferenceCandidate(value) {
				if _, err := image.ValidateAndParseImageInfo(value); err != nil {
					return fmt.Errorf("invalid image %q in env var: %w", value, err)
				}
				imageSet[value] = struct{}{}
			}
		}
	}

	return nil
}
