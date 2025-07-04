package resources

import (
	"errors"
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/cli-runtime/pkg/resource"
	"ocm.software/ocm/api/ocm/compdesc"
	ocmv1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/extensions/artifacttypes"

	"github.com/kyma-project/modulectl/internal/service/componentdescriptor/resources/accesshandler"
	"github.com/kyma-project/modulectl/internal/service/contentprovider"
	"github.com/kyma-project/modulectl/internal/service/verifier"
)

const (
	moduleImageResourceName = "module-image"
	rawManifestResourceName = "raw-manifest"
	defaultCRResourceName   = "default-cr"
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

func (s *Service) GenerateModuleResources(moduleConfig *contentprovider.ModuleConfig,
	manifestPath, defaultCRPath string) ([]Resource, error) {
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

	for idx := range resources {
		resources[idx].Version = moduleConfig.Version
	}
	return resources, nil
}

func (s *Service) VerifyModuleResources(moduleConfig *contentprovider.ModuleConfig, manifestPath string) error {
	resources, err := parseRawManifest(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to parse raw manifest: %w", err)
	}
	imageVersionVerifier := verifier.NewImageVersionVerifier(moduleConfig.Version, moduleConfig.Manager.Name)
	if err := imageVersionVerifier.VerifyModuleImageVersion(resources); err != nil {
		return fmt.Errorf("failed to verify module image version: %w", err)
	}
	return nil
}

func parseRawManifest(path string) ([]*unstructured.Unstructured, error) {
	var objects []*unstructured.Unstructured
	builder := resource.NewLocalBuilder().
		Unstructured().
		Path(false, path).
		Flatten().
		ContinueOnError()

	result := builder.Do()

	if err := result.Err(); err != nil {
		return nil, fmt.Errorf("parse manifest: %w", err)
	}
	items, err := result.Infos()
	if err != nil {
		return nil, fmt.Errorf("parse manifest to resource infos: %w", err)
	}
	for _, item := range items {
		unstructuredItem, ok := item.Object.(*unstructured.Unstructured)
		if !ok {
			continue
		}

		objects = append(objects, unstructuredItem)
	}
	return objects, nil
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
