package cli

import (
	"fmt"

	"github.com/dinacomputer/cli/internal/api"
	"github.com/spf13/cobra"
)

var appsCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new application",
	Long: `Create a new application on the Dina platform. The app's URL is assigned
automatically based on its name.`,
	Example: `  dina apps create my-app`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := api.NewClient()
		if err != nil {
			return err
		}
		app, err := client.CreateApp(args[0])
		if err != nil {
			return err
		}
		fmt.Printf("App created: %s\n  URL: %s\n", app.Name, app.URL)
		return nil
	},
}

func init() {
	appsCmd.AddCommand(appsCreateCmd)
}
