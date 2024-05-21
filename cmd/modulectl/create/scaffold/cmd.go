package scaffold

import (
	_ "embed"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kyma-project/modulectl/internal/common/errors"
	"github.com/kyma-project/modulectl/internal/service/scaffold"
	"github.com/kyma-project/modulectl/tools/io"
)

//go:embed use.txt
var use string

//go:embed short.txt
var short string

//go:embed long.txt
var long string

//go:embed example.txt
var example string

type ScaffoldService interface {
	CreateScaffold(opts scaffold.Options) error
}

func NewCmd(scaffoldService ScaffoldService) (*cobra.Command, error) {
	if scaffoldService == nil {
		return nil, fmt.Errorf("%w: scaffoldService must not be nil", errors.ErrInvalidArg)
	}

	opts := scaffold.Options{}

	cmd := &cobra.Command{
		Use:     use,
		Short:   short,
		Long:    long,
		Example: example,
		Args:    cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return scaffoldService.CreateScaffold(opts)
		},
	}

	opts.Out = io.NewDefaultOut(cmd.OutOrStdout())
	parseFlags(cmd.Flags(), &opts)

	return cmd, nil
}
