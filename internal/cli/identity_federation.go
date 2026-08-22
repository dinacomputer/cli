package cli

import (
	"fmt"
	"strings"

	"github.com/dinacomputer/cli/internal/api"
	"github.com/spf13/cobra"
)

var identityFederationListCmd = &cobra.Command{
	Use:   "list",
	Short: "List federations",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := ValidateOutput(); err != nil {
			return err
		}
		client, err := api.NewClient()
		if err != nil {
			return err
		}
		orgID, err := resolveOrgID(client)
		if err != nil {
			return err
		}
		Infoln("Fetching federations...")
		feds, err := client.ListFederations(orgID)
		if err != nil {
			return err
		}
		if JSONOutput() {
			return writeJSON(feds)
		}
		if len(feds) == 0 {
			fmt.Println("No federations found.")
			return nil
		}
		for _, f := range feds {
			state := "enabled"
			if f.Disabled {
				state = "disabled"
			}
			fmt.Printf("%-24s  %-20s  %-40s  %s\n", f.ID, f.Name, f.Issuer, state)
		}
		return nil
	},
}

var (
	fedCreateName      string
	fedCreateIssuer    string
	fedCreateAudiences []string
	fedCreateSubject   string
)

var identityFederationCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Register a new OIDC federation",
	Long: `Register an external OIDC issuer (a CI provider) as a trusted federation.

Tokens minted by the issuer can then be exchanged for short-lived Dina
credentials. Add a mapping to grant scopes: dina identity mapping create.`,
	Example: `  dina identity federation create \
    --name github-actions \
    --issuer https://token.actions.githubusercontent.com`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := ValidateOutput(); err != nil {
			return err
		}
		client, err := api.NewClient()
		if err != nil {
			return err
		}
		orgID, err := resolveOrgID(client)
		if err != nil {
			return err
		}
		fed, err := client.CreateFederation(orgID, api.CreateFederationInput{
			Name:         fedCreateName,
			Issuer:       fedCreateIssuer,
			Audiences:    fedCreateAudiences,
			SubjectClaim: fedCreateSubject,
		})
		if err != nil {
			return err
		}
		if JSONOutput() {
			return writeJSON(fed)
		}
		fmt.Printf("Federation created: %s (%s)\n  client_id: %s\n", fed.Name, fed.ID, fed.ClientID)
		return nil
	},
}

var identityFederationGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Show a federation's details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := ValidateOutput(); err != nil {
			return err
		}
		client, err := api.NewClient()
		if err != nil {
			return err
		}
		orgID, err := resolveOrgID(client)
		if err != nil {
			return err
		}
		fed, err := client.GetFederation(orgID, args[0])
		if err != nil {
			return err
		}
		if JSONOutput() {
			return writeJSON(fed)
		}
		printFederation(fed)
		return nil
	},
}

var (
	fedUpdateName      string
	fedUpdateAudiences []string
	fedUpdateSubject   string
	fedUpdateDisabled  bool
)

var identityFederationUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a federation",
	Args:  cobra.ExactArgs(1),
	Example: `  # disable a federation without deleting it
  dina identity federation update fed_123 --disabled`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := ValidateOutput(); err != nil {
			return err
		}
		input := api.UpdateFederationInput{}
		if cmd.Flags().Changed("name") {
			input.Name = fedUpdateName
		}
		if cmd.Flags().Changed("audience") {
			input.Audiences = fedUpdateAudiences
		}
		if cmd.Flags().Changed("subject-claim") {
			input.SubjectClaim = fedUpdateSubject
		}
		if cmd.Flags().Changed("disabled") {
			input.Disabled = &fedUpdateDisabled
		}

		client, err := api.NewClient()
		if err != nil {
			return err
		}
		orgID, err := resolveOrgID(client)
		if err != nil {
			return err
		}
		fed, err := client.UpdateFederation(orgID, args[0], input)
		if err != nil {
			return err
		}
		if JSONOutput() {
			return writeJSON(fed)
		}
		fmt.Printf("Federation %q updated.\n", fed.Name)
		return nil
	},
}

var fedDeleteForce bool

var identityFederationDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a federation",
	Long: `Permanently delete a federation and its mappings.

By default you will be prompted to confirm. Pass --force to skip the prompt.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := confirmByName("federation", args[0], fedDeleteForce); err != nil {
			return err
		}
		client, err := api.NewClient()
		if err != nil {
			return err
		}
		orgID, err := resolveOrgID(client)
		if err != nil {
			return err
		}
		if err := client.DeleteFederation(orgID, args[0]); err != nil {
			return err
		}
		Infof("Federation %q deleted.\n", args[0])
		return nil
	},
}

func printFederation(f *api.Federation) {
	auds := "(server origin)"
	if len(f.Audiences) > 0 {
		auds = strings.Join(f.Audiences, ", ")
	}
	state := "enabled"
	if f.Disabled {
		state = "disabled"
	}
	fmt.Printf("ID:            %s\n", f.ID)
	fmt.Printf("Name:          %s\n", f.Name)
	fmt.Printf("Issuer:        %s\n", f.Issuer)
	fmt.Printf("Audiences:     %s\n", auds)
	fmt.Printf("Subject claim: %s\n", f.SubjectClaim)
	fmt.Printf("Client ID:     %s\n", f.ClientID)
	fmt.Printf("State:         %s\n", state)
}

func init() {
	identityFederationCreateCmd.Flags().StringVar(&fedCreateName, "name", "", "Human-readable name")
	identityFederationCreateCmd.Flags().StringVar(&fedCreateIssuer, "issuer", "", "External OIDC issuer URL")
	identityFederationCreateCmd.Flags().StringArrayVar(&fedCreateAudiences, "audience", nil, "Accepted token audience (repeatable; empty accepts the server origin)")
	identityFederationCreateCmd.Flags().StringVar(&fedCreateSubject, "subject-claim", "", "Claim identifying the principal (defaults to 'sub')")
	identityFederationCreateCmd.MarkFlagRequired("name")
	identityFederationCreateCmd.MarkFlagRequired("issuer")

	identityFederationUpdateCmd.Flags().StringVar(&fedUpdateName, "name", "", "Human-readable name")
	identityFederationUpdateCmd.Flags().StringArrayVar(&fedUpdateAudiences, "audience", nil, "Accepted token audience (repeatable; replaces the existing set)")
	identityFederationUpdateCmd.Flags().StringVar(&fedUpdateSubject, "subject-claim", "", "Claim identifying the principal")
	identityFederationUpdateCmd.Flags().BoolVar(&fedUpdateDisabled, "disabled", false, "Disable (true) or enable (false) the federation")

	identityFederationDeleteCmd.Flags().BoolVarP(&fedDeleteForce, "force", "f", false, "Skip confirmation prompt")

	identityFederationCmd.AddCommand(identityFederationListCmd)
	identityFederationCmd.AddCommand(identityFederationCreateCmd)
	identityFederationCmd.AddCommand(identityFederationGetCmd)
	identityFederationCmd.AddCommand(identityFederationUpdateCmd)
	identityFederationCmd.AddCommand(identityFederationDeleteCmd)
}
