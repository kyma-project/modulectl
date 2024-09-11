package componentdescriptor

import (
	"fmt"
	ocm "ocm.software/ocm/api/ocm/compdesc/versions/v2"
)

type FileSystem interface {
	FileExists(path string) (bool, error)
}

type SecurityConfig struct {
	fileSystem FileSystem
	Sources    ocm.Sources
}

func (s *SecurityConfig) AddSecurityConfigData(securityConfigFile string) error {
	if exists, err := s.fileSystem.FileExists(securityConfigFile); err != nil {
		return fmt.Errorf("failed to check if security config file exists: %w", err)
	} else if !exists {
		return nil
	}

	return nil
}
