package cli

import (
	"fmt"
	"os"

	"github.com/dinacomputer/cli/internal/api"
	"github.com/spf13/cobra"
)

var (
	appsDeleteApp   string
	appsDeleteForce bool
)

var appsDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an application",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := confirmByName("app", appsDeleteApp, appsDeleteForce); err != nil {
			return err
		}
		client, err := api.NewClient()
		if err != nil {
			return err
		}
		if err := client.DeleteApp(appsDeleteApp); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "App %q deleted.\n", appsDeleteApp)
		return nil
	},
}

func init() {
	appsDeleteCmd.Flags().StringVarP(&appsDeleteApp, "app", "a", "", "Application name")
	appsDeleteCmd.Flags().BoolVarP(&appsDeleteForce, "force", "f", false, "Skip confirmation prompt")
	appsDeleteCmd.MarkFlagRequired("app")
	appsCmd.AddCommand(appsDeleteCmd)
}
