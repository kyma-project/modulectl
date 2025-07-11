package manifestparser_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/kyma-project/modulectl/internal/service/manifestparser"
)

func TestNewParser_ReturnsValidParser(t *testing.T) {
	parser := manifestparser.NewParser()
	require.NotNil(t, parser)
	require.IsType(t, &manifestparser.ManifestParser{}, parser)
}

func TestParse_WhenFileNotFound_ReturnsError(t *testing.T) {
	parser := manifestparser.NewParser()

	_, err := parser.Parse("/nonexistent/file.yaml")

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to read manifest file")
}

func TestParse_WhenCalledWithValidMultiDocumentYAML_ReturnsManifests(t *testing.T) {
	parser := manifestparser.NewParser()
	content := `apiVersion: v1
kind: Namespace
metadata:
  name: test-namespace
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
spec:
  replicas: 1
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
---
apiVersion: v1
kind: Service
metadata:
  name: test-service
spec:
  ports:
  - port: 80`

	tmpFile := createTempFile(t, content)
	defer os.Remove(tmpFile)

	manifests, err := parser.Parse(tmpFile)

	require.NoError(t, err)
	require.Len(t, manifests, 3)
	require.Equal(t, "Namespace", manifests[0].GetKind())
	require.Equal(t, "test-namespace", manifests[0].GetName())
	require.Equal(t, "Deployment", manifests[1].GetKind())
	require.Equal(t, "test-deployment", manifests[1].GetName())
	require.Equal(t, "Service", manifests[2].GetKind())
	require.Equal(t, "test-service", manifests[2].GetName())
}

func TestParse_WhenCalledWithSingleDocumentYAML_ReturnsManifest(t *testing.T) {
	parser := manifestparser.NewParser()
	content := `apiVersion: v1
kind: ConfigMap
metadata:
  name: test-config
data:
  key: value`

	tmpFile := createTempFile(t, content)
	defer os.Remove(tmpFile)

	manifests, err := parser.Parse(tmpFile)

	require.NoError(t, err)
	require.Len(t, manifests, 1)
	require.Equal(t, "ConfigMap", manifests[0].GetKind())
	require.Equal(t, "test-config", manifests[0].GetName())
}

func TestParse_WhenCalledWithEmptyDocuments_SkipsEmptyDocuments(t *testing.T) {
	parser := manifestparser.NewParser()
	content := `apiVersion: v1
kind: Namespace
metadata:
  name: test-namespace
---
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-config
---

---`

	tmpFile := createTempFile(t, content)
	defer os.Remove(tmpFile)

	manifests, err := parser.Parse(tmpFile)

	require.NoError(t, err)
	require.Len(t, manifests, 2)
	require.Equal(t, "Namespace", manifests[0].GetKind())
	require.Equal(t, "ConfigMap", manifests[1].GetKind())
}

func TestParse_WhenCalledWithDocumentsWithoutKind_SkipsInvalidDocuments(t *testing.T) {
	parser := manifestparser.NewParser()
	content := `apiVersion: v1
kind: Namespace
metadata:
  name: test-namespace
---
apiVersion: v1
metadata:
  name: invalid-no-kind
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

func TestParse_WhenCalledWithInvalidYAML_ReturnsError(t *testing.T) {
	parser := manifestparser.NewParser()
	content := `apiVersion: v1
kind: Namespace
metadata:
  name: test-namespace
  invalid: [unclosed array`

	tmpFile := createTempFile(t, content)
	defer os.Remove(tmpFile)

	_, err := parser.Parse(tmpFile)

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse YAML document")
}

func TestParse_WhenCalledWithDocumentContainingSeparator_PreservesContent(t *testing.T) {
	parser := manifestparser.NewParser()
	content := `apiVersion: v1
kind: ConfigMap
metadata:
  name: test-config
data:
  script: |
    echo "--- This is not a separator ---"
    echo "Just content"
---
apiVersion: v1
kind: Namespace
metadata:
  name: test-namespace`

	tmpFile := createTempFile(t, content)
	defer os.Remove(tmpFile)

	manifests, err := parser.Parse(tmpFile)

	require.NoError(t, err)
	require.Len(t, manifests, 2)
	require.Equal(t, "ConfigMap", manifests[0].GetKind())
	require.Equal(t, "Namespace", manifests[1].GetKind())

	data, found, _ := unstructured.NestedStringMap(manifests[0].Object, "data")
	require.True(t, found)
	require.Contains(t, data["script"], "--- This is not a separator ---")
}

func TestParse_WhenCalledWithComplexDeployment_ParsesCorrectly(t *testing.T) {
	parser := manifestparser.NewParser()
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

	replicasFloat, ok := replicasRaw.(float64)
	require.True(t, ok)
	replicas := int64(replicasFloat)
	require.Equal(t, int64(3), replicas)
}

func TestParse_WhenCalledWithRealWorldManifest_ParsesCorrectly(t *testing.T) {
	parser := manifestparser.NewParser()
	content := `apiVersion: v1
kind: Namespace
metadata:
  name: template-operator-system
  labels:
    app.kubernetes.io/component: template-operator.kyma-project.io
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: template-operator-controller-manager
  namespace: template-operator-system
  labels:
    app.kubernetes.io/component: template-operator.kyma-project.io
spec:
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/component: template-operator.kyma-project.io
  template:
    metadata:
      labels:
        app.kubernetes.io/component: template-operator.kyma-project.io
    spec:
      containers:
      - name: manager
        image: europe-docker.pkg.dev/kyma-project/prod/template-operator:latest
        args:
        - --leader-elect
        - --final-state=Ready
        ports:
        - containerPort: 40000
        env:
        - name: WEBHOOK_IMAGE
          value: europe-docker.pkg.dev/kyma-project/prod/webhook:v1.0.0
        resources:
          limits:
            cpu: 500m
            memory: 128Mi
          requests:
            cpu: 10m
            memory: 64Mi
---
apiVersion: v1
kind: Service
metadata:
  name: template-operator-metrics-service
  namespace: template-operator-system
spec:
  ports:
  - name: https
    port: 8443
    protocol: TCP
    targetPort: https
  selector:
    app.kubernetes.io/component: template-operator.kyma-project.io`

	tmpFile := createTempFile(t, content)
	defer os.Remove(tmpFile)

	manifests, err := parser.Parse(tmpFile)

	require.NoError(t, err)
	require.Len(t, manifests, 3)

	kinds := make([]string, len(manifests))
	for i, manifest := range manifests {
		kinds[i] = manifest.GetKind()
	}
	require.Contains(t, kinds, "Namespace")
	require.Contains(t, kinds, "Deployment")
	require.Contains(t, kinds, "Service")
}

func createTempFile(t *testing.T, content string) string {
	t.Helper()
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.yaml")

	err := os.WriteFile(tmpFile, []byte(content), 0o600)
	require.NoError(t, err)

	return tmpFile
}
