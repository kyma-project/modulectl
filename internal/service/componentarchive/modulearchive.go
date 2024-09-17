package componentarchive

import (
	"fmt"

	"github.com/mandelsoft/vfs/pkg/vfs"
	"gopkg.in/yaml.v3"
	"ocm.software/ocm/api/ocm/compdesc"
	ocm "ocm.software/ocm/api/ocm/compdesc"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/extensions/repositories/comparch"
	"ocm.software/ocm/api/utils/accessobj"
)

const componentDescriptorFileName = "component-descriptor.yaml"

type ArchiveFileSystem interface {
	CreateArchiveFileSystem(path string) error
	WriteFile(data []byte, fileName string) error
	GetArchiveFileSystem() vfs.FileSystem
}

type Service struct {
	fileSystem ArchiveFileSystem
}

func NewService(fileSystem ArchiveFileSystem) *Service {
	return &Service{
		fileSystem: fileSystem,
	}
}

func (s *Service) CreateComponentArchive(path string,
	componentDescriptor *ocm.ComponentDescriptor) (*comparch.ComponentArchive,
	error) {
	if err := s.fileSystem.CreateArchiveFileSystem(path); err != nil {
		return nil, fmt.Errorf("failed to create archive file system, %w", err)
	}

	encodeOptions := &compdesc.EncodeOptions{
		SchemaVersion: componentDescriptor.SchemaVersion(),
	}
	versionedDescriptor, err := compdesc.Convert(componentDescriptor, encodeOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to convert component descriptor, %w", err)
	}

	descriptorData, err := yaml.Marshal(versionedDescriptor)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal component descriptor")
	}

	if err := s.fileSystem.WriteFile(descriptorData, componentDescriptorFileName); err != nil {
		return nil, fmt.Errorf("failed to write to component descriptor file, %w", err)
	}

	componentArchive, err := comparch.New(cpi.DefaultContext(),
		accessobj.ACC_CREATE, s.fileSystem.GetArchiveFileSystem(), nil, nil, vfs.ModePerm)

	if err != nil {
		return nil, fmt.Errorf("failed to created component archive, %w", err)
	}

	return componentArchive, nil
}
