package componentdescriptor

import (
	"fmt"
	"strings"

	"ocm.software/ocm/api/ocm/compdesc"
	ocmv1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/ociartifact"
	ociartifacttypes "ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/types/ociartifact"

	"github.com/kyma-project/modulectl/internal/image"
)

func AddOciArtifactsToDescriptor(descriptor *compdesc.ComponentDescriptor, images []string) error {
	for _, image := range images {
		if err := AppendImageResource(descriptor, image); err != nil {
			return fmt.Errorf("failed to append image %s: %w", image, err)
		}
	}
	return nil
}

func AppendImageResource(descriptor *compdesc.ComponentDescriptor, imageURL string) error {
	imgName, imgTag, err := image.ParseImageReference(imageURL)
	if err != nil {
		return fmt.Errorf("failed to parse image: %w", err)
	}

	typeLabel, err := CreateImageTypeLabel()
	if err != nil {
		return fmt.Errorf("failed to create label: %w", err)
	}

	var version string
	if strings.HasPrefix(imgTag, "sha256:") {
		digest := strings.TrimPrefix(imgTag, "sha256:")
		version = "0.0.0+sha256." + digest
	} else {
		version = imgTag
	}
	access := ociartifact.New(imageURL)
	access.SetType(ociartifact.Type)

	resource := compdesc.Resource{
		ResourceMeta: compdesc.ResourceMeta{
			Type:     ociartifacttypes.TYPE,
			Relation: ocmv1.ExternalRelation,
			ElementMeta: compdesc.ElementMeta{
				Name:    imgName,
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
