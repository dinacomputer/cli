package cli

import (
	"github.com/spf13/cobra"
)

var clustersCmd = &cobra.Command{
	Use:     "clusters",
	Short:   "List and connect to Kubernetes clusters",
	Long:    "List the Kubernetes clusters you can access and connect to them by writing a kubectl context that authenticates through the Dina CLI.",
	GroupID: groupDeploy,
}

func init() {
	rootCmd.AddCommand(clustersCmd)
}
