package image

import (
	"fmt"

	"ocm.software/ocm/api/ocm/compdesc"
	ocmv1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/ociartifact"
	ociartifacttypes "ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/types/ociartifact"
)

const (
	TypeManifestImage   = "third-party-image"
	secScanBaseLabelKey = "scan.security.kyma-project.io"
	typeLabelKey        = "type"
	ocmVersion          = "v1"
)

// AddImagesToOcmDescriptor - as per ticket requirements
func (s *Service) AddImagesToOcmDescriptor(descriptor *compdesc.ComponentDescriptor, images []string) error {
	for _, image := range images {
		if err := s.appendImageResource(descriptor, image); err != nil {
			return fmt.Errorf("failed to append image %s: %w", image, err)
		}
	}
	return nil
}

func (s *Service) appendImageResource(descriptor *compdesc.ComponentDescriptor, imageURL string) error {
	imgName, imgTag, err := ParseImageReference(imageURL)
	if err != nil {
		return fmt.Errorf("failed to parse image: %w", err)
	}

	typeLabel, err := createImageTypeLabel()
	if err != nil {
		return fmt.Errorf("failed to create label: %w", err)
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
				Version: imgTag,
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

func createImageTypeLabel() (*ocmv1.Label, error) {
	labelKey := fmt.Sprintf("%s/%s", secScanBaseLabelKey, typeLabelKey)
	label, err := ocmv1.NewLabel(labelKey, TypeManifestImage, ocmv1.WithVersion(ocmVersion))
	if err != nil {
		return nil, fmt.Errorf("failed to create OCM label: %w", err)
	}
	return label, nil
}
