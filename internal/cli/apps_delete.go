package cli

import (
	"fmt"

	"github.com/dinacomputer/cli/internal/api"
	"github.com/spf13/cobra"
)

var appsDeleteApp string

var appsDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an application",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := api.NewClient()
		if err != nil {
			return err
		}
		if err := client.DeleteApp(appsDeleteApp); err != nil {
			return err
		}
		fmt.Printf("App %q deleted.\n", appsDeleteApp)
		return nil
	},
}

func init() {
	appsDeleteCmd.Flags().StringVarP(&appsDeleteApp, "app", "a", "", "Application name")
	appsDeleteCmd.MarkFlagRequired("app")
	appsCmd.AddCommand(appsDeleteCmd)
}
