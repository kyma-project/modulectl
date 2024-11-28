package componentarchive

import (
	"fmt"

	"github.com/mandelsoft/vfs/pkg/vfs"
	"ocm.software/ocm/api/ocm/compdesc"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/extensions/repositories/comparch"
	"ocm.software/ocm/api/utils/accessobj"
	"sigs.k8s.io/yaml"

	commonerrors "github.com/kyma-project/modulectl/internal/common/errors"
	"github.com/kyma-project/modulectl/internal/service/componentdescriptor"
)

const (
	componentDescriptorPath     = "./mod"
	componentDescriptorFileName = "component-descriptor.yaml"
)

type ArchiveFileSystem interface {
	CreateArchiveFileSystem(path string) error
	WriteFile(data []byte, fileName string) error
	GetArchiveFileSystem() vfs.FileSystem
	GenerateTarFileSystemAccess(filePath string) (cpi.BlobAccess, error)
}

type Service struct {
	fileSystem ArchiveFileSystem
}

func NewService(fileSystem ArchiveFileSystem) (*Service, error) {
	if fileSystem == nil {
		return nil, fmt.Errorf("fileSystem must not be nil: %w", commonerrors.ErrInvalidArg)
	}

	return &Service{
		fileSystem: fileSystem,
	}, nil
}

func (s *Service) CreateComponentArchive(descriptor *compdesc.ComponentDescriptor) (*comparch.ComponentArchive, error) {
	if err := s.fileSystem.CreateArchiveFileSystem(componentDescriptorPath); err != nil {
		return nil, fmt.Errorf("failed to create archive file system, %w", err)
	}

	encodeOptions := &compdesc.EncodeOptions{
		SchemaVersion: descriptor.SchemaVersion(),
	}
	versionedDescriptor, err := compdesc.Convert(descriptor, encodeOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to convert component descriptor, %w", err)
	}

	descriptorData, err := yaml.Marshal(versionedDescriptor)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal component descriptor, %w", err)
	}

	if err := s.fileSystem.WriteFile(descriptorData, componentDescriptorFileName); err != nil {
		return nil, fmt.Errorf("failed to write to component descriptor file, %w", err)
	}

	componentArchive, err := comparch.New(cpi.DefaultContext(),
		accessobj.ACC_CREATE, s.fileSystem.GetArchiveFileSystem(), nil, nil, vfs.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("failed to create component archive, %w", err)
	}
	return componentArchive, nil
}

type ComponentArchive interface {
	AddBlob(blob cpi.BlobAccess, artType string, refName string, global cpi.AccessSpec, opts ...cpi.BlobUploadOption) (cpi.AccessSpec, error)
	SetResource(meta *cpi.ResourceMeta, access compdesc.AccessSpec, opts ...cpi.ModificationOption) error
	Close() error
}

func (s *Service) AddModuleResourcesToArchive(archive ComponentArchive, moduleResources []componentdescriptor.Resource) error {
	for _, resource := range moduleResources {
		if resource.Path != "" {
			access, err := s.fileSystem.GenerateTarFileSystemAccess(resource.Path)
			if err != nil {
				return fmt.Errorf("failed to generate tar file access, %w", err)
			}

			blobAccess, err := archive.AddBlob(access, access.MimeType(), resource.Name, nil)
			if err != nil {
				return fmt.Errorf("failed to add blob, %w", err)
			}

			if err = archive.SetResource(&resource.ResourceMeta, blobAccess,
				cpi.ModifyResource(true)); err != nil {
				return fmt.Errorf("failed to set resource, %w", err)
			}
		}
	}
	return nil
}
