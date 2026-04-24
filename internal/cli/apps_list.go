package cli

import (
	"fmt"

	"github.com/dinacomputer/cli/internal/api"
	"github.com/spf13/cobra"
)

var appsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all applications",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := api.NewClient()
		if err != nil {
			return err
		}
		Infoln("Fetching apps...")
		apps, err := client.ListApps()
		if err != nil {
			return err
		}
		if len(apps) == 0 {
			fmt.Println("No apps found.")
			return nil
		}
		for _, entry := range apps {
			a := entry.App
			fmt.Printf("%-20s  %-40s  %s\n", a.Name, a.URL, a.CreatedAt.Format("2006-01-02 15:04"))
		}
		return nil
	},
}

func init() {
	appsCmd.AddCommand(appsListCmd)
}
