package componentdescriptor

import (
	"fmt"

	commonerrors "github.com/kyma-project/modulectl/internal/common/errors"
)

type ModuleDefinition struct {
	Name          string
	Version       string
	SchemaVersion string

	// TODO: Define Layer struct based on the new ocm and remember to add the raw-manifest layer
	// Layers []Layer

}

func (d *ModuleDefinition) ValidateModuleDefinition() error {
	if d.Name == "" {
		return fmt.Errorf("%w: module name must not be empty", commonerrors.ErrInvalidArg)
	}

	if d.Version == "" {
		return fmt.Errorf("%w: module version must not be empty", commonerrors.ErrInvalidArg)
	}

	if d.SchemaVersion == "" {
		return fmt.Errorf("%w: module schema version must not be empty", commonerrors.ErrInvalidArg)
	}

	return nil
}
