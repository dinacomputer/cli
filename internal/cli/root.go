package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var noInput bool

var rootCmd = &cobra.Command{
	Use:   "dina",
	Short: "Dina CLI",
	Long:  "Dina CLI – deploy applications, manage apps and environment variables, view logs, configure custom hostnames, and install AI agent skills for the Dina platform.",
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&noInput, "no-input", false, "Disable interactive prompts (fail if input is required)")
}

// NoInput reports whether the --no-input flag was passed.
func NoInput() bool {
	return noInput
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
