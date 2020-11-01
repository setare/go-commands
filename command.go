package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use: os.Args[0],
}

// Execute starts the microservice based on its configuration.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(EC_ROOTCMD_FAILED)
	}
}

// AddCommand is a helper for `Rootcmd.AddCommand`.
func AddCommand(cmds ...*cobra.Command) {
	RootCmd.AddCommand(cmds...)
}
