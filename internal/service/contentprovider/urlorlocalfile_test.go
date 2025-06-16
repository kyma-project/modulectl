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
}

func Test_UrlOrLocalFile_FromString_Fails_When_IncorrectURL(t *testing.T) {
	err := (&contentprovider.UrlOrLocalFile{}).FromString("https:///config.yaml")

	require.ErrorIs(t, err, commonerrors.ErrInvalidArg)
	assert.Contains(t, err.Error(), "Missing host")
}
