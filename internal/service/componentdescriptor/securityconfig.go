package componentdescriptor

import (
	"errors"
	"fmt"
	"strings"

	ocm "ocm.software/ocm/api/ocm/compdesc"
	ocmv1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/ociartifact"
	ociartifacttypes "ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/types/ociartifact"

	"github.com/kyma-project/modulectl/internal/service/contentprovider"
	"gopkg.in/yaml.v3"
)

var (
	errInvalidURL = errors.New("invalid image URL")
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
	imageTagSlicesLength      = 2
	ocmIdentityName           = "module-sources"
	ocmVersion                = "v1"
	refLabel                  = "git.kyma-project.io/ref"
)

type SecurityConfigService struct {
	gitService GitService
}

func NewSecurityConfigService(gitService GitService) *SecurityConfigService {
	return &SecurityConfigService{
		gitService: gitService,
	}
}

func (s *SecurityConfigService) ParseSecurityConfigData(gitRepoURL, securityConfigFile string) (
	*contentprovider.SecurityScanConfig,
	error) {
	latestCommit, err := s.gitService.GetLatestCommit(gitRepoURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest commit: %w", err)
	}

	securityConfigContent, err := s.gitService.GetRemoteGitFileContent(gitRepoURL, latestCommit, securityConfigFile)
	if err != nil {
		return nil, fmt.Errorf("failed to get security config content: %w", err)
	}

	securityConfig := &contentprovider.SecurityScanConfig{}
	if err := yaml.Unmarshal([]byte(securityConfigContent), securityConfig); err != nil {
		return nil, fmt.Errorf("failed to parse security config file: %w", err)
	}

	return securityConfig, nil
}

func (s *SecurityConfigService) AppendSecurityScanConfig(descriptor *ocm.ComponentDescriptor,
	securityConfig contentprovider.SecurityScanConfig) error {
	if err := appendSecurityLabelToDescriptor(descriptor); err != nil {
		return fmt.Errorf("failed to append security label to descriptor: %w", err)
	}

	if err := appendSecurityLabelsToSources(securityConfig, descriptor.Sources); err != nil {
		return fmt.Errorf("failed to append security labels to sources: %w", err)
	}

	if err := appendProtecodeImagesLayers(descriptor, securityConfig); err != nil {
		return fmt.Errorf("failed to append protecode images layers: %w", err)
	}

	return nil
}

func appendSecurityLabelToDescriptor(descriptor *ocm.ComponentDescriptor) error {
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

func appendSecurityLabelsToSources(securityScanConfig contentprovider.SecurityScanConfig,
	sources ocm.Sources) error {
	for _, src := range sources {
		if err := appendLabelToSource(&src, rcTagLabelKey, securityScanConfig.RcTag); err != nil {
			return fmt.Errorf("failed to append security label to source: %w", err)
		}

		if err := appendLabelToSource(&src, languageLabelKey,
			securityScanConfig.WhiteSource.Language); err != nil {
			return fmt.Errorf("failed to append security label to source: %w", err)
		}

		if err := appendLabelToSource(&src, devBranchLabelKey, securityScanConfig.DevBranch); err != nil {
			return fmt.Errorf("failed to append security label to source: %w", err)
		}

		if err := appendLabelToSource(&src, subProjectsLabelKey,
			securityScanConfig.WhiteSource.SubProjects); err != nil {
			return fmt.Errorf("failed to append security label to source: %w", err)
		}

		excludeWhiteSourceProjects := strings.Join(securityScanConfig.WhiteSource.Exclude, ",")
		if err := appendLabelToSource(&src, excludeLabelKey,
			excludeWhiteSourceProjects); err != nil {
			return fmt.Errorf("failed to append security label to source: %w", err)
		}
	}

	return nil
}

func appendProtecodeImagesLayers(componentDescriptor *ocm.ComponentDescriptor,
	securityScanConfig contentprovider.SecurityScanConfig) error {
	protecodeImages := securityScanConfig.Protecode
	for _, img := range protecodeImages {
		imgName, imgTag, err := getImageNameAndTag(img)
		if err != nil {
			return fmt.Errorf("failed to get image name and tag: %w", err)
		}

		imageTypeLabelKey := fmt.Sprintf("%s/%s", secScanBaseLabelKey, typeLabelKey)
		imageTypeLabel, err := ocmv1.NewLabel(imageTypeLabelKey, thirdPartyImageLabelValue)
		if err != nil {
			return fmt.Errorf("failed to create security label: %w", err)
		}

		access := ociartifact.New(img)
		access.SetType(ociartifact.LegacyType)
		if err != nil {
			return fmt.Errorf("failed to convert access to unstructured object: %w", err)
		}
		proteccodeImageLayer := ocm.Resource{
			ResourceMeta: ocm.ResourceMeta{
				Type:     ociartifacttypes.LEGACY_TYPE,
				Relation: ocmv1.ExternalRelation,
				ElementMeta: ocm.ElementMeta{
					Name:    imgName,
					Labels:  []ocmv1.Label{*imageTypeLabel},
					Version: imgTag,
				},
			},
			Access: access,
		}
		componentDescriptor.Resources = append(componentDescriptor.Resources, proteccodeImageLayer)
	}
	ocm.DefaultResources(componentDescriptor)
	if err := ocm.Validate(componentDescriptor); err != nil {
		return fmt.Errorf("failed to validate component descriptor: %w", err)
	}

	return nil
}

func appendLabelToSource(source *ocm.Source, labelKey, labelValue string) error {
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
	if len(imageTag) != imageTagSlicesLength {
		return "", "", fmt.Errorf("%w: , image URL: %s", errInvalidURL, imageURL)
	}

	imageName := strings.Split(imageTag[0], "/")
	if len(imageName) == 0 {
		return "", "", fmt.Errorf("%w: , image URL: %s", errInvalidURL, imageURL)
	}

	return imageName[len(imageName)-1], imageTag[len(imageTag)-1], nil
}
