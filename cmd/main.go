package main

import (
	"fmt"
	"os"

	"github.com/kyma-project/modulectl/cmd/modulectl"
)

func main() {
	cmd, err := modulectl.NewCmd()
	if err != nil {
		fmt.Println(fmt.Errorf("failed to build modulectl command: %w", err))
		os.Exit(-1)
	}

	if err = cmd.Execute(); err != nil {
		fmt.Println(fmt.Errorf("failed to execute modulectl command: %w", err))
		os.Exit(-1)
	}
}
