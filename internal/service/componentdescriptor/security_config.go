package componentdescriptor

import (
	"fmt"

	"gopkg.in/yaml.v3"
	ocmv1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	ocm "ocm.software/ocm/api/ocm/compdesc/versions/v2"

	"github.com/kyma-project/modulectl/internal/service/contentprovider"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/ociartifact"
	"ocm.software/ocm/api/utils/runtime"
	ociartifact2 "ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/types/ociartifact"
	"strings"
)

const (
	secBaseLabelKey           = "security.kyma-project.io"
	secScanBaseLabelKey       = "scan.security.kyma-project.io"
	scanLabelKey              = "scan"
	secScanEnabled            = "enabled"
	rcTagLabelKey             = "rc-tag"
	languageLabelKey          = "language"
	devBranchLabelKey         = "dev-branch"
	subProjectsLabelKey       = "subprojects"
	excludeLabelKey           = "exclude"
	typeLabelKey              = "type"
	thirdPartyImageLabelValue = "third-party-image"
)

type FileSystem interface {
	FileExists(path string) (bool, error)
	ReadFile(path string) ([]byte, error)
}

type SecurityConfigService struct {
	fileSystem           FileSystem
	Sources              ocm.Sources
	securiyScannerConfig *contentprovider.SecurityScanConfig
}

func (s *SecurityConfigService) NewSecurityConfig(fileSystem FileSystem) *SecurityConfigService {
	return &SecurityConfigService{
		fileSystem: fileSystem,
	}

}

func (s *SecurityConfigService) AddSecurityConfigData(securityConfigFile string) error {
	if exists, err := s.fileSystem.FileExists(securityConfigFile); err != nil {
		return fmt.Errorf("failed to check if security config file exists: %w", err)
	} else if !exists {
		return nil
	}

	return nil
}

func (s *SecurityConfigService) ParseSecurityConfig(securityConfigFile string) error {
	securityConfigData, err := s.fileSystem.ReadFile(securityConfigFile)
	if err != nil {
		return fmt.Errorf("failed to read module config file: %w", err)
	}

	securityConfig := &contentprovider.SecurityScanConfig{}
	if err := yaml.Unmarshal(securityConfigData, securityConfig); err != nil {
		return fmt.Errorf("failed to parse security config file: %w", err)
	}

	s.securiyScannerConfig = securityConfig

	return nil
}

func (s *SecurityConfigService) AppendSecurityLabelToDescriptor(descriptor ocm.ComponentDescriptor) error {
	labels := descriptor.GetLabels()
	securityLabelKey := fmt.Sprintf("%s/%s", secBaseLabelKey, scanLabelKey)
	securityLabel, err := ocmv1.NewLabel(securityLabelKey, secScanEnabled)
	if err != nil {
		return fmt.Errorf("failed to create security label: %w", err)
	}

	labels = append(labels, *securityLabel)
	descriptor.SetLabels(labels)
	return nil
}

func (s *SecurityConfigService) AppendSecurityLabelsToSources() error {
	if s.Sources == nil {
		return fmt.Errorf("no sources found in component descriptor")
	}

	for i := range s.Sources {
		if err := appendLabelToSource(s.Sources[i], rcTagLabelKey, s.securiyScannerConfig.RcTag); err != nil {
			return fmt.Errorf("failed to append security label to source: %w", err)
		}

		if err := appendLabelToSource(s.Sources[i], languageLabelKey,
			s.securiyScannerConfig.WhiteSource.Language); err != nil {
			return fmt.Errorf("failed to append security label to source: %w", err)
		}

		if err := appendLabelToSource(s.Sources[i], devBranchLabelKey, s.securiyScannerConfig.DevBranch); err != nil {
			return fmt.Errorf("failed to append security label to source: %w", err)
		}

		if err := appendLabelToSource(s.Sources[i], subProjectsLabelKey,
			s.securiyScannerConfig.WhiteSource.SubProjects); err != nil {
			return fmt.Errorf("failed to append security label to source: %w", err)
		}

		excludeWhiteSourceProjects := strings.Join(s.securiyScannerConfig.WhiteSource.Exclude, ",")
		if err := appendLabelToSource(s.Sources[i], excludeLabelKey,
			excludeWhiteSourceProjects); err != nil {
			return fmt.Errorf("failed to append security label to source: %w", err)
		}
	}

	return nil
}

func (s *SecurityConfigService) ConstructProtecodeImagesLayers() (ocm.Resources, error) {
	var resources ocm.Resources
	protecodeImages := s.securiyScannerConfig.Protecode
	for i := range protecodeImages {
		imgName, imgTag, err := getImageNameAndTag(protecodeImages[i])
		if err != nil {
			return nil, fmt.Errorf("failed to get image name and tag: %w", err)
		}

		imageTypeLabelKey := fmt.Sprintf("%s/%s", secScanBaseLabelKey, typeLabelKey)
		imageTypeLabel, err := ocmv1.NewLabel(imageTypeLabelKey, thirdPartyImageLabelValue)
		if err != nil {
			return nil, fmt.Errorf("failed to create security label: %w", err)
		}

		access := ociartifact.New(protecodeImages[i])
		access.SetType(ociartifact.LegacyType)
		accessUnstructured, err := runtime.ToUnstructuredTypedObject(access)
		if err != nil {
			return nil, fmt.Errorf("failed to convert access to unstructured object: %w", err)
		}
		proteccodeImageLayer := ocm.Resource{
			ElementMeta: ocm.ElementMeta{
				Name:    imgName,
				Labels:  []ocmv1.Label{*imageTypeLabel},
				Version: imgTag,
			},
			Type:     ociartifact2.LEGACY_TYPE,
			Relation: ocmv1.ExternalRelation,
			Access:   accessUnstructured,
		}
		resources = append(resources, proteccodeImageLayer)
	}
	return resources, nil
}

func appendLabelToSource(source ocm.Source, labelKey, labelValue string) error {
	labels := source.GetLabels()
	securityLabelKey := fmt.Sprintf("%s/%s", secScanBaseLabelKey, labelKey)
	securityLabel, err := ocmv1.NewLabel(securityLabelKey, labelValue)
	if err != nil {
		return fmt.Errorf("failed to create security label: %w", err)
	}
	labels = append(labels, *securityLabel)
	source.SetLabels(labels)
	return nil
}

func getImageNameAndTag(imageURL string) (string, string, error) {
	imageTag := strings.Split(imageURL, ":")
	if len(imageTag) != 2 {
		return "", "", fmt.Errorf("invalid image URL: %s", imageURL)
	}

	imageName := strings.Split(imageTag[0], "/")
	if len(imageName) == 0 {
		return "", "", fmt.Errorf("invalid image URL: %s", imageURL)
	}

	return imageName[len(imageName)-1], imageTag[len(imageTag)-1], nil
}
