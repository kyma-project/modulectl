package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

func main() {
	var rootCommand = &cobra.Command{
		Use:   "kyma",
		Short: "This is the Kyma Module Controller CLI",
		Long:  "A CLI from the Kyma Module Controller. Wonderful to use.",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("This is the Kyma ModuleCtl command executed")
		},
	}

	if err := rootCommand.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
