package componentdescriptor

import (
	"fmt"
	ocm "ocm.software/ocm/api/ocm/compdesc/versions/v2"
)

type Service struct {
	ocm.ComponentDescriptor
}

const (
	providerName = "kyma-project.io"
)

func NewService() *Service {
	return &Service{
		ComponentDescriptor: ocm.ComponentDescriptor{},
	}
}

func (s *Service) PopulateComponentDescriptorMetadata(moduleName string,
	moduleVersion string) {
	s.ComponentDescriptor.SetName(moduleName)
	s.ComponentDescriptor.SetVersion(moduleVersion)

	s.ComponentDescriptor.Provider = providerName

	ocm.DefaultResources(&s.ComponentDescriptor)
}

func (s *Service) ValidateComponentDescriptor() error {
	if err := s.Validate(); err != nil {
		return fmt.Errorf("failed to validate component descriptor: %w", err)
	}

	return nil
}
