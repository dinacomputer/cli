package cli

import (
	"fmt"
	"runtime"

	"github.com/dinacomputer/cli/internal/api"
	"github.com/spf13/cobra"
)

// Version is set at build time via -ldflags.
var Version = "dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the CLI version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("dina " + Version)
	},
}

func init() {
	api.UserAgent = fmt.Sprintf("dina/%s (%s/%s)", Version, runtime.GOOS, runtime.GOARCH)
	rootCmd.AddCommand(versionCmd)
}
