package cli

import (
	"fmt"

	"github.com/dinacomputer/cli/internal/api"
	"github.com/spf13/cobra"
)

var appsCreateOrg string

var appsCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new application",
	Long: `Create a new application on the Dina platform. The app's URL is assigned
automatically based on its name.

The app belongs to an organization: --org, the active organization
(` + "`dina orgs use`" + `), or yours if you belong to exactly one. CI jobs authenticated with ` + "`dina auth login --federated`" + ` can deploy the
organization's apps.`,
	Example: `  dina apps create my-app
  dina apps create my-app --org acme`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := api.NewClient()
		if err != nil {
			return err
		}
		input := api.CreateAppInput{Name: args[0]}
		if org, _ := orgSelection(appsCreateOrg); org != "" {
			if input.OrgID, err = resolveOrgID(client, appsCreateOrg); err != nil {
				return err
			}
		}
		app, err := client.CreateApp(input)
		if err != nil {
			return err
		}
		fmt.Printf("App created: %s\n  URL: %s\n", app.Name, app.URL)
		return nil
	},
}

func init() {
	appsCreateCmd.Flags().StringVar(&appsCreateOrg, "org", "", orgFlagUsage)
	appsCmd.AddCommand(appsCreateCmd)
}
