package crdparser

import (
	"fmt"
	"strings"

	"bytes"
	"gopkg.in/yaml.v3"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
)

type FileSystem interface {
	ReadFile(path string) ([]byte, error)
}

type Service struct {
	fileSystem FileSystem
}

func NewService(fileSystem FileSystem) *Service {
	return &Service{
		fileSystem: fileSystem,
	}
}

type resource struct {
	Kind       string `yaml:"kind"`
	ApiVersion string `yaml:"apiVersion"`
	Spec       struct {
		Group string `yaml:"group"`
		Names struct {
			Kind string `yaml:"kind"`
		}
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

	var cr resource
	if err := yaml.Unmarshal(crData, &cr); err != nil {
		return false, fmt.Errorf("error parsing default CR: %w", err)
	}

	group := strings.Split(cr.ApiVersion, "/")[0]

	manifestData, err := s.fileSystem.ReadFile(manifestPath)
	if err != nil {
		return false, fmt.Errorf("error reading manifest file: %w", err)
	}

	crdScope, err := getCrdScopeFromManifest(manifestData, group, cr.Kind)
	if err != nil {
		return false, fmt.Errorf("error finding CRD file in the %q file: %w", manifestPath, err)
	}

	return crdScope == apiextensions.ClusterScoped, nil
}

func getCrdScopeFromManifest(manifestData []byte, group, kind string) (apiextensions.ResourceScope, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(manifestData))

	for {
		var res resource

		// Decode each document in the YAML stream
		err := decoder.Decode(&res)
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return "", fmt.Errorf("failed to parse YAML document: %v", err)
		}

		if res.Kind == "CustomResourceDefinition" && res.Spec.Group == group && res.Spec.Names.Kind == kind {
			return res.Spec.Scope, nil
		}
	}

	return "", nil
}
