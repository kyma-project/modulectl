package componentdescriptor

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"ocm.software/ocm/api/ocm/compdesc"
	ocmv1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/ociartifact"
	ociartifacttypes "ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/types/ociartifact"

	"github.com/kyma-project/modulectl/internal/service/image"
)

const (
	// Semantic versioning format following e.g: x.y.z or vx.y,z
	semverPattern = `^v?(\d+)\.(\d+)\.(\d+)(?:-([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?(?:\+([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?$`
)

var ErrInvalidImageFormat = errors.New("invalid image format")

func AddOciArtifactsToDescriptor(descriptor *compdesc.ComponentDescriptor, images []string) error {
	for _, img := range images {
		valid, err := image.IsValidImage(img)
		if err != nil {
			return fmt.Errorf("image validation failed for %s: %w", img, err)
		}
		if !valid {
			return fmt.Errorf("%w: %s", ErrInvalidImageFormat, img)
		}

		if err := appendOciArtifactResource(descriptor, img); err != nil {
			return fmt.Errorf("failed to append image %s: %w", img, err)
		}
	}
	return nil
}

func appendOciArtifactResource(descriptor *compdesc.ComponentDescriptor, imageURL string) error {
	imageInfo, err := image.ParseImageInfo(imageURL)
	if err != nil {
		return fmt.Errorf("failed to parse image: %w", err)
	}

	typeLabel, err := CreateImageTypeLabel()
	if err != nil {
		return fmt.Errorf("failed to create label: %w", err)
	}

	// Generate OCM-compatible version and resource name
	version, resourceName := generateOCMVersionAndName(imageInfo)

	if resourceExists(descriptor, resourceName, version) {
		return nil // Skip duplicate resource
	}

	access := ociartifact.New(imageURL)
	access.SetType(ociartifact.Type)

	resource := compdesc.Resource{
		ResourceMeta: compdesc.ResourceMeta{
			Type:     ociartifacttypes.TYPE,
			Relation: ocmv1.ExternalRelation,
			ElementMeta: compdesc.ElementMeta{
				Name:    resourceName,
				Labels:  []ocmv1.Label{*typeLabel},
				Version: version,
			},
		},
		Access: access,
	}

	descriptor.Resources = append(descriptor.Resources, resource)
	compdesc.DefaultResources(descriptor)

	if err = compdesc.Validate(descriptor); err != nil {
		return fmt.Errorf("failed to validate component descriptor: %w", err)
	}

	return nil
}

func CreateImageTypeLabel() (*ocmv1.Label, error) {
	labelKey := fmt.Sprintf("%s/%s", secScanBaseLabelKey, typeLabelKey)
	label, err := ocmv1.NewLabel(labelKey, thirdPartyImageLabelValue, ocmv1.WithVersion(ocmVersion))
	if err != nil {
		return nil, fmt.Errorf("failed to create OCM label: %w", err)
	}
	return label, nil
}

func resourceExists(descriptor *compdesc.ComponentDescriptor, name, version string) bool {
	for _, resource := range descriptor.Resources {
		if resource.Name == name && resource.Version == version {
			return true
		}
	}
	return false
}

func generateOCMVersionAndName(info *image.ImageInfo) (string, string) {
	if info.Digest != "" {
		shortDigest := info.Digest[:12]
		var version string
		switch {
		case info.Tag != "" && isValidSemverForOCM(info.Tag):
			version = fmt.Sprintf("%s+sha256.%s", info.Tag, shortDigest)
		case info.Tag != "":
			version = fmt.Sprintf("0.0.0-%s+sha256.%s", normalizeTagForOCM(info.Tag), shortDigest)
		default:
			version = "0.0.0+sha256." + shortDigest
		}
		resourceName := fmt.Sprintf("%s-%s", info.Name, info.Digest[:8])
		return version, resourceName
	}

	var version string
	if isValidSemverForOCM(info.Tag) {
		version = info.Tag
	} else {
		version = "0.0.0-" + normalizeTagForOCM(info.Tag)
	}

	resourceName := info.Name
	return version, resourceName
}

func normalizeTagForOCM(tag string) string {
	reg := regexp.MustCompile(`[^a-zA-Z0-9.-]`)
	normalized := reg.ReplaceAllString(tag, "-")
	normalized = strings.Trim(normalized, "-.")
	if normalized == "" {
		normalized = "unknown"
	}
	return normalized
}

func isValidSemverForOCM(version string) bool {
	matched, _ := regexp.MatchString(semverPattern, version)
	return matched
}
