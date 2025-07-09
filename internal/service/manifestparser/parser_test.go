package manifestparser_test

import (
	"os"
	"path/filepath"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"

	"github.com/kyma-project/modulectl/internal/service/manifestparser"
)

func TestService_Parse(t *testing.T) {
	// Single resource manifest
	manifest := `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deploy
spec:
  template:
    spec:
      containers:
      - name: test
        image: nginx:1.14.2
`
	// Multiple resources in one file
	multiManifest := `
apiVersion: v1
kind: Service
metadata:
  name: test-svc
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deploy
`
	tmpDir := t.TempDir()

	// Write single resource file
	tmpFile := filepath.Join(tmpDir, "deploy.yaml")
	writeToFile(t, tmpFile, []byte(manifest))

	// Write multiple resources file
	multiFile := filepath.Join(tmpDir, "multi.yaml")
	writeToFile(t, multiFile, []byte(multiManifest))

	// Write empty file
	emptyFile := filepath.Join(tmpDir, "empty.yaml")
	writeToFile(t, emptyFile, []byte(""))

	// Prepare expected objects
	var obj1 unstructured.Unstructured
	if err := yaml.Unmarshal([]byte(manifest), &obj1); err != nil {
		t.Fatalf("failed to unmarshal manifest: %v", err)
	}
	// For multiManifest, unmarshal both
	var svcObj, deployObj unstructured.Unstructured
	docs := [][]byte{}
	for _, doc := range []string{
		"apiVersion: v1\nkind: Service\nmetadata:\n  name: test-svc\n",
		"apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: test-deploy\n",
	} {
		docs = append(docs, []byte(doc))
	}
	if err := yaml.Unmarshal(docs[0], &svcObj); err != nil {
		t.Fatalf("failed to unmarshal service: %v", err)
	}
	if err := yaml.Unmarshal(docs[1], &deployObj); err != nil {
		t.Fatalf("failed to unmarshal deployment: %v", err)
	}

	tests := []struct {
		name    string
		args    struct{ path string }
		want    []*unstructured.Unstructured
		wantErr bool
	}{
		{
			name:    "valid deployment manifest",
			args:    struct{ path string }{path: tmpFile},
			want:    []*unstructured.Unstructured{&obj1},
			wantErr: false,
		},
		{
			name:    "file does not exist",
			args:    struct{ path string }{path: "nonexistent.yaml"},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "multiple resources in one file",
			args:    struct{ path string }{path: multiFile},
			want:    []*unstructured.Unstructured{&svcObj, &deployObj},
			wantErr: false,
		},
		{
			name:    "empty file",
			args:    struct{ path string }{path: emptyFile},
			want:    []*unstructured.Unstructured{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &manifestparser.Service{}
			got, err := s.Parse(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(got) != len(tt.want) {
					t.Errorf("Parse() got %d objects, want %d", len(got), len(tt.want))
					return
				}
				for i := range got {
					if got[i].GetKind() != tt.want[i].GetKind() || got[i].GetName() != tt.want[i].GetName() {
						t.Errorf("Parse() got[%d] = %v, want %v", i, got[i], tt.want[i])
					}
				}
			}
			if tt.wantErr && got != nil {
				t.Errorf("Parse() expected nil result on error, got %v", got)
			}
		})
	}
}

func writeToFile(t *testing.T, name string, data []byte) {
	t.Helper()
	//nolint:gosec // This is a test, so we can ignore the file permissions
	if err := os.WriteFile(name, data, 0o644); err != nil {
		t.Fatalf("failed to write manifest: %v", err)
	}
}
