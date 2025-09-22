package version

import (
	"github.com/spf13/cobra"
)

const (
	use   = "version"
	short = "Prints the current modulectl version."
	long  = "This command prints the current semantic version of the modulectl binary set at build time."
)

// Version will contain the binary version injected by make build target.
var Version string //nolint:gochecknoglobals // This is a variable meant to be set at build time

func NewCmd() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:     use,
		Short:   short,
		Long:    long,
		Args:    cobra.NoArgs,
		Aliases: []string{"v"},
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println(Version)
		},
	}

	return cmd, nil
}
