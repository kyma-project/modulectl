package contentprovider

import "github.com/kyma-project/modulectl/internal/common/types"

type DefaultCR struct{}

func NewDefaultCR() *DefaultCR {
	return &DefaultCR{}
}

func (s *DefaultCR) GetDefaultContent(_ types.KeyValueArgs) (string, error) {
	return `# This is the file that contains the defaultCR for your create, which is the Custom Resource that will be created upon create enablement.
# Make sure this file contains *ONLY* the Custom Resource (not the Custom Resource Definition, which should be a part of your create manifest)

`, nil
}
