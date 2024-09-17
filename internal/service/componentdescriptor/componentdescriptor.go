package componentdescriptor

import (
	"fmt"

	ocm "ocm.software/ocm/api/ocm/compdesc/versions/v2"
)

const (
	providerName = "kyma-project.io"
)

func InitializeComponentDescriptor(moduleName string,
	moduleVersion string,
) (*ocm.ComponentDescriptor, error) {
	componentDescriptor := &ocm.ComponentDescriptor{}
	componentDescriptor.SetName(moduleName)
	componentDescriptor.SetVersion(moduleVersion)
	componentDescriptor.Metadata.Version = ocm.SchemaVersion

	componentDescriptor.Provider = providerName

	ocm.DefaultResources(componentDescriptor)
	if err := componentDescriptor.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate component descriptor: %w", err)
	}

	return componentDescriptor, nil
}
