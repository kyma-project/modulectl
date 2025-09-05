package componentconstructor

import (
	"fmt"

	"github.com/kyma-project/modulectl/internal/common/types/component"
	"github.com/kyma-project/modulectl/internal/service/contentprovider"
	"github.com/kyma-project/modulectl/internal/service/image"
	iotools "github.com/kyma-project/modulectl/tools/io"
)

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) AddResourcesAndCreateConstructorFile(
	componentConstructor *component.Constructor,
	moduleConfig *contentprovider.ModuleConfig,
	manifestFilePath string,
	defaultCRFilePath string,
	cmdOutput iotools.Out,
	outputFile string,
) error {
	cmdOutput.Write("- Generating module resources\n")
	componentConstructor.AddRawManifestResource(manifestFilePath)
	if defaultCRFilePath != "" {
		componentConstructor.AddDefaultCRResource(defaultCRFilePath)
	}
	if err := componentConstructor.AddMetadataResource(moduleConfig); err != nil {
		return fmt.Errorf("failed to add metadata resource: %w", err)
	}

	cmdOutput.Write("- Creating component constructor file\n")
	if err := componentConstructor.CreateComponentConstructorFile(outputFile); err != nil {
		return fmt.Errorf("failed to create component constructor file: %w", err)
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
