package componentdescriptor

import (
	"errors"
	"fmt"
	"strings"

	"github.com/kyma-project/modulectl/internal/common"
	"github.com/kyma-project/modulectl/internal/common/types/component"
	"ocm.software/ocm/api/ocm/compdesc"
	ocmv1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"

	commonerrors "github.com/kyma-project/modulectl/internal/common/errors"
	"github.com/kyma-project/modulectl/internal/service/contentprovider"
	"gopkg.in/yaml.v3"
)

const (
	SecBaseLabelKey = "security.kyma-project.io"

	ScanLabelKey        = "scan"
	SecScanEnabled      = "enabled"
	RCTagLabelKey       = "rc-tag"
	LanguageLabelKey    = "language"
	DevBranchLabelKey   = "dev-branch"
	SubProjectsLabelKey = "subprojects"
	ExcludeLabelKey     = "exclude"
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
	if err := appendLabelToAccessor(descriptor, ScanLabelKey, SecScanEnabled, SecBaseLabelKey); err != nil {
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
		if err := appendLabelToAccessor(src, RCTagLabelKey, securityScanConfig.RcTag,
			common.SecScanBaseLabelKey); err != nil {
			return fmt.Errorf("failed to append security label to source: %w", err)
		}

		if err := appendLabelToAccessor(src, LanguageLabelKey,
			securityScanConfig.Mend.Language, common.SecScanBaseLabelKey); err != nil {
			return fmt.Errorf("failed to append security label to source: %w", err)
		}

		if err := appendLabelToAccessor(src, DevBranchLabelKey, securityScanConfig.DevBranch,
			common.SecScanBaseLabelKey); err != nil {
			return fmt.Errorf("failed to append security label to source: %w", err)
		}

		if err := appendLabelToAccessor(src, SubProjectsLabelKey,
			securityScanConfig.Mend.SubProjects, common.SecScanBaseLabelKey); err != nil {
			return fmt.Errorf("failed to append security label to source: %w", err)
		}

		excludeMendProjects := strings.Join(securityScanConfig.Mend.Exclude, ",")
		if err := appendLabelToAccessor(src, ExcludeLabelKey,
			excludeMendProjects, common.SecScanBaseLabelKey); err != nil {
			return fmt.Errorf("failed to append security label to source: %w", err)
		}
	}

	return nil
}

func appendLabelToAccessor(labeled compdesc.LabelsAccessor, key, value, baseKey string) error {
	labels := labeled.GetLabels()
	securityLabelKey := fmt.Sprintf("%s/%s", baseKey, key)
	labelValue, err := ocmv1.NewLabel(securityLabelKey, value, ocmv1.WithVersion(common.OCMVersion))
	if err != nil {
		return fmt.Errorf("failed to create security label: %w", err)
	}
	labels = append(labels, *labelValue)
	labeled.SetLabels(labels)
	return nil
}

func (s *SecurityConfigService) AppendSecurityScanConfigToConstructor(constructor *component.Constructor,
	securityConfig contentprovider.SecurityScanConfig) {
	constructor.AddLabel(fmt.Sprintf("%s/%s", SecBaseLabelKey, ScanLabelKey), SecScanEnabled, common.OCMVersion)

	labelKeyValues := map[string]string{
		RCTagLabelKey:       securityConfig.RcTag,
		LanguageLabelKey:    securityConfig.Mend.Language,
		DevBranchLabelKey:   securityConfig.DevBranch,
		SubProjectsLabelKey: securityConfig.Mend.SubProjects,
		ExcludeLabelKey:     strings.Join(securityConfig.Mend.Exclude, ","),
	}
	for key, value := range labelKeyValues {
		constructor.AddLabelToSources(fmt.Sprintf("%s/%s", common.SecScanBaseLabelKey, key), value,
			common.OCMVersion)
	}
}
