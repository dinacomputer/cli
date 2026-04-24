package cli

import (
	"fmt"
	"os"

	"github.com/dinacomputer/cli/internal/api"
	"github.com/spf13/cobra"
)

var appsHostnamesCmd = &cobra.Command{
	Use:   "hostnames",
	Short: "Manage custom hostnames for an application",
}

var appsHostnamesAddApp string

var appsHostnamesAddCmd = &cobra.Command{
	Use:   "add <hostname>",
	Short: "Add a custom hostname",
	Long: `Attach a custom hostname to an application. You are responsible for pointing
the hostname's DNS record (CNAME or A) at the app's Dina URL.`,
	Example: `  dina apps hostnames add -a my-app api.example.com`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := api.NewClient()
		if err != nil {
			return err
		}
		h, err := client.AddHostname(appsHostnamesAddApp, args[0])
		if err != nil {
			return err
		}
		fmt.Printf("Hostname %s added (id: %s)\n", h.Hostname, h.ID)
		return nil
	},
}

var (
	appsHostnamesRemoveApp   string
	appsHostnamesRemoveForce bool
)

var appsHostnamesRemoveCmd = &cobra.Command{
	Use:   "remove <hostname>",
	Short: "Remove a custom hostname",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		prompt := fmt.Sprintf("Remove hostname %s from %s?", args[0], appsHostnamesRemoveApp)
		if err := confirmYesNo(prompt, appsHostnamesRemoveForce); err != nil {
			return err
		}
		client, err := api.NewClient()
		if err != nil {
			return err
		}
		if err := client.RemoveHostname(appsHostnamesRemoveApp, args[0]); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "Hostname %s removed.\n", args[0])
		return nil
	},
}

func init() {
	appsHostnamesAddCmd.Flags().StringVarP(&appsHostnamesAddApp, "app", "a", "", "Application name")
	appsHostnamesAddCmd.MarkFlagRequired("app")
	appsHostnamesCmd.AddCommand(appsHostnamesAddCmd)

	appsHostnamesRemoveCmd.Flags().StringVarP(&appsHostnamesRemoveApp, "app", "a", "", "Application name")
	appsHostnamesRemoveCmd.Flags().BoolVarP(&appsHostnamesRemoveForce, "force", "f", false, "Skip confirmation prompt")
	appsHostnamesRemoveCmd.MarkFlagRequired("app")
	appsHostnamesCmd.AddCommand(appsHostnamesRemoveCmd)

	appsCmd.AddCommand(appsHostnamesCmd)
}
