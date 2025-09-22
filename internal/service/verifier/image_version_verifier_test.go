package verifier_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/kyma-project/modulectl/internal/service/contentprovider"
	"github.com/kyma-project/modulectl/internal/service/verifier"
)

func makeUnstructuredFromObj(obj interface{}) *unstructured.Unstructured {
	u := &unstructured.Unstructured{}
	m, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		panic(err)
	}
	u.SetUnstructuredContent(m)
	return u
}

type fakeParser struct {
	resources []*unstructured.Unstructured
}

// Implement the Parse method to satisfy manifestparser.Service.
func (f *fakeParser) Parse(_ string) ([]*unstructured.Unstructured, error) {
	return f.resources, nil
}

func TestService_VerifyModuleResources(t *testing.T) {
	gvkDeployment := &metav1.GroupVersionKind{
		Group:   "apps",
		Version: "v1",
		Kind:    "Deployment",
	}

	gvkStatefulSet := &metav1.GroupVersionKind{
		Group:   "apps",
		Version: "v1",
		Kind:    "StatefulSet",
	}

	tests := []struct {
		name      string
		resources []*unstructured.Unstructured
		version   string
		manager   *contentprovider.Manager
		wantErr   bool
	}{
		{
			name: "Deployment with matching image tag and name",
			resources: []*unstructured.Unstructured{
				makeUnstructuredFromObj(&appsv1.Deployment{
					TypeMeta:   metav1.TypeMeta{Kind: "Deployment"},
					ObjectMeta: metav1.ObjectMeta{Name: "test-manager"},
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{Image: "repo/test-manager:1.2.3", Name: "manager"},
								},
							},
						},
					},
				}),
			},
			version: "1.2.3",
			manager: &contentprovider.Manager{Name: "test-manager", GroupVersionKind: *gvkDeployment},
			wantErr: false,
		},
		{
			name: "StatefulSet with matching image tag and name",
			resources: []*unstructured.Unstructured{
				makeUnstructuredFromObj(&appsv1.StatefulSet{
					TypeMeta:   metav1.TypeMeta{Kind: "StatefulSet"},
					ObjectMeta: metav1.ObjectMeta{Name: "test-manager"},
					Spec: appsv1.StatefulSetSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{Image: "repo/test-manager:1.2.3", Name: "manager"},
								},
							},
						},
					},
				}),
			},
			version: "1.2.3",
			manager: &contentprovider.Manager{Name: "test-manager", GroupVersionKind: *gvkStatefulSet},
			wantErr: false,
		},
		{
			name: "Deployment with non-matching image tag",
			resources: []*unstructured.Unstructured{
				makeUnstructuredFromObj(&appsv1.Deployment{
					TypeMeta:   metav1.TypeMeta{Kind: "Deployment"},
					ObjectMeta: metav1.ObjectMeta{Name: "test-manager"},
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{Image: "repo/test-manager:2.0.0", Name: "manager"},
								},
							},
						},
					},
				}),
			},
			version: "1.2.3",
			manager: &contentprovider.Manager{Name: "test-manager", GroupVersionKind: *gvkDeployment},
			wantErr: true,
		},
		{
			name: "No matching manager name",
			resources: []*unstructured.Unstructured{
				makeUnstructuredFromObj(&appsv1.Deployment{
					TypeMeta:   metav1.TypeMeta{Kind: "Deployment"},
					ObjectMeta: metav1.ObjectMeta{Name: "other-manager"},
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{Image: "repo/other:1.2.3", Name: "manager"},
								},
							},
						},
					},
				}),
			},
			version: "1.2.3",
			manager: &contentprovider.Manager{Name: "test-manager", GroupVersionKind: *gvkDeployment},
			wantErr: true,
		},
		{
			name: "Container name mismatch",
			resources: []*unstructured.Unstructured{
				makeUnstructuredFromObj(&appsv1.Deployment{
					TypeMeta:   metav1.TypeMeta{Kind: "Deployment"},
					ObjectMeta: metav1.ObjectMeta{Name: "test-manager"},
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{Image: "repo/other:1.2.3", Name: "other"},
								},
							},
						},
					},
				}),
			},
			version: "1.2.3",
			manager: &contentprovider.Manager{Name: "test-manager", GroupVersionKind: *gvkDeployment},
			wantErr: true,
		},
		{
			name:      "No resources",
			resources: []*unstructured.Unstructured{},
			version:   "1.2.3",
			manager:   &contentprovider.Manager{Name: "test-manager", GroupVersionKind: *gvkDeployment},
			wantErr:   true,
		},
		{
			name:      "No manager in config",
			resources: []*unstructured.Unstructured{},
			version:   "1.2.3",
			manager:   nil,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := fakeParser{resources: tt.resources}
			svc := verifier.NewService(&parser)
			cfg := &contentprovider.ModuleConfig{
				Version: tt.version,
				Manager: tt.manager,
			}
			err := svc.VerifyModuleResources(cfg, "dummy.yaml")
			if (err != nil) != tt.wantErr {
				t.Errorf("VerifyModuleResources() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_VerifyModuleResources_ParseError(t *testing.T) {
	parser := &fakeParserWithError{}
	svc := verifier.NewService(parser)
	configWithManager := &contentprovider.ModuleConfig{
		Version: "1.2.3",
		Manager: &contentprovider.Manager{Name: "test-manager"},
	}
	err := svc.VerifyModuleResources(configWithManager, "dummy.yaml")
	require.ErrorIs(t, err, errParse)

	configWithoutManager := &contentprovider.ModuleConfig{
		Version: "1.2.3",
		Manager: nil,
	}
	err = svc.VerifyModuleResources(configWithoutManager, "dummy.yaml")
	require.ErrorIs(t, err, errParse)
}

type fakeParserWithError struct{}

var errParse = errors.New("parse error")

func (f *fakeParserWithError) Parse(_ string) ([]*unstructured.Unstructured, error) {
	return nil, errParse
}
