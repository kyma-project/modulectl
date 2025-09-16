package types

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type KeyValueArgs map[string]string

type RawManifestParser interface {
	Parse(filePath string) ([]*unstructured.Unstructured, error)
}

type ResourcePaths struct {
	DefaultCR      string
	RawManifest    string
	ModuleTemplate string
}

func NewResourcePaths(defaultCRPath, rawManifestPath, moduleTemplatePath string) *ResourcePaths {
	return &ResourcePaths{
		DefaultCR:      defaultCRPath,
		RawManifest:    rawManifestPath,
		ModuleTemplate: moduleTemplatePath,
	}
}
