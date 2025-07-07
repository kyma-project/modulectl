package verifier_test

import (
	"testing"

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

// Implement the Parse method to satisfy manifestparser.Service
func (f *fakeParser) Parse(_ string) ([]*unstructured.Unstructured, error) {
	return f.resources, nil
}

func TestService_VerifyModuleResources(t *testing.T) {
	tests := []struct {
		name      string
		resources []*unstructured.Unstructured
		version   string
		manager   *contentprovider.Manager
		wantErr   bool
	}{
		{
			name: "Deployment with matching image tag",
			resources: []*unstructured.Unstructured{
				makeUnstructuredFromObj(&appsv1.Deployment{
					TypeMeta: metav1.TypeMeta{
						Kind: "Deployment",
					},
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{Image: "repo/test-manager:1.2.3"},
								},
							},
						},
					},
				}),
			},
			version: "1.2.3",
			manager: &contentprovider.Manager{Name: "test-manager"},
			wantErr: false,
		},
		{
			name: "Deployment with non-matching image tag",
			resources: []*unstructured.Unstructured{
				makeUnstructuredFromObj(&appsv1.Deployment{
					TypeMeta: metav1.TypeMeta{
						Kind: "Deployment",
					},
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{Image: "repo/test-manager:2.0.0"},
								},
							},
						},
					},
				}),
			},
			version: "1.2.3",
			manager: &contentprovider.Manager{Name: "test-manager"},
			wantErr: true,
		},
		{
			name: "StatefulSet with matching image tag",
			resources: []*unstructured.Unstructured{
				makeUnstructuredFromObj(&appsv1.StatefulSet{
					TypeMeta: metav1.TypeMeta{
						Kind: "StatefulSet",
					},
					Spec: appsv1.StatefulSetSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{Image: "repo/test-manager:1.2.3"},
								},
							},
						},
					},
				}),
			},
			version: "1.2.3",
			manager: &contentprovider.Manager{Name: "test-manager"},
			wantErr: false,
		},
		{
			name: "No matching container name",
			resources: []*unstructured.Unstructured{
				makeUnstructuredFromObj(&appsv1.Deployment{
					TypeMeta: metav1.TypeMeta{
						Kind: "Deployment",
					},
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{Image: "repo/other:1.2.3"},
								},
							},
						},
					},
				}),
			},
			version: "1.2.3",
			manager: &contentprovider.Manager{Name: "test-manager"},
			wantErr: true,
		},
		{
			name:      "No Deployment or StatefulSet",
			resources: []*unstructured.Unstructured{},
			version:   "1.2.3",
			manager:   &contentprovider.Manager{Name: "test-manager"},
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
				t.Errorf("verifyModuleImageVersion() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
