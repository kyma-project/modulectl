package componentdescriptor

import (
	"fmt"

	ocm "ocm.software/ocm/api/ocm/compdesc"
	ocmv1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
)

const (
	providerName  = "kyma-project.io"
	schemaVersion = "v2"
)

func InitializeComponentDescriptor(moduleName string,
	moduleVersion string,
) (*ocm.ComponentDescriptor, error) {
	componentDescriptor := &ocm.ComponentDescriptor{}
	componentDescriptor.SetName(moduleName)
	componentDescriptor.SetVersion(moduleVersion)
	componentDescriptor.Metadata.ConfiguredVersion = schemaVersion

	builtByModulectl, err := ocmv1.NewLabel("kyma-project.io/built-by", "modulectl", ocmv1.WithVersion("v1"))
	if err != nil {
		return nil, fmt.Errorf("failed to create label: %w", err)
	}
	componentDescriptor.Provider = ocmv1.Provider{Name: providerName, Labels: ocmv1.Labels{*builtByModulectl}}

	ocm.DefaultResources(componentDescriptor)
	if err := ocm.Validate(componentDescriptor); err != nil {
		return nil, fmt.Errorf("failed to validate component descriptor: %w", err)
	}

	return componentDescriptor, nil
}
