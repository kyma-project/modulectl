package verifier

import (
	"fmt"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	apiappsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

type ImageVersionVerifier struct {
	version     string
	managerName string
}

func NewImageVersionVerifier(version, managerName string) *ImageVersionVerifier {
	return &ImageVersionVerifier{
		version:     version,
		managerName: managerName,
	}
}

func (i *ImageVersionVerifier) VerifyModuleImageVersion(resources []*unstructured.Unstructured) error {
	for _, res := range resources {
		if res.GetKind() == "Deployment" {
			var deploy apiappsv1.Deployment
			err := runtime.DefaultUnstructuredConverter.FromUnstructured(res.Object, &deploy)
			if err != nil {
				return fmt.Errorf("failed to convert unstructured to Deployment: %w", err)
			}
			if i.foundMatchedVersionInDeployment(deploy) {
				return nil
			}
		}
		if res.GetKind() == "StatefulSet" {
			var statefulSet apiappsv1.StatefulSet
			err := runtime.DefaultUnstructuredConverter.FromUnstructured(res.Object, &statefulSet)
			if err != nil {
				return fmt.Errorf("failed to convert unstructured to StatefulSet: %w", err)
			}
			if i.foundMatchedVersionInStatefulSet(statefulSet) {
				return nil
			}
		}
	}
	return fmt.Errorf("no matched version %s found in Deployment or StatefulSet for managerName %s", i.version,
		i.managerName)
}

func (i *ImageVersionVerifier) foundMatchedVersionInContainers(containers []corev1.Container) bool {
	for _, c := range containers {
		if strings.Contains(c.Image, i.managerName) {
			imageTag, err := getImageTag(c.Image)
			if err != nil {
				return false
			}
			if imageTag == i.version {
				return true
			}
		}
	}
	return false
}

func (i *ImageVersionVerifier) foundMatchedVersionInDeployment(deploy apiappsv1.Deployment) bool {
	return i.foundMatchedVersionInContainers(deploy.Spec.Template.Spec.Containers)
}

func (i *ImageVersionVerifier) foundMatchedVersionInStatefulSet(statefulSet apiappsv1.StatefulSet) bool {
	return i.foundMatchedVersionInContainers(statefulSet.Spec.Template.Spec.Containers)
}

func getImageTag(image string) (string, error) {
	ref, err := name.ParseReference(image)
	if err != nil {
		return "", fmt.Errorf("invalid image reference: %w", err)
	}
	if tag, ok := ref.(name.Tag); ok {
		return tag.TagStr(), nil
	}
	return "", fmt.Errorf("image does not have a tag")
}
