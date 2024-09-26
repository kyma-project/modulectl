package scaffold

import (
	"fmt"

	"github.com/spf13/cobra"

	commonerrors "github.com/kyma-project/modulectl/internal/common/errors"
	"github.com/kyma-project/modulectl/internal/service/scaffold"
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

type Service interface {
	Run(opts scaffold.Options) error
}

func NewCmd(service Service) (*cobra.Command, error) {
	if service == nil {
		return nil, fmt.Errorf("%w: service must not be nil", commonerrors.ErrInvalidArg)
	}

	opts := scaffold.Options{}

	cmd := &cobra.Command{
		Use:     use,
		Short:   short,
		Long:    long,
		Example: example,
		Args:    cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return service.Run(opts)
		},
	}

	opts.Out = iotools.NewDefaultOut(cmd.OutOrStdout())
	parseFlags(cmd.Flags(), &opts)

	return cmd, nil
}
