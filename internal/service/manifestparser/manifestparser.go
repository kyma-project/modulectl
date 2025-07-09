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

	documents := p.splitYAMLDocuments(string(content))

	for docIndex, doc := range documents {
		doc = strings.TrimSpace(doc)
		if doc == "" {
			continue
		}

		manifest, err := p.parseYAMLDocument([]byte(doc))
		if err != nil {
			// Skip documents that are missing required fields or are empty
			if errors.Is(err, ErrEmptyDocument) || errors.Is(err, ErrMissingKind) || errors.Is(err, ErrMissingAPIVersion) {
				continue
			}
			return nil, fmt.Errorf("failed to parse YAML document %d: %w", docIndex+1, err)
		}
		if manifest != nil {
			manifests = append(manifests, manifest)
		}
	}

	return manifests, nil
}

func (p *ManifestParser) splitYAMLDocuments(content string) []string {
	lines := strings.Split(content, "\n")
	var documents []string
	var currentDoc strings.Builder

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "---" {
			if currentDoc.Len() > 0 {
				documents = append(documents, currentDoc.String())
				currentDoc.Reset()
			}
		} else {
			currentDoc.WriteString(line)
			currentDoc.WriteString("\n")
		}
	}

	// Add the last document if it exists
	if currentDoc.Len() > 0 {
		documents = append(documents, currentDoc.String())
	}

	return documents
}

func (p *ManifestParser) parseYAMLDocument(content []byte) (*unstructured.Unstructured, error) {
	decoder := yaml.NewYAMLOrJSONDecoder(strings.NewReader(string(content)), yamlDecoderBufferSize)

	var rawObj map[string]interface{}
	if err := decoder.Decode(&rawObj); err != nil {
		if errors.Is(err, io.EOF) {
			return nil, ErrEmptyDocument
		}
		return nil, fmt.Errorf("failed to decode YAML: %w", err)
	}

	if rawObj == nil {
		return nil, ErrEmptyDocument
	}

	kind, hasKind := rawObj["kind"]
	if !hasKind || kind == nil || kind == "" {
		return nil, ErrMissingKind
	}

	apiVersion, hasAPIVersion := rawObj["apiVersion"]
	if !hasAPIVersion || apiVersion == nil || apiVersion == "" {
		return nil, ErrMissingAPIVersion
	}

	unstructuredObj := &unstructured.Unstructured{Object: rawObj}

	return unstructuredObj, nil
}
