package crdparser

import (
	"bytes"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"

	commonerrors "github.com/kyma-project/modulectl/internal/common/errors"
)

type FileSystem interface {
	ReadFile(path string) ([]byte, error)
}

type Service struct {
	fileSystem FileSystem
}

func NewService(fileSystem FileSystem) (*Service, error) {
	if fileSystem == nil {
		return nil, fmt.Errorf("%w: fileSystem must not be nil", commonerrors.ErrInvalidArg)
	}

	return &Service{
		fileSystem: fileSystem,
	}, nil
}

type Resource struct {
	Kind       string `yaml:"kind"`
	APIVersion string `yaml:"apiVersion"`
	Spec       struct {
		Group string `yaml:"group"`
		Names struct {
			Kind string `yaml:"kind"`
		} `yaml:"names"`
		Scope apiextensions.ResourceScope `yaml:"scope"`
	} `yaml:"spec"`
}

func (s *Service) IsCRDClusterScoped(crPath, manifestPath string) (bool, error) {
	if crPath == "" {
		return false, nil
	}

	crData, err := s.fileSystem.ReadFile(crPath)
	if err != nil {
		return false, fmt.Errorf("error reading CRD file: %w", err)
	}

	var customResource Resource
	if err := yaml.Unmarshal(crData, &customResource); err != nil {
		return false, fmt.Errorf("error parsing default CR: %w", err)
	}

	group := strings.Split(customResource.APIVersion, "/")[0]

	manifestData, err := s.fileSystem.ReadFile(manifestPath)
	if err != nil {
		return false, fmt.Errorf("error reading manifest file: %w", err)
	}

	crdScope, err := getCrdScopeFromManifest(manifestData, group, customResource.Kind)
	if err != nil {
		return false, fmt.Errorf("error finding CRD file in the %q file: %w", manifestPath, err)
	}

	return crdScope == apiextensions.ClusterScoped, nil
}

func getCrdScopeFromManifest(manifestData []byte, group, kind string) (apiextensions.ResourceScope, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(manifestData))

	for {
		var res Resource
		err := decoder.Decode(&res)
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return "", fmt.Errorf("failed to parse YAML document: %w", err)
		}

		if res.Kind == "CustomResourceDefinition" && res.Spec.Group == group && res.Spec.Names.Kind == kind {
			return res.Spec.Scope, nil
		}
	}

	return "", nil
}
