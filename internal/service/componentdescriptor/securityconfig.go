package componentdescriptor

import (
	"errors"
	"fmt"

	"gopkg.in/yaml.v3"
	"ocm.software/ocm/api/ocm/compdesc"
	ocmv1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"

	"github.com/kyma-project/modulectl/internal/common"
	commonerrors "github.com/kyma-project/modulectl/internal/common/errors"
	"github.com/kyma-project/modulectl/internal/common/types/component"
	"github.com/kyma-project/modulectl/internal/service/contentprovider"
)

const (
	SecBaseLabelKey = "security.kyma-project.io"

	ScanLabelKey   = "scan"
	SecScanEnabled = "enabled"
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
) error {
	if err := appendLabelToAccessor(descriptor, ScanLabelKey, SecScanEnabled, SecBaseLabelKey); err != nil {
		return fmt.Errorf("failed to append security label to descriptor: %w", err)
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

func (s *SecurityConfigService) AppendSecurityScanConfigToConstructor(constructor *component.Constructor) {
	constructor.AddLabel(fmt.Sprintf("%s/%s", SecBaseLabelKey, ScanLabelKey), SecScanEnabled, common.OCMVersion)
}
