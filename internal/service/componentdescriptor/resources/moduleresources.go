package resources

import (
	"errors"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/json"
	"ocm.software/ocm/api/ocm/compdesc"
	ocmv1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/extensions/artifacttypes"

	"github.com/kyma-project/modulectl/internal/service/componentdescriptor/resources/accesshandler"
	"github.com/kyma-project/modulectl/internal/service/contentprovider"
)

const (
	moduleImageResourceName = "module-image"
	rawManifestResourceName = "raw-manifest"
	defaultCRResourceName   = "default-cr"
	ociRegistryCredLabel    = "oci-registry-cred" //nolint:gosec // it's a label
)

var ErrNilArchiveFileSystem = errors.New("archiveFileSystem must not be nil")

type Service struct {
	archiveFileSystem accesshandler.ArchiveFileSystem
}

func NewService(archiveFileSystem accesshandler.ArchiveFileSystem) (*Service, error) {
	if archiveFileSystem == nil {
		return nil, ErrNilArchiveFileSystem
	}

	return &Service{
		archiveFileSystem: archiveFileSystem,
	}, nil
}

type AccessHandler interface {
	GenerateBlobAccess() (cpi.BlobAccess, error)
}

type Resource struct {
	compdesc.Resource
	AccessHandler AccessHandler
}

func (s *Service) GenerateModuleResources(moduleConfig *contentprovider.ModuleConfig, manifestPath, defaultCRPath,
	registryCredSelector string,
) ([]Resource, error) {
	moduleImageResource := GenerateModuleImageResource()
	metadataResource, err := GenerateMetadataResource(moduleConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to generate metadata resource: %w", err)
	}
	rawManifestResource := GenerateRawManifestResource(s.archiveFileSystem, manifestPath)
	resources := []Resource{moduleImageResource, metadataResource, rawManifestResource}
	if defaultCRPath != "" {
		defaultCRResource := GenerateDefaultCRResource(s.archiveFileSystem, defaultCRPath)
		resources = append(resources, defaultCRResource)
	}

	credentialsLabel, err := CreateCredMatchLabels(registryCredSelector)
	if err != nil {
		return nil, fmt.Errorf("failed to create credentials label: %w", err)
	}

	for idx := range resources {
		resources[idx].Version = moduleConfig.Version
		if len(credentialsLabel) > 0 {
			resources[idx].SetLabels([]ocmv1.Label{
				{
					Name:  ociRegistryCredLabel,
					Value: credentialsLabel,
				},
			})
		}
	}
	return resources, nil
}

func GenerateModuleImageResource() Resource {
	return Resource{
		Resource: compdesc.Resource{
			ResourceMeta: compdesc.ResourceMeta{
				ElementMeta: compdesc.ElementMeta{
					Name: moduleImageResourceName,
				},
				Type:     artifacttypes.OCI_ARTIFACT,
				Relation: ocmv1.ExternalRelation,
			},
		},
	}
}

func GenerateRawManifestResource(archiveFileSystem accesshandler.ArchiveFileSystem, manifestPath string) Resource {
	return Resource{
		Resource: compdesc.Resource{
			ResourceMeta: compdesc.ResourceMeta{
				ElementMeta: compdesc.ElementMeta{
					Name: rawManifestResourceName,
				},
				Type:     artifacttypes.DIRECTORY_TREE,
				Relation: ocmv1.LocalRelation,
			},
		},
		AccessHandler: accesshandler.NewTar(archiveFileSystem, manifestPath),
	}
}

func GenerateDefaultCRResource(archiveFileSystem accesshandler.ArchiveFileSystem, defaultCRPath string) Resource {
	return Resource{
		Resource: compdesc.Resource{
			ResourceMeta: compdesc.ResourceMeta{
				ElementMeta: compdesc.ElementMeta{
					Name: defaultCRResourceName,
				},
				Type:     artifacttypes.DIRECTORY_TREE,
				Relation: ocmv1.LocalRelation,
			},
		},
		AccessHandler: accesshandler.NewTar(archiveFileSystem, defaultCRPath),
	}
}

func CreateCredMatchLabels(registryCredSelector string) ([]byte, error) {
	var matchLabels []byte
	if registryCredSelector != "" {
		selector, err := metav1.ParseToLabelSelector(registryCredSelector)
		if err != nil {
			return nil, fmt.Errorf("failed to parse label selector: %w", err)
		}
		matchLabels, err = json.Marshal(selector.MatchLabels)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal match labels: %w", err)
		}
	}
	return matchLabels, nil
}
