package accesshandler_test

import (
	"io"
	"testing"

	"github.com/stretchr/testify/require"
	"ocm.software/ocm/api/utils/mime"

	"github.com/kyma-project/modulectl/internal/service/componentdescriptor/resources/accesshandler"
)

func TestYaml_GenerateBlobAccess(t *testing.T) {
	t.Run("should generate blob access successfully", func(t *testing.T) {
		// given
		yamlContent := "key: value"
		yaml := accesshandler.NewYaml(yamlContent)

		// when
		blobAccess, err := yaml.GenerateBlobAccess()

		// then
		require.NoError(t, err)
		require.NotNil(t, blobAccess)
		require.Equal(t, mime.MIME_YAML, blobAccess.MimeType())

		reader, err := blobAccess.Reader()
		require.NoError(t, err)
		readerContent, err := io.ReadAll(reader)
		require.NoError(t, err)
		require.YAMLEq(t, yamlContent, string(readerContent))
	})
}
