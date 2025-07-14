package image_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/kyma-project/modulectl/internal/service/image"
)

func TestNewService_WhenCalledWithValidParser_ReturnsService(t *testing.T) {
	mockParser := &mockManifestParser{}

	service, err := image.NewService(mockParser)

	require.NoError(t, err)
	require.NotNil(t, service)
}

func TestNewService_WhenCalledWithNilParser_ReturnsError(t *testing.T) {
	service, err := image.NewService(nil)

	require.Error(t, err)
	require.Nil(t, service)
	require.Contains(t, err.Error(), "manifestParser must not be nil")
	require.ErrorIs(t, err, image.ErrParserNil)
}

func TestParseImageReference_WhenCalledWithValidImageWithTag_ReturnsNameAndTag(t *testing.T) {
	tests := []struct {
		name         string
		imageURL     string
		expectedName string
		expectedTag  string
	}{
		{
			name:         "basic image with tag",
			imageURL:     "alpine:3.15.4",
			expectedName: "alpine",
			expectedTag:  "3.15.4",
		},
		{
			name:         "registry with port and tag",
			imageURL:     "localhost:5000/myimage:v1.0.0",
			expectedName: "myimage",
			expectedTag:  "v1.0.0",
		},
		{
			name:         "complex registry path with tag",
			imageURL:     "europe-docker.pkg.dev/kyma-project/prod/external/istio/proxyv2:1.25.3-distroless",
			expectedName: "proxyv2",
			expectedTag:  "1.25.3-distroless",
		},
		{
			name:         "gcr.io with tag",
			imageURL:     "gcr.io/kubebuilder/kube-rbac-proxy:v0.13.1",
			expectedName: "kube-rbac-proxy",
			expectedTag:  "v0.13.1",
		},
		{
			name:         "docker hub with organization",
			imageURL:     "istio/proxyv2:1.19.0",
			expectedName: "proxyv2",
			expectedTag:  "1.19.0",
		},
		{
			name:         "image name only with tag",
			imageURL:     "nginx:latest",
			expectedName: "nginx",
			expectedTag:  "latest",
		},
		{
			name:         "registry with port and complex path",
			imageURL:     "localhost:5000/org/project/subproject/image:v1.2.3",
			expectedName: "image",
			expectedTag:  "v1.2.3",
		},
		{
			name:         "multiple colons in registry path",
			imageURL:     "registry.io:5000/project/namespace/image:tag",
			expectedName: "image",
			expectedTag:  "tag",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, tag, err := image.ParseImageReference(tt.imageURL)

			require.NoError(t, err)
			require.Equal(t, tt.expectedName, name)
			require.Equal(t, tt.expectedTag, tag)
		})
	}
}

func TestParseImageReference_WhenCalledWithValidImageWithDigest_ReturnsNameAndDigest(t *testing.T) {
	tests := []struct {
		name         string
		imageURL     string
		expectedName string
		expectedTag  string
	}{
		{
			name:         "digest format",
			imageURL:     "docker.io/alpine@sha256:abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234",
			expectedName: "alpine",
			expectedTag:  "sha256:abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234",
		},
		{
			name:         "complex digest",
			imageURL:     "gcr.io/distroless/static@sha256:1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			expectedName: "static",
			expectedTag:  "sha256:1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		},
		{
			name:         "registry with port and digest",
			imageURL:     "localhost:5000/myimage@sha256:fedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321",
			expectedName: "myimage",
			expectedTag:  "sha256:fedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, tag, err := image.ParseImageReference(tt.imageURL)

			require.NoError(t, err)
			require.Equal(t, tt.expectedName, name)
			require.Equal(t, tt.expectedTag, tag)
		})
	}
}

func TestParseImageReference_WhenCalledWithInvalidInputs_ReturnsError(t *testing.T) {
	tests := []struct {
		name        string
		imageURL    string
		expectedErr string
	}{
		{
			name:        "empty string",
			imageURL:    "",
			expectedErr: "empty image URL",
		},
		{
			name:        "no tag or digest",
			imageURL:    "docker.io/alpine",
			expectedErr: "no tag or digest found in docker.io/alpine: no tag or digest found",
		},
		{
			name:        "registry port only",
			imageURL:    "localhost:5000/myimage",
			expectedErr: "no tag or digest found in localhost:5000/myimage: no tag or digest found",
		},
		{
			name:        "path with colon but no valid tag",
			imageURL:    "registry.io:5000/path/to/image",
			expectedErr: "no tag or digest found in registry.io:5000/path/to/image: no tag or digest found",
		},
		{
			name:        "colon in path but no tag",
			imageURL:    "example.com:8080/project/image",
			expectedErr: "no tag or digest found in example.com:8080/project/image: no tag or digest found",
		},
		{
			name:        "invalid digest format",
			imageURL:    "alpine@sha256:invalid",
			expectedErr: "invalid image reference",
		},
		{
			name:        "malformed reference",
			imageURL:    "alpine@sha256:",
			expectedErr: "invalid image reference",
		},
		{
			name:        "multiple @ symbols",
			imageURL:    "alpine@sha256:abc@def",
			expectedErr: "invalid image reference",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, tag, err := image.ParseImageReference(tt.imageURL)

			require.Error(t, err)
			require.Empty(t, name)
			require.Empty(t, tag)
			require.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

type mockManifestParser struct {
	manifests []*unstructured.Unstructured
	err       error
}

func (m *mockManifestParser) Parse(path string) ([]*unstructured.Unstructured, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.manifests, nil
}
