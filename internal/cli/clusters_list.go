package cli

import (
	"fmt"

	"github.com/dinacomputer/cli/internal/api"
	"github.com/spf13/cobra"
)

var clustersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List Kubernetes clusters",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := ValidateOutput(); err != nil {
			return err
		}
		client, err := api.NewClient()
		if err != nil {
			return err
		}
		Infoln("Fetching clusters...")
		clusters, err := client.ListClusters()
		if err != nil {
			return err
		}
		if JSONOutput() {
			return writeJSON(clusters)
		}
		if len(clusters) == 0 {
			fmt.Println("No clusters found.")
			return nil
		}
		for _, c := range clusters {
			fmt.Printf("%-20s  %-16s  %-12s  %s\n", c.Name, c.Region, c.Status, c.KubernetesVersion)
		}
		return nil
	},
}

func init() {
	clustersCmd.AddCommand(clustersListCmd)
}
