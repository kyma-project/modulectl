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

func NewResourcePaths(DefaultCRPath, RawManifestPath, ModuleTemplatePath string) *ResourcePaths {
	return &ResourcePaths{
		DefaultCR:      DefaultCRPath,
		RawManifest:    RawManifestPath,
		ModuleTemplate: ModuleTemplatePath,
	}
}
