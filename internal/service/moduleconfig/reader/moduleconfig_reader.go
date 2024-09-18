package moduleconfigreader

import (
	"fmt"
	"net/url"

	"gopkg.in/yaml.v3"

	commonerrors "github.com/kyma-project/modulectl/internal/common/errors"
	"github.com/kyma-project/modulectl/internal/common/validations"
	"github.com/kyma-project/modulectl/internal/service/contentprovider"
)

const (
	defaultCRFilePattern       = "kyma-module-default-cr-*.yaml"
	defaultManifestFilePattern = "kyma-module-manifest-*.yaml"
)

type FileSystem interface {
	ReadFile(path string) ([]byte, error)
}

type TempFileSystem interface {
	DownloadTempFile(dir, pattern string, url *url.URL) (string, error)
	RemoveTempFiles() []error
}

type Service struct {
	fileSystem    FileSystem
	tmpFileSystem TempFileSystem
}

func NewService(fileSystem FileSystem, tmpFileSystem TempFileSystem) (*Service, error) {
	if fileSystem == nil {
		return nil, fmt.Errorf("%w: fileSystem must not be nil", commonerrors.ErrInvalidArg)
	}

	if tmpFileSystem == nil {
		return nil, fmt.Errorf("%w: tmpFileSystem must not be nil", commonerrors.ErrInvalidArg)
	}

	return &Service{
		fileSystem:    fileSystem,
		tmpFileSystem: tmpFileSystem,
	}, nil
}

func (s *Service) ParseModuleConfig(configFilePath string) (*contentprovider.ModuleConfig, error) {
	moduleConfigData, err := s.fileSystem.ReadFile(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read module config file: %w", err)
	}

	moduleConfig := &contentprovider.ModuleConfig{}
	if err := yaml.Unmarshal(moduleConfigData, moduleConfig); err != nil {
		return nil, fmt.Errorf("failed to parse module config file: %w", err)
	}

	if moduleConfig.Namespace == "" {
		moduleConfig.Namespace = "kcp-system"
	}

	return moduleConfig, nil
}

func (s *Service) GetDefaultCRPath(defaultCRPath string) (string, error) {
	if defaultCRPath == "" {
		return defaultCRPath, nil
	}

	path := defaultCRPath
	if parsedURL, err := s.ParseURL(defaultCRPath); err == nil {
		path, err = s.tmpFileSystem.DownloadTempFile("", defaultCRFilePattern, parsedURL)
		if err != nil {
			return "", fmt.Errorf("failed to download default CR file: %w", err)
		}
	}

	return path, nil
}

func (s *Service) GetManifestPath(manifestPath string) (string, error) {
	path := manifestPath
	if parsedURL, err := s.ParseURL(manifestPath); err == nil {
		path, err = s.tmpFileSystem.DownloadTempFile("", defaultManifestFilePattern, parsedURL)
		if err != nil {
			return "", fmt.Errorf("failed to download Manifest file: %w", err)
		}
	}

	return path, nil
}

func (s *Service) ParseURL(urlString string) (*url.URL, error) {
	urlParsed, err := url.Parse(urlString)
	if err == nil && urlParsed.Scheme != "" && urlParsed.Host != "" {
		return urlParsed, nil
	}
	return nil, fmt.Errorf("%w: parsing url failed for %s", commonerrors.ErrInvalidArg, urlString)
}

func (*Service) ValidateModuleConfig(moduleConfig *contentprovider.ModuleConfig) error {
	if err := validations.ValidateModuleName(moduleConfig.Name); err != nil {
		return fmt.Errorf("failed to validate module name: %w", err)
	}

	if err := validations.ValidateModuleVersion(moduleConfig.Version); err != nil {
		return fmt.Errorf("failed to validate module version: %w", err)
	}

	if err := validations.ValidateModuleChannel(moduleConfig.Channel); err != nil {
		return fmt.Errorf("failed to validate module channel: %w", err)
	}

	if err := validations.ValidateModuleNamespace(moduleConfig.Namespace); err != nil {
		return fmt.Errorf("failed to validate module namespace: %w", err)
	}

	if moduleConfig.ManifestPath == "" {
		return fmt.Errorf("%w: manifest path must not be empty", commonerrors.ErrInvalidArg)
	}

	return nil
}

func (s *Service) CleanupTempFiles() []error {
	return s.tmpFileSystem.RemoveTempFiles()
}
