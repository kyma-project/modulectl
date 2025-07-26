package componentdescriptor

import (
	"errors"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
	"ocm.software/ocm/api/ocm/compdesc"
	ocmv1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"

	commonerrors "github.com/kyma-project/modulectl/internal/common/errors"
	"github.com/kyma-project/modulectl/internal/service/contentprovider"
)

const (
	secBaseLabelKey     = "security.kyma-project.io"
	secScanBaseLabelKey = "scan.security.kyma-project.io"
	scanLabelKey        = "scan"
	secScanEnabled      = "enabled"
	rcTagLabelKey       = "rc-tag"
	languageLabelKey    = "language"
	devBranchLabelKey   = "dev-branch"
	subProjectsLabelKey = "subprojects"
	excludeLabelKey     = "exclude"
	ocmIdentityName     = "module-sources"
	ocmVersion          = "v1"
	refLabel            = "git.kyma-project.io/ref"
)

var ErrSecurityConfigFileDoesNotExist = errors.New("security config file does not exist")

type FileReader interface {
	FileExists(path string) (bool, error)
	ReadFile(path string) ([]byte, error)
}

type SecurityConfigService struct {
	fileReader FileReader
}

func NewSecurityConfigService(fileReader FileReader) (*SecurityConfigService, error) {
	if fileReader == nil {
		return nil, fmt.Errorf("fileReader must not be nil: %w", commonerrors.ErrInvalidArg)
	}

	return &SecurityConfigService{
		fileReader: fileReader,
	}, nil
}

func (s *SecurityConfigService) ParseSecurityConfigData(securityConfigFile string) (
	*contentprovider.SecurityScanConfig,
	error,
) {
	exists, err := s.fileReader.FileExists(securityConfigFile)
	if err != nil {
		return nil, fmt.Errorf("failed to check if security config file exists: %w", err)
	}
	if !exists {
		return nil, ErrSecurityConfigFileDoesNotExist
	}

	securityConfigContent, err := s.fileReader.ReadFile(securityConfigFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read security config file: %w", err)
	}

	securityConfig := &contentprovider.SecurityScanConfig{}
	if err := yaml.Unmarshal(securityConfigContent, securityConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal security config file: %w", err)
	}

	return securityConfig, nil
}

func (s *SecurityConfigService) AppendSecurityScanConfig(descriptor *compdesc.ComponentDescriptor,
	securityConfig contentprovider.SecurityScanConfig,
) error {
	if err := appendLabelToAccessor(descriptor, scanLabelKey, secScanEnabled, secBaseLabelKey); err != nil {
		return fmt.Errorf("failed to append security label to descriptor: %w", err)
	}

	if err := AppendSecurityLabelsToSources(securityConfig, descriptor.Sources); err != nil {
		return fmt.Errorf("failed to append security labels to sources: %w", err)
	}

	return nil
}

func AppendSecurityLabelsToSources(securityScanConfig contentprovider.SecurityScanConfig,
	sources compdesc.Sources,
) error {
	for srcIndex := range sources {
		src := &sources[srcIndex]
		if err := appendLabelToAccessor(src, rcTagLabelKey, securityScanConfig.RcTag,
			secScanBaseLabelKey); err != nil {
			return fmt.Errorf("failed to append security label to source: %w", err)
		}

		if err := appendLabelToAccessor(src, languageLabelKey,
			securityScanConfig.Mend.Language, secScanBaseLabelKey); err != nil {
			return fmt.Errorf("failed to append security label to source: %w", err)
		}

		if err := appendLabelToAccessor(src, devBranchLabelKey, securityScanConfig.DevBranch,
			secScanBaseLabelKey); err != nil {
			return fmt.Errorf("failed to append security label to source: %w", err)
		}

		if err := appendLabelToAccessor(src, subProjectsLabelKey,
			securityScanConfig.Mend.SubProjects, secScanBaseLabelKey); err != nil {
			return fmt.Errorf("failed to append security label to source: %w", err)
		}

		excludeMendProjects := strings.Join(securityScanConfig.Mend.Exclude, ",")
		if err := appendLabelToAccessor(src, excludeLabelKey,
			excludeMendProjects, secScanBaseLabelKey); err != nil {
			return fmt.Errorf("failed to append security label to source: %w", err)
		}
	}

	return nil
}

func appendLabelToAccessor(labeled compdesc.LabelsAccessor, key, value, baseKey string) error {
	labels := labeled.GetLabels()
	securityLabelKey := fmt.Sprintf("%s/%s", baseKey, key)
	labelValue, err := ocmv1.NewLabel(securityLabelKey, value, ocmv1.WithVersion(ocmVersion))
	if err != nil {
		return fmt.Errorf("failed to create security label: %w", err)
	}
	labels = append(labels, *labelValue)
	labeled.SetLabels(labels)
	return nil
}
