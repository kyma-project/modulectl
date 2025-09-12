package componentconstructor

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/kyma-project/modulectl/internal/common"
	"github.com/kyma-project/modulectl/internal/common/types"
	"github.com/kyma-project/modulectl/internal/common/types/component"
	"github.com/kyma-project/modulectl/internal/service/componentdescriptor/resources"
	"github.com/kyma-project/modulectl/internal/service/contentprovider"
	"github.com/kyma-project/modulectl/internal/service/image"
)

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) AddResources(
	componentConstructor *component.Constructor,
	moduleConfig *contentprovider.ModuleConfig,
	resourcePaths *types.ResourcePaths,
) error {
	metadataYaml, err := resources.GenerateMetadataYaml(moduleConfig)
	if err != nil {
		return fmt.Errorf("failed to generate metadata yaml: %w", err)
	}

	componentConstructor.AddBinaryDataAsFileResource(common.MetadataResourceName, metadataYaml)
	err = componentConstructor.AddFileAsDirResource(common.RawManifestResourceName, resourcePaths.RawManifest)
	if err != nil {
		return fmt.Errorf("failed to create raw manifest resource: %w", err)
	}
	if resourcePaths.DefaultCR != "" {
		err = componentConstructor.AddFileAsDirResource(common.DefaultCRResourceName, resourcePaths.DefaultCR)
		if err != nil {
			return fmt.Errorf("failed to create default CR resource: %w", err)
		}
	}
	err = componentConstructor.AddFileResource(common.ModuleTemplateResourceName, resourcePaths.ModuleTemplate)
	if err != nil {
		return fmt.Errorf("failed to create moduletemplate resource: %w", err)
	}
	return nil
}

func (s *Service) CreateConstructorFile(componentConstructor *component.Constructor, filePath string) error {
	marshal, err := yaml.Marshal(componentConstructor)
	if err != nil {
		return fmt.Errorf("unable to marshal component constructor: %w", err)
	}

	filePermission := 0o600
	if err = os.WriteFile(filePath, marshal, os.FileMode(filePermission)); err != nil {
		return fmt.Errorf("unable to write component constructor file: %w", err)
	}
	return nil
}

func (s *Service) AddImagesToConstructor(
	componentConstructor *component.Constructor,
	images []string,
) error {
	imageInfos := make([]*image.ImageInfo, 0, len(images))
	for _, img := range images {
		imageInfo, err := image.ValidateAndParseImageInfo(img)
		if err != nil {
			return fmt.Errorf("image validation failed for %s: %w", img, err)
		}
		imageInfos = append(imageInfos, imageInfo)
	}
	componentConstructor.AddImageAsResource(imageInfos)
	return nil
}
