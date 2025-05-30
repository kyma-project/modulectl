package accesshandler

import (
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/utils/blobaccess"
	"ocm.software/ocm/api/utils/mime"
)

type Yaml struct {
	String string
}

func NewYaml(stringData string) *Yaml {
	return &Yaml{
		String: stringData,
	}
}

func (y *Yaml) GenerateBlobAccess() (cpi.BlobAccess, error) {
	return blobaccess.ForString(mime.MIME_YAML, y.String), nil
}
