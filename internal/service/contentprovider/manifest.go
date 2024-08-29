package contentprovider

import (
	"github.com/kyma-project/modulectl/internal/common/types"
)

type Manifest struct{}

func NewManifest() *Manifest {
	return &Manifest{}
}

func (s *Manifest) GetDefaultContent(_ types.KeyValueArgs) (string, error) {
	return `# This file holds the Manifest of your create, encompassing all resources installed in the cluster once the create is activated.
# It should include the Custom Resource Definition for your create's default CustomResource, if it exists.

`, nil
}
