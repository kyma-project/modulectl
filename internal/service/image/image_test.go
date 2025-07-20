package image_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kyma-project/modulectl/internal/service/image"
)

func TestParseImageReference_ValidImagesWithTag(t *testing.T) {
	tests := []struct {
		name         string
		imageURL     string
		expectedName string
		expectedTag  string
	}{
		{"basic image with tag", "alpine:3.15.4", "alpine", "3.15.4"},
		{"registry with port and tag", "localhost:5000/myimage:v1.0.0", "myimage", "v1.0.0"},
		{
			"complex registry path with tag",
			"europe-docker.pkg.dev/kyma-project/prod/external/istio/proxyv2:1.25.3-distroless",
			"proxyv2",
			"1.25.3-distroless",
		},
		{"gcr.io with tag", "gcr.io/kubebuilder/kube-rbac-proxy:v0.13.1", "kube-rbac-proxy", "v0.13.1"},
		{"docker hub with organization", "istio/proxyv2:1.19.0", "proxyv2", "1.19.0"},
		{"image name only with tag", "nginx:latest", "nginx", "latest"},
		{
			"registry with port and complex path",
			"localhost:5000/org/project/subproject/image:v1.2.3",
			"image",
			"v1.2.3",
		},
		{"multiple colons in registry path", "registry.io:5000/project/namespace/image:tag", "image", "tag"},
		{
			"k3d registry with port and path",
			"k3d-abc-registry.com:443/kyma-project/prod/bar-image:2.0.1",
			"bar-image",
			"2.0.1",
		},
		{"image with numeric tag", "nginx:1.21.0", "nginx", "1.21.0"},
		{"image with semantic version tag", "postgres:13.7-alpine", "postgres", "13.7-alpine"},
		// Removed the build tag test as it's not valid according to Docker reference format
		{"single character image name", "a:tag", "a", "tag"},
		{"numeric image name", "123:tag", "123", "tag"},
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

func TestParseImageReference_ValidImagesWithDigest(t *testing.T) {
	tests := []struct {
		name         string
		imageURL     string
		expectedName string
		expectedTag  string
	}{
		{
			"digest format",
			"docker.io/alpine@sha256:abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234",
			"alpine",
			"sha256:abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234",
		},
		{
			"complex digest",
			"gcr.io/distroless/static@sha256:1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			"static",
			"sha256:1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		},
		{
			"registry with port and digest",
			"localhost:5000/myimage@sha256:fedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321",
			"myimage",
			"sha256:fedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321",
		},
		{
			"sha512 digest",
			"alpine@sha512:abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234",
			"alpine",
			"sha512:abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234",
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

func TestParseImageReference_InvalidInputs_ReturnsError(t *testing.T) {
	tests := []struct {
		name        string
		imageURL    string
		expectedErr error
	}{
		{"empty string", "", image.ErrEmptyImageURL},
		{"no tag or digest", "docker.io/alpine", image.ErrNoTagOrDigest},
		{"registry port only", "localhost:5000/myimage", image.ErrNoTagOrDigest},
		{"path with colon but no valid tag", "registry.io:5000/path/to/image", image.ErrNoTagOrDigest},
		{"colon in path but no tag", "example.com:8080/project/image", image.ErrNoTagOrDigest},
		{"invalid digest format", "alpine@sha256:invalid", nil}, // Will be caught by reference.ParseAnyReference
		{"malformed reference", "alpine@sha256:", nil},          // Will be caught by reference.ParseAnyReference
		{"multiple @ symbols", "alpine@sha256:abc@def", nil},    // Will be caught by reference.ParseAnyReference
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, tag, err := image.ParseImageReference(tt.imageURL)
			require.Error(t, err)
			require.Empty(t, name)
			require.Empty(t, tag)

			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
			}
		})
	}
}

func TestParseImageReference_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		imageURL    string
		expectError bool
		expectName  string
		expectTag   string
	}{
		{
			// When both tag and digest are present, the reference library prioritizes tag, not digest
			"image with both tag and digest (tag takes precedence)",
			"nginx:1.21@sha256:abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234",
			false,
			"nginx",
			"1.21", // Tag is returned, not digest
		},
		{
			"image name with underscores",
			"my_image:v1.0.0",
			false,
			"my_image",
			"v1.0.0",
		},
		{
			"image name with hyphens",
			"my-image:v1.0.0",
			false,
			"my-image",
			"v1.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, tag, err := image.ParseImageReference(tt.imageURL)

			if tt.expectError {
				require.Error(t, err)
				require.Empty(t, name)
				require.Empty(t, tag)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectName, name)
				require.Equal(t, tt.expectTag, tag)
			}
		})
	}
}

func TestIsValidImage_ValidCases(t *testing.T) {
	tests := []struct {
		name     string
		imageURL string
	}{
		{"valid image with semantic version", "alpine:3.15.4"},
		{"valid image with build version", "myapp:v1.0.0-beta.1"},
		{
			"valid image with digest",
			"alpine@sha256:abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234",
		},
		{"valid complex registry path", "gcr.io/kubebuilder/kube-rbac-proxy:v0.13.1"},
		{"valid registry with port", "localhost:5000/myimage:v1.0.0"},
		{"valid docker hub image", "nginx:1.21.0"},
		{"valid image with hyphenated tag", "redis:6.2-alpine"},
		{"valid image with dot in tag", "postgres:13.7.0"},
		{"valid image with numeric tag", "myapp:123"},
		{"valid image with alphanumeric tag", "myapp:abc123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := image.IsValidImage(tt.imageURL)
			require.NoError(t, err)
			require.True(t, valid)
		})
	}
}

func TestIsValidImage_InvalidCases(t *testing.T) {
	tests := []struct {
		name      string
		imageURL  string
		wantErr   error
		errSubstr string
	}{
		{"image with latest tag", "nginx:latest", image.ErrDisallowedTag, "image tag is disallowed"},
		{"image with main tag", "nginx:main", image.ErrDisallowedTag, "image tag is disallowed"},
		{"image with LATEST tag (case insensitive)", "nginx:LATEST", image.ErrDisallowedTag, "image tag is disallowed"},
		{"image with Main tag (case insensitive)", "nginx:Main", image.ErrDisallowedTag, "image tag is disallowed"},
		{"malformed image", "nginx:::", nil, "invalid image reference"},
		{"invalid digest", "nginx@sha256:invalid", nil, "invalid image reference"},
		{"empty digest", "nginx@sha256:", nil, "invalid image reference"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := image.IsValidImage(tt.imageURL)
			require.False(t, valid)
			require.Error(t, err)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			}

			if tt.errSubstr != "" {
				require.Contains(t, err.Error(), tt.errSubstr)
			}
		})
	}
}

func TestIsValidImage_InvalidFormat_ReturnsFalse(t *testing.T) {
	tests := []struct {
		name     string
		imageURL string
	}{
		{"too short", "a:"},
		{"too long", strings.Repeat("a", 255) + ":tag"},
		{"with space", "nginx :latest"},
		{"with tab", "nginx\t:latest"},
		{"with newline", "nginx\n:latest"},
		{"with carriage return", "nginx\r:latest"},
		{"no separator", "nginx"},
		{"empty string", ""},
		{"only colon", ":"},
		{"only at symbol", "@"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := image.IsValidImage(tt.imageURL)
			require.False(t, valid)
			require.NoError(t, err) // isValidImageFormat returns false, so no error is propagated
		})
	}
}

func TestIsValidImage_InvalidFormat_ReturnsError(t *testing.T) {
	// These cases pass format validation but fail in parsing
	tests := []struct {
		name     string
		imageURL string
	}{
		{"colon at start", ":nginx"},
		{"at symbol at start", "@nginx"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := image.IsValidImage(tt.imageURL)
			require.False(t, valid)
			require.Error(t, err) // These will fail in ParseImageReference
			require.Contains(t, err.Error(), "invalid image reference")
		})
	}
}

func TestIsValidImage_MissingTag_Cases(t *testing.T) {
	tests := []struct {
		name        string
		imageURL    string
		expectError bool
	}{
		// These fail format validation (no : or @), so return false with no error
		{"simple image name without tag", "alpine", false},
		{"complex path without tag", "gcr.io/project/image", false},
		// This passes format validation but fails parsing
		{"registry with port but no tag", "localhost:5000/myimage", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := image.IsValidImage(tt.imageURL)
			require.False(t, valid)

			if tt.expectError {
				require.Error(t, err)
				require.Contains(t, err.Error(), "no tag or digest found")
			} else {
				require.NoError(t, err) // Failed format validation
			}
		})
	}
}

func TestIsValidImage_NoTagImage_ReturnsFalseNoError(t *testing.T) {
	// Test the specific case that was failing
	valid, err := image.IsValidImage("nginx")
	require.False(t, valid)
	require.NoError(t, err) // This fails isValidImageFormat, so no error
}

func TestIsValidImage_ErrorPropagation(t *testing.T) {
	tests := []struct {
		name        string
		imageURL    string
		expectError bool
		errorType   error
	}{
		{
			"valid format but parse error",
			"invalid@sha256:xyz",
			true,
			nil, // Will be a parsing error from reference package
		},
		{
			"valid format and parse but disallowed tag",
			"nginx:latest",
			true,
			image.ErrDisallowedTag,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := image.IsValidImage(tt.imageURL)
			require.False(t, valid)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorType != nil {
					require.ErrorIs(t, err, tt.errorType)
				}
			}
		})
	}
}

// Additional test for build tags that don't work with Docker reference format
func TestParseImageReference_InvalidBuildTag(t *testing.T) {
	// This test documents that build metadata in tags is not supported by Docker reference format
	name, tag, err := image.ParseImageReference("myimage:v1.0.0-rc.1+build.123")
	require.Error(t, err)
	require.Empty(t, name)
	require.Empty(t, tag)
	require.Contains(t, err.Error(), "invalid reference format")
}
