package componentdescriptor

import (
	"fmt"

	"github.com/kyma-project/modulectl/internal/common"
	"github.com/kyma-project/modulectl/internal/service/componentdescriptor/resources"
	"github.com/kyma-project/modulectl/internal/service/image"
	ocmv1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"

	"ocm.software/ocm/api/ocm/compdesc"
)

func InitializeComponentDescriptor(moduleName string, moduleVersion string) (*compdesc.ComponentDescriptor, error) {
	componentDescriptor := &compdesc.ComponentDescriptor{}
	componentDescriptor.SetName(moduleName)
	componentDescriptor.SetVersion(moduleVersion)
	componentDescriptor.Metadata.ConfiguredVersion = common.VersionV2

	providerLabel, err := ocmv1.NewLabel(common.BuiltByLabelKey, common.BuiltByLabelValue,
		ocmv1.WithVersion(common.VersionV1))
	if err != nil {
		return nil, fmt.Errorf("failed to create label: %w", err)
	}

	componentDescriptor.Provider = ocmv1.Provider{Name: common.ProviderName, Labels: ocmv1.Labels{*providerLabel}}

	compdesc.DefaultResources(componentDescriptor)

	if err = compdesc.Validate(componentDescriptor); err != nil {
		return nil, fmt.Errorf("failed to validate component descriptor: %w", err)
	}

	return componentDescriptor, nil
}

func AddOciArtifactsToDescriptor(descriptor *compdesc.ComponentDescriptor, images []string) error {
	for _, img := range images {
		imageInfo, err := image.ValidateAndParseImageInfo(img)
		if err != nil {
			return fmt.Errorf("image validation failed for %s: %w", img, err)
		}

		resource, err := resources.NewOciArtifactResource(imageInfo)
		if err != nil {
			return fmt.Errorf("failed to create resource for %s: %w", img, err)
		}

		resources.AddResourceIfNotExists(descriptor, resource)
	}
	err := compdesc.Validate(descriptor)
	if err != nil {
		return fmt.Errorf("failed to validate component descriptor: %w", err)
	}

	return nil
}
