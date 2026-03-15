package cli

import (
	"fmt"

	"github.com/dinacomputer/cli/internal/api"
	"github.com/spf13/cobra"
)

var (
	appsUpdateApp  string
	appsUpdateName string
)

var appsUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update an application",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := api.NewClient()
		if err != nil {
			return err
		}
		input := api.UpdateAppInput{}
		if cmd.Flags().Changed("name") {
			input.Name = appsUpdateName
		}
		app, err := client.UpdateApp(appsUpdateApp, input)
		if err != nil {
			return err
		}
		fmt.Printf("App updated: %s (id: %s)\n", app.Name, app.ID)
		return nil
	},
}

func init() {
	appsUpdateCmd.Flags().StringVarP(&appsUpdateApp, "app", "a", "", "Application name")
	appsUpdateCmd.Flags().StringVar(&appsUpdateName, "name", "", "New application name")
	appsUpdateCmd.MarkFlagRequired("app")
	appsCmd.AddCommand(appsUpdateCmd)
}
