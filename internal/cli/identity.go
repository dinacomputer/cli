package cli

import (
	"github.com/spf13/cobra"
)

// identityOrg is the --org flag shared across identity subcommands. It accepts
// an organization id or name; when empty the caller's sole organization is used.
var identityOrg string

var identityCmd = &cobra.Command{
	Use:   "identity",
	Short: "Manage workload identity federation",
	Long: `Manage keyless CI authentication: register external OIDC issuers (federations)
and map their token claims to Dina scopes (mappings).

A federation trusts a CI provider's OIDC issuer; a mapping under it binds
specific token claims (repository, ref, environment, …) to the scopes a job
receives. CI then authenticates with ` + "`dina auth login --federated`" + ` — no stored
secret.`,
	GroupID: groupAccount,
}

var identityFederationCmd = &cobra.Command{
	Use:     "federation",
	Aliases: []string{"federations", "fed"},
	Short:   "Manage OIDC federations (trusted CI issuers)",
}

var identityMappingCmd = &cobra.Command{
	Use:     "mapping",
	Aliases: []string{"mappings", "map"},
	Short:   "Manage claim-to-scope mappings within a federation",
}

func init() {
	identityCmd.PersistentFlags().StringVar(&identityOrg, "org", "", "Organization id or name (defaults to your only organization)")
	identityCmd.AddCommand(identityFederationCmd)
	identityCmd.AddCommand(identityMappingCmd)
	rootCmd.AddCommand(identityCmd)
}
