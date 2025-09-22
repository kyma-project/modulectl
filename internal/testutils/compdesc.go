package testutils

import (
	"ocm.software/ocm/api/ocm/compdesc"

	// Make sure init is being called and v2 is registered.
	_ "ocm.software/ocm/api/ocm/compdesc/versions/v2"
)

// CreateComponentDescriptor creates a new component descriptor for testing purposes.
// This exists due to the fact that the ocm compdesc package is using init functions in /versions/v2
// that need to be called for certain functions of
// ocm/compdesc to work properly, like Convert(), Validate(), etc.
// If you got "schema not found" error, saying that v2 is unsupported,
// use this function to create a component descriptor in tests!
func CreateComponentDescriptor(name, version string) *compdesc.ComponentDescriptor {
	return compdesc.New(name, version)
}
