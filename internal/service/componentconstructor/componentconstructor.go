package componentconstructor

import (
	"fmt"

	"github.com/kyma-project/modulectl/internal/common/types/component"
	"github.com/kyma-project/modulectl/internal/service/contentprovider"
	"github.com/kyma-project/modulectl/internal/service/image"
	"github.com/kyma-project/modulectl/tools/io"
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
	cmdOutput io.Out,
	outputFile string,
) error {
	cmdOutput.Write("- Generating module resources\n")
	componentConstructor.AddRawManifestResource(manifestFilePath)
	if defaultCRFilePath != "" {
		componentConstructor.AddDefaultCRResource(defaultCRFilePath)
	}
	componentConstructor.AddMetadataResource(moduleConfig)

	cmdOutput.Write("- Creating component constructor file\n")
	return componentConstructor.CreateComponentConstructorFile(outputFile)
}

func (s *Service) AddImagesToConstructor(
	componentConstructor *component.Constructor,
	images []string,
) error {
	var imageInfos []*image.ImageInfo
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
