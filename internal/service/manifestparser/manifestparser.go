package manifestparser

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
)

var (
	ErrEmptyDocument     = errors.New("empty document")
	ErrMissingKind       = errors.New("document missing kind field")
	ErrMissingAPIVersion = errors.New("document missing apiVersion field")
)

const yamlDecoderBufferSize = 4096

type ManifestParser struct{}

func NewParser() *ManifestParser {
	return &ManifestParser{}
}

func (p *ManifestParser) Parse(path string) ([]*unstructured.Unstructured, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest file %s: %w", path, err)
	}

	return p.parseYAMLContent(content)
}

func (p *ManifestParser) parseYAMLContent(content []byte) ([]*unstructured.Unstructured, error) {
	var manifests []*unstructured.Unstructured

	decoder := yaml.NewYAMLOrJSONDecoder(strings.NewReader(string(content)), yamlDecoderBufferSize)

	docIndex := 0
	for {
		var rawObj map[string]interface{}
		err := decoder.Decode(&rawObj)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf("failed to parse YAML document %d: failed to decode YAML: %w", docIndex+1, err)
		}

		// Skip empty docs
		if rawObj == nil {
			continue
		}

		kind, hasKind := rawObj["kind"]
		if !hasKind || kind == nil || kind == "" {
			continue // or return ErrMissingKind if you want strict behavior
		}
		apiVersion, hasAPIVersion := rawObj["apiVersion"]
		if !hasAPIVersion || apiVersion == nil || apiVersion == "" {
			continue // or return ErrMissingAPIVersion
		}

		unstructuredObj := &unstructured.Unstructured{Object: rawObj}
		manifests = append(manifests, unstructuredObj)
		docIndex++
	}

	return manifests, nil
}
