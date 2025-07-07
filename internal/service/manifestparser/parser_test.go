package manifestparser

import (
	"os"
	"path/filepath"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

func TestService_Parse(t *testing.T) {
	// Create a simple Deployment manifest as YAML
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
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "deploy.yaml")
	if err := os.WriteFile(tmpFile, []byte(manifest), 0644); err != nil {
		t.Fatalf("failed to write temp manifest: %v", err)
	}

	// Prepare expected object
	var obj unstructured.Unstructured
	if err := yaml.Unmarshal([]byte(manifest), &obj); err != nil {
		t.Fatalf("failed to unmarshal manifest: %v", err)
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
			want:    []*unstructured.Unstructured{&obj},
			wantErr: false,
		},
		{
			name:    "file does not exist",
			args:    struct{ path string }{path: "nonexistent.yaml"},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{}
			got, err := s.Parse(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(got) == 1 {
				// Compare Kind and Name only for brevity
				if got[0].GetKind() != tt.want[0].GetKind() || got[0].GetName() != tt.want[0].GetName() {
					t.Errorf("Parse() got = %v, want %v", got[0], tt.want[0])
				}
			}
			if tt.wantErr && got != nil {
				t.Errorf("Parse() expected nil result on error, got %v", got)
			}
		})
	}
}
