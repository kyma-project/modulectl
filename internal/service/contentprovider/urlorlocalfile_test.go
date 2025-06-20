package contentprovider_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	commonerrors "github.com/kyma-project/modulectl/internal/common/errors"
	"github.com/kyma-project/modulectl/internal/service/contentprovider"
)

func Test_UrlOrLocalFile_FromString_Succeeds_WhenCorrectURL(t *testing.T) {
	var res contentprovider.UrlOrLocalFile
	err := res.FromString("https://example.com/config.yaml")

	require.NoError(t, err)
	assert.True(t, res.IsURL())
	assert.False(t, res.IsEmpty())
	assert.Equal(t, "https", res.URL().Scheme)
}

func Test_UrlOrLocalFile_FromString_Fails_When_IncorrectURL(t *testing.T) {
	err := (&contentprovider.UrlOrLocalFile{}).FromString("https:///config.yaml")

	require.ErrorIs(t, err, commonerrors.ErrInvalidArg)
	assert.Contains(t, err.Error(), "Missing host")
}

func Test_UrlOrLocalFile_MustUrlOrLocalFile_Succeeds_WhenLocalFile(t *testing.T) {
	localFileRef := contentprovider.MustUrlOrLocalFile("manifest.yaml")

	assert.Equal(t, "manifest.yaml", localFileRef.String())
}

func Test_UrlOrLocalFile_MustUrlOrLocalFile_Panics_WhenInvalidURL(t *testing.T) {
	willPanicFn := func() {
		contentprovider.MustUrlOrLocalFile("\\\\https://example.com/manifest.yaml")
	}
	require.Panics(t, willPanicFn, "MustUrlOrLocalFile should panic on invalid URL")
}
