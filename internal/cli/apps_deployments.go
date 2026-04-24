package cli

import (
	"fmt"

	"github.com/dinacomputer/cli/internal/api"
	"github.com/spf13/cobra"
)

var appsDeploymentsApp string

var appsDeploymentsCmd = &cobra.Command{
	Use:   "deployments",
	Short: "List deployments for an application",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := ValidateOutput(); err != nil {
			return err
		}
		client, err := api.NewClient()
		if err != nil {
			return err
		}
		Infof("Fetching deployments for %s...\n", appsDeploymentsApp)
		deps, err := client.ListDeployments(appsDeploymentsApp)
		if err != nil {
			return err
		}
		if JSONOutput() {
			return writeJSON(deps)
		}
		if len(deps) == 0 {
			fmt.Println("No deployments found.")
			return nil
		}
		for _, d := range deps {
			fmt.Printf("%s  %-10s  %-12s  %s  port=%d  replicas=%d\n",
				d.ID, d.Status, d.SourceType, d.CreatedAt.Format("2006-01-02 15:04"), d.Port, d.Replicas)
		}
		return nil
	},
}

// --- deployment logs subcommand ---

var (
	deploymentLogsApp string
	deploymentLogsID  string
)

var deploymentLogsCmd = &cobra.Command{
	Use:   "logs",
	Short: "View build logs for a deployment",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := api.NewClient()
		if err != nil {
			return err
		}
		Infof("Fetching build logs for %s...\n", deploymentLogsID)
		logs, err := client.GetDeploymentLogs(deploymentLogsApp, deploymentLogsID)
		if err != nil {
			return err
		}
		fmt.Print(logs)
		return nil
	},
}

func init() {
	appsDeploymentsCmd.Flags().StringVarP(&appsDeploymentsApp, "app", "a", "", "Application name")
	appsDeploymentsCmd.MarkFlagRequired("app")
	appsCmd.AddCommand(appsDeploymentsCmd)

	deploymentLogsCmd.Flags().StringVarP(&deploymentLogsApp, "app", "a", "", "Application name")
	deploymentLogsCmd.Flags().StringVar(&deploymentLogsID, "id", "", "Deployment ID")
	deploymentLogsCmd.MarkFlagRequired("app")
	deploymentLogsCmd.MarkFlagRequired("id")
	appsDeploymentsCmd.AddCommand(deploymentLogsCmd)
}
