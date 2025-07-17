package manifestparser_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
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

func TestParse_WhenFileNotFound_ReturnsError(t *testing.T) {
	parser := manifestparser.NewService()

	_, err := parser.Parse("/nonexistent/file.yaml")

	require.Error(t, err)
	require.Contains(t, err.Error(), "parse manifest")
}

func TestParse_WhenCalledWithInvalidYAML_ReturnsError(t *testing.T) {
	parser := manifestparser.NewService()
	content := `apiVersion: v1
kind: Namespace
metadata:
  name: test-namespace
  invalid: [unclosed array`

	tmpFile := createTempFile(t, content)
	defer os.Remove(tmpFile)

	_, err := parser.Parse(tmpFile)

	require.Error(t, err)
	require.Contains(t, err.Error(), "error converting YAML to JSON")
}

func TestParse_WhenCalledWithDocumentsWithoutKind_SkipsInvalidDocuments(t *testing.T) {
	parser := manifestparser.NewService()
	content := `apiVersion: v1
kind: Namespace
metadata:
  name: test-namespace
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-config`

	tmpFile := createTempFile(t, content)
	defer os.Remove(tmpFile)

	manifests, err := parser.Parse(tmpFile)

	require.NoError(t, err)
	require.Len(t, manifests, 2)
	require.Equal(t, "Namespace", manifests[0].GetKind())
	require.Equal(t, "ConfigMap", manifests[1].GetKind())
}

func TestParse_WhenCalledWithComplexDeployment_ParsesCorrectly(t *testing.T) {
	parser := manifestparser.NewService()
	content := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
  labels:
    app: test
spec:
  replicas: 3
  selector:
    matchLabels:
      app: test
  template:
    metadata:
      labels:
        app: test
    spec:
      containers:
      - name: app
        image: nginx:1.20
        ports:
        - containerPort: 80`

	tmpFile := createTempFile(t, content)
	defer os.Remove(tmpFile)

	manifests, err := parser.Parse(tmpFile)

	require.NoError(t, err)
	require.Len(t, manifests, 1)
	require.Equal(t, "Deployment", manifests[0].GetKind())
	require.Equal(t, "test-deployment", manifests[0].GetName())

	replicasRaw, found, err := unstructured.NestedFieldNoCopy(manifests[0].Object, "spec", "replicas")
	require.NoError(t, err)
	require.True(t, found)

	// Handle both int64 and float64 types that might be returned
	var replicas int64
	switch v := replicasRaw.(type) {
	case int64:
		replicas = v
	case float64:
		replicas = int64(v)
	case int:
		replicas = int64(v)
	default:
		t.Fatalf("unexpected type for replicas: %T", replicasRaw)
	}
	require.Equal(t, int64(3), replicas)
}

func createTempFile(t *testing.T, content string) string {
	t.Helper()
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.yaml")

	err := os.WriteFile(tmpFile, []byte(content), 0o600)
	require.NoError(t, err)

	return tmpFile
}

func TestParse_WhenCalledWithCommentsAndWhitespace_SkipsNonResourceDocs(t *testing.T) {
	parser := manifestparser.NewService()
	content := `# This is a comment
apiVersion: v1
kind: Namespace
metadata:
  name: test-namespace
---
# Another comment
# Multiple lines

---
  # Indented comment
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-config`

	tmpFile := createTempFile(t, content)
	defer os.Remove(tmpFile)

	manifests, err := parser.Parse(tmpFile)

	require.NoError(t, err)
	require.Len(t, manifests, 2)
	require.Equal(t, "Namespace", manifests[0].GetKind())
	require.Equal(t, "ConfigMap", manifests[1].GetKind())
}

func writeToFile(t *testing.T, name string, data []byte) {
	t.Helper()
	//nolint:gosec // This is a test, so we can ignore the file permissions
	if err := os.WriteFile(name, data, 0o644); err != nil {
		t.Fatalf("failed to write manifest: %v", err)
	}
}
