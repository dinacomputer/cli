package cli

import (
	"fmt"
	"os"

	"github.com/dinacomputer/cli/internal/api"
	"github.com/spf13/cobra"
)

var (
	appsLogsApp   string
	appsLogsLines int
)

var appsLogsCmd = &cobra.Command{
	Use:   "logs",
	Short: "View application logs",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := api.NewClient()
		if err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "Fetching logs for %s...\n", appsLogsApp)
		logs, err := client.GetAppLogs(appsLogsApp, appsLogsLines)
		if err != nil {
			return err
		}
		fmt.Print(logs)
		return nil
	},
}

func init() {
	appsLogsCmd.Flags().StringVarP(&appsLogsApp, "app", "a", "", "Application name")
	appsLogsCmd.Flags().IntVarP(&appsLogsLines, "lines", "n", 100, "Number of log lines")
	appsLogsCmd.MarkFlagRequired("app")
	appsCmd.AddCommand(appsLogsCmd)
}
