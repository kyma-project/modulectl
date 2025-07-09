package manifestparser

import (
	"fmt"
	"io"
	"os"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
)

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

	// Split YAML documents by separator
	documents := p.splitYAMLDocuments(string(content))

	for i, doc := range documents {
		doc = strings.TrimSpace(doc)
		if doc == "" {
			continue
		}

		manifest, err := p.parseYAMLDocument([]byte(doc))
		if err != nil {
			return nil, fmt.Errorf("failed to parse YAML document %d: %w", i+1, err)
		}

		if manifest != nil {
			manifests = append(manifests, manifest)
		}
	}

	return manifests, nil
}

func (p *ManifestParser) splitYAMLDocuments(content string) []string {
	// Split by --- but be careful about lines that contain --- as part of content
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
	decoder := yaml.NewYAMLOrJSONDecoder(strings.NewReader(string(content)), 4096)

	var rawObj map[string]interface{}
	if err := decoder.Decode(&rawObj); err != nil {
		if err == io.EOF {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to decode YAML: %w", err)
	}

	if rawObj == nil {
		return nil, nil
	}

	kind, hasKind := rawObj["kind"]
	if !hasKind || kind == nil || kind == "" {
		return nil, nil
	}

	apiVersion, hasAPIVersion := rawObj["apiVersion"]
	if !hasAPIVersion || apiVersion == nil || apiVersion == "" {
		return nil, nil
	}

	unstructuredObj := &unstructured.Unstructured{Object: rawObj}

	return unstructuredObj, nil
}
