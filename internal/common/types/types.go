package types

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type KeyValueArgs map[string]string

type RawManifestParser interface {
	Parse(filePath string) ([]*unstructured.Unstructured, error)
}
