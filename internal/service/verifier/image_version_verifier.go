package verifier

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/kyma-project/modulectl/internal/service/contentprovider"
)

type Service struct {
	rawManifestParser RawManifestParser
}

var (
	errImageNoTag            = errors.New("no image tag")
	errNoMatchedVersionFound = errors.New("no matched version found")
)

type RawManifestParser interface {
	Parse(filePath string) ([]*unstructured.Unstructured, error)
}

func NewService(parser RawManifestParser) *Service {
	return &Service{
		rawManifestParser: parser,
	}
}

func (s *Service) VerifyModuleResources(moduleConfig *contentprovider.ModuleConfig, filePath string) error {
	resources, err := s.rawManifestParser.Parse(filePath)
	if err != nil {
		return fmt.Errorf("failed to parse raw manifest: %w", err)
	}
	if moduleConfig.Manager == nil {
		return nil
	}

	if err := verifyModuleImageVersion(resources, moduleConfig.Version, moduleConfig.Manager.Name); err != nil {
		return fmt.Errorf("failed to verify module image version: %w", err)
	}
	return nil
}

func verifyModuleImageVersion(resources []*unstructured.Unstructured, version string,
	managerName string,
) error {
	for _, res := range resources {
		if res.GetKind() == "Deployment" {
			var deploy appsv1.Deployment
			err := runtime.DefaultUnstructuredConverter.FromUnstructured(res.Object, &deploy)
			if err != nil {
				return fmt.Errorf("failed to convert unstructured to Deployment: %w", err)
			}
			if foundMatchedVersionInDeployment(deploy, version, managerName) {
				return nil
			}
		}
		if res.GetKind() == "StatefulSet" {
			var statefulSet appsv1.StatefulSet
			err := runtime.DefaultUnstructuredConverter.FromUnstructured(res.Object, &statefulSet)
			if err != nil {
				return fmt.Errorf("failed to convert unstructured to StatefulSet: %w", err)
			}
			if foundMatchedVersionInStatefulSet(statefulSet, version, managerName) {
				return nil
			}
		}
	}
	return fmt.Errorf("no matched version %s found in Deployment or StatefulSet for managerName %s: %w", version,
		managerName, errNoMatchedVersionFound)
}

func foundMatchedVersionInContainers(containers []corev1.Container, version string,
	managerName string,
) bool {
	for _, c := range containers {
		if strings.Contains(c.Image, managerName) {
			imageTag, err := getImageTag(c.Image)
			if err != nil {
				return false
			}
			if imageTag == version {
				return true
			}
		}
	}
	return false
}

func foundMatchedVersionInDeployment(deploy appsv1.Deployment, version string,
	managerName string,
) bool {
	return foundMatchedVersionInContainers(deploy.Spec.Template.Spec.Containers, version, managerName)
}

func foundMatchedVersionInStatefulSet(statefulSet appsv1.StatefulSet, version string,
	managerName string,
) bool {
	return foundMatchedVersionInContainers(statefulSet.Spec.Template.Spec.Containers, version, managerName)
}

func getImageTag(image string) (string, error) {
	ref, err := name.ParseReference(image)
	if err != nil {
		return "", fmt.Errorf("invalid image reference: %w", err)
	}
	if tag, ok := ref.(name.Tag); ok {
		return tag.TagStr(), nil
	}
	return "", errImageNoTag
}
