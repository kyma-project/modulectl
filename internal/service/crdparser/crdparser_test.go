package crdparser_test

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"

	"github.com/kyma-project/modulectl/internal/common/types"
	"github.com/kyma-project/modulectl/internal/service/crdparser"
)

const (
	defaultCRPath   = "/path/to/defaultCR"
	rawManifestPath = "/path/to/manifest"
)

func TestService_NewService_ReturnsErrorWhenCalledWithNil(t *testing.T) {
	_, err := crdparser.NewService(nil)

	require.Error(t, err)
}

func TestService_IsCRDClusterScoped_ReturnsTrueWhenClusterScoped(t *testing.T) {
	crdParserService, _ := crdparser.NewService(&fileSystemClusterScopedExistsStub{})

	resourcePaths := types.NewResourcePaths(defaultCRPath, rawManifestPath, "")
	isClusterScoped, err := crdParserService.IsCRDClusterScoped(resourcePaths)
	require.NoError(t, err)
	require.True(t, isClusterScoped)
}

func TestService_IsCRDClusterScoped_ReturnsFalseWhenNamespaceScoped(t *testing.T) {
	crdParserService, _ := crdparser.NewService(&fileSystemNamespacedScopedExistsStub{})

	resourcePaths := types.NewResourcePaths(defaultCRPath, rawManifestPath, "")
	isClusterScoped, err := crdParserService.IsCRDClusterScoped(resourcePaths)
	require.NoError(t, err)
	require.False(t, isClusterScoped)
}

func TestService_IsCRDClusterScoped_ReturnsErrorWhenFileReadingRetrievalError(t *testing.T) {
	crdParserService, _ := crdparser.NewService(&fileSystemNotExistStub{})

	resourcePaths := types.NewResourcePaths(defaultCRPath, rawManifestPath, "")
	_, err := crdParserService.IsCRDClusterScoped(resourcePaths)
	require.ErrorContains(t, err, "error reading default CR file")
}

type fileSystemClusterScopedExistsStub struct{}

func (*fileSystemClusterScopedExistsStub) ReadFile(path string) ([]byte, error) {
	var fileContent []byte
	if strings.Contains(path, "defaultCR") {
		content := crdparser.Resource{
			APIVersion: "operator.kyma-project.io/v1",
			Kind:       "Sample",
		}
		fileContent, _ = yaml.Marshal(content)
	} else {
		content := []crdparser.Resource{
			{
				APIVersion: "v1",
				Kind:       "Namespace",
			},
			{
				APIVersion: "apiextensions.k8s.io/v1",
				Kind:       "CustomResourceDefinition",
				Spec: struct {
					Group string `yaml:"group"`
					Names struct {
						Kind string `yaml:"kind"`
					} `yaml:"names"`
					Scope apiextensions.ResourceScope `yaml:"scope"`
				}{
					Group: "operator.kyma-project.io",
					Names: struct {
						Kind string `yaml:"kind"`
					}{
						Kind: "Sample",
					},
					Scope: "Cluster",
				},
			},
			{
				APIVersion: "apiextensions.k8s.io/v1",
				Kind:       "CustomResourceDefinition",
				Spec: struct {
					Group string `yaml:"group"`
					Names struct {
						Kind string `yaml:"kind"`
					} `yaml:"names"`
					Scope apiextensions.ResourceScope `yaml:"scope"`
				}{
					Group: "operator.kyma-project.io",
					Names: struct {
						Kind string `yaml:"kind"`
					}{
						Kind: "Managed",
					},
					Scope: "Namespaced",
				},
			},
		}
		var buffer bytes.Buffer

		for i, res := range content {
			data, err := yaml.Marshal(res)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal resource: %w", err)
			}

			if i > 0 {
				buffer.WriteString("\n---\n") // Add separator for multiple resources
			}
			buffer.Write(data)
		}

		fileContent = buffer.Bytes()
	}
	return fileContent, nil
}

type fileSystemNamespacedScopedExistsStub struct{}

func (*fileSystemNamespacedScopedExistsStub) ReadFile(path string) ([]byte, error) {
	var fileContent []byte
	if strings.Contains(path, "defaultCR") {
		content := crdparser.Resource{
			APIVersion: "operator.kyma-project.io/v1",
			Kind:       "Sample",
		}
		fileContent, _ = yaml.Marshal(content)
	} else {
		content := []crdparser.Resource{
			{
				APIVersion: "v1",
				Kind:       "Namespace",
			},
			{
				APIVersion: "apiextensions.k8s.io/v1",
				Kind:       "CustomResourceDefinition",
				Spec: struct {
					Group string `yaml:"group"`
					Names struct {
						Kind string `yaml:"kind"`
					} `yaml:"names"`
					Scope apiextensions.ResourceScope `yaml:"scope"`
				}{
					Group: "operator.kyma-project.io",
					Names: struct {
						Kind string `yaml:"kind"`
					}{
						Kind: "Sample",
					},
					Scope: "Namespaced",
				},
			},
			{
				APIVersion: "apiextensions.k8s.io/v1",
				Kind:       "CustomResourceDefinition",
				Spec: struct {
					Group string `yaml:"group"`
					Names struct {
						Kind string `yaml:"kind"`
					} `yaml:"names"`
					Scope apiextensions.ResourceScope `yaml:"scope"`
				}{
					Group: "operator.kyma-project.io",
					Names: struct {
						Kind string `yaml:"kind"`
					}{
						Kind: "Managed",
					},
					Scope: "Cluster",
				},
			},
		}
		var buffer bytes.Buffer

		for i, res := range content {
			data, err := yaml.Marshal(res)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal resource: %w", err)
			}

			if i > 0 {
				buffer.WriteString("\n---\n") // Add separator for multiple resources
			}
			buffer.Write(data)
		}
		fileContent = buffer.Bytes()
	}
	return fileContent, nil
}

type fileSystemNotExistStub struct{}

func (*fileSystemNotExistStub) ReadFile(_ string) ([]byte, error) {
	return nil, errors.New("failed to read file")
}
