package modulectl

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kyma-project/modulectl/cmd/modulectl/create"

	_ "embed"
)

//go:embed use.txt
var use string

//go:embed short.txt
var short string

//go:embed long.txt
var long string

func NewCmd() (*cobra.Command, error) {
	rootCmd := &cobra.Command{
		Use:   use,
		Short: short,
		Long:  long,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("This is the Kyma ModuleCtl command executed")
		},
	}

	cmd, err := create.NewCmd()
	if err != nil {
		return nil, fmt.Errorf("failed to build create command: %w", err)
	}

	rootCmd.AddCommand(cmd)

	return rootCmd, nil
}
