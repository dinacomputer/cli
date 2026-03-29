package cli

import (
	"github.com/spf13/cobra"
)

var appsCmd = &cobra.Command{
	Use:   "apps",
	Short: "Manage applications",
	Long:  "Create, list, inspect, update, and delete applications. Also manage logs, deployments, environment variables, and custom hostnames.",
}

func init() {
	rootCmd.AddCommand(appsCmd)
}
