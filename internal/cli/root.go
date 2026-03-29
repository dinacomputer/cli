package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "dina",
	Short: "Dina CLI",
	Long:  "Dina CLI – deploy applications, manage apps and environment variables, view logs, configure custom hostnames, and install AI agent skills for the Dina platform.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
