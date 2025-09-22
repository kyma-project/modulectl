package resources

import (
	"errors"

	"ocm.software/ocm/api/ocm/compdesc"
	ocmv1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/extensions/artifacttypes"

	"github.com/kyma-project/modulectl/internal/common"
	"github.com/kyma-project/modulectl/internal/common/types"
	"github.com/kyma-project/modulectl/internal/service/componentdescriptor/resources/accesshandler"
)

var ErrNilTarGenerator = errors.New("tarGenerator must not be nil")

type Service struct {
	tarGenerator accesshandler.TarGenerator
}

func NewService(tarGen accesshandler.TarGenerator) (*Service, error) {
	if tarGen == nil {
		return nil, ErrNilTarGenerator
	}

	return &Service{
		tarGenerator: tarGen,
	}, nil
}

type AccessHandler interface {
	GenerateBlobAccess() (cpi.BlobAccess, error)
}

type Resource struct {
	compdesc.Resource

	AccessHandler AccessHandler
}

func (s *Service) GenerateModuleResources(resourcePaths *types.ResourcePaths, version string,
) []Resource {
	moduleImageResource := GenerateModuleImageResource()
	rawManifestResource := GenerateRawManifestResource(s.tarGenerator, resourcePaths.RawManifest)
	resources := []Resource{moduleImageResource, rawManifestResource}
	if resourcePaths.DefaultCR != "" {
		defaultCRResource := GenerateDefaultCRResource(s.tarGenerator, resourcePaths.DefaultCR)
		resources = append(resources, defaultCRResource)
	}

	for idx := range resources {
		resources[idx].Version = version
	}
	return resources
}

func GenerateModuleImageResource() Resource {
	return Resource{
		Resource: compdesc.Resource{
			ResourceMeta: compdesc.ResourceMeta{
				ElementMeta: compdesc.ElementMeta{
					Name: common.ModuleImageResourceName,
				},
				Type:     artifacttypes.OCI_ARTIFACT,
				Relation: ocmv1.ExternalRelation,
			},
		},
	}
}

func GenerateRawManifestResource(tarGen accesshandler.TarGenerator, manifestPath string) Resource {
	return Resource{
		Resource: compdesc.Resource{
			ResourceMeta: compdesc.ResourceMeta{
				ElementMeta: compdesc.ElementMeta{
					Name: common.RawManifestResourceName,
				},
				Type:     artifacttypes.DIRECTORY_TREE,
				Relation: ocmv1.LocalRelation,
			},
		},
		AccessHandler: accesshandler.NewTar(tarGen, manifestPath),
	}
}

func GenerateDefaultCRResource(tarGen accesshandler.TarGenerator, defaultCRPath string) Resource {
	return Resource{
		Resource: compdesc.Resource{
			ResourceMeta: compdesc.ResourceMeta{
				ElementMeta: compdesc.ElementMeta{
					Name: common.DefaultCRResourceName,
				},
				Type:     artifacttypes.DIRECTORY_TREE,
				Relation: ocmv1.LocalRelation,
			},
		},
		AccessHandler: accesshandler.NewTar(tarGen, defaultCRPath),
	}
}
