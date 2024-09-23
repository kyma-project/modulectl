package moduleconfigreader

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"

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

func (s *Service) ParseAndValidateModuleConfig(moduleConfigFile string,
) (*contentprovider.ModuleConfig, error) {
	moduleConfig, err := ParseModuleConfig(moduleConfigFile, s.fileSystem)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to parse module config file", err)
	}

	if err := ValidateModuleConfig(moduleConfig); err != nil {
		return nil, fmt.Errorf("%w: failed to value module config", err)
	}

	moduleConfig.DefaultCRPath, err = GetDefaultCRPath(moduleConfig.DefaultCRPath, s.tmpFileSystem)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to get default CR path", err)
	}

	moduleConfig.ManifestPath, err = GetManifestPath(moduleConfig.ManifestPath, s.tmpFileSystem)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to get manifest path", err)
	}

	return moduleConfig, nil
}

func (s *Service) GetDefaultCRData(defaultCRPath string) ([]byte, error) {
	defaultCRData, err := s.fileSystem.ReadFile(defaultCRPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read default CR file: %w", err)
	}

	return defaultCRData, nil
}

func (s *Service) CleanupTempFiles() []error {
	return s.tmpFileSystem.RemoveTempFiles()
}

func GetManifestPath(manifestPath string, tempFileSystem TempFileSystem) (string, error) {
	path := manifestPath

	if parsedURL, err := ParseURL(manifestPath); err == nil {
		path, err = tempFileSystem.DownloadTempFile("", defaultManifestFilePattern, parsedURL)
		if err != nil {
			return "", fmt.Errorf("failed to download Manifest file: %w", err)
		}
		return path, nil
	}

	if !filepath.IsAbs(manifestPath) {
		// Get the current working directory
		homeDir, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get the current directory: %w", err)
		}
		// Get the relative path from the current directory
		path = filepath.Join(homeDir, path)
		path, err = filepath.Abs(path)
		if err != nil {
			return "", fmt.Errorf("failed to obtain absolute path to manifest file: %w", err)
		}
		return path, nil
	}

	return path, nil
}

func ParseURL(urlString string) (*url.URL, error) {
	urlParsed, err := url.Parse(urlString)
	if err == nil && urlParsed.Scheme != "" && urlParsed.Host != "" {
		return urlParsed, nil
	}
	return nil, fmt.Errorf("%w: parsing url failed for %s", commonerrors.ErrInvalidArg, urlString)
}

func ValidateModuleConfig(moduleConfig *contentprovider.ModuleConfig) error {
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
		return fmt.Errorf("%w: manifest path must not be empty", commonerrors.ErrInvalidOption)
	}

	return nil
}

func ParseModuleConfig(configFilePath string, fileSystem FileSystem) (*contentprovider.ModuleConfig, error) {
	moduleConfigData, err := fileSystem.ReadFile(configFilePath)
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

func GetDefaultCRPath(defaultCRPath string, tempFileSystem TempFileSystem) (string, error) {
	if defaultCRPath == "" {
		return defaultCRPath, nil
	}

	path := defaultCRPath

	if parsedURL, err := ParseURL(defaultCRPath); err == nil {
		path, err = tempFileSystem.DownloadTempFile("", defaultCRFilePattern, parsedURL)
		if err != nil {
			return "", fmt.Errorf("failed to download default CR file: %w", err)
		}
		return path, nil
	}

	if !filepath.IsAbs(defaultCRPath) {
		// Get the current working directory
		homeDir, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get the current working directory: %w", err)
		}
		// Get the relative path from the current directory
		path = filepath.Join(homeDir, path)
		path, err = filepath.Abs(path)
		if err != nil {
			return "", fmt.Errorf("failed to obtain absolute path to deefault CR file: %w", err)
		}
		return path, nil
	}

	return path, nil
}
