package module

import (
	"fmt"

	"github.com/spf13/cobra"

	commonerrors "github.com/kyma-project/modulectl/internal/common/errors"
	"github.com/kyma-project/modulectl/internal/service/module"
	iotools "github.com/kyma-project/modulectl/tools/io"

	_ "embed"
)

//go:embed use.txt
var use string

//go:embed short.txt
var short string

//go:embed long.txt
var long string

//go:embed example.txt
var example string

type ModuleService interface {
	CreateModule(opts module.Options) error
}

func NewCmd(moduleService ModuleService) (*cobra.Command, error) {
	if moduleService == nil {
		return nil, fmt.Errorf("%w: moduleService must not be nil", commonerrors.ErrInvalidArg)
	}

	opts := module.Options{}

	cmd := &cobra.Command{
		Use:     use,
		Short:   short,
		Long:    long,
		Example: example,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return moduleService.CreateModule(opts)
		},
	}

	opts.Out = iotools.NewDefaultOut(cmd.OutOrStdout())
	parseFlags(cmd.Flags(), &opts)

	return cmd, nil
}
