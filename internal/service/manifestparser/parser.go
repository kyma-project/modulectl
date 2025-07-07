package manifestparser

import (
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/cli-runtime/pkg/resource"
)

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Parse(path string) ([]*unstructured.Unstructured, error) {
	objects := []*unstructured.Unstructured{}
	builder := resource.NewLocalBuilder().
		Unstructured().
		Path(false, path).
		Flatten().
		ContinueOnError()

	result := builder.Do()

	if err := result.Err(); err != nil {
		return nil, fmt.Errorf("parse manifest: %w", err)
	}
	items, err := result.Infos()
	if err != nil {
		return nil, fmt.Errorf("parse manifest to resource infos: %w", err)
	}
	for _, item := range items {
		unstructuredItem, ok := item.Object.(*unstructured.Unstructured)
		if !ok {
			continue
		}

		objects = append(objects, unstructuredItem)
	}
	return objects, nil
}
