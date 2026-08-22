package cli

import (
	"fmt"
	"strings"

	"github.com/dinacomputer/cli/internal/api"
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

// resolveOrgID turns the --org flag (id or name) into an organization id. With
// no flag it returns the caller's organization when they have exactly one,
// erroring otherwise so the command never guesses.
func resolveOrgID(client *api.Client) (string, error) {
	orgs, err := client.ListOrganizations()
	if err != nil {
		return "", err
	}
	if len(orgs) == 0 {
		return "", fmt.Errorf("no organizations found for this account")
	}

	if identityOrg != "" {
		for _, o := range orgs {
			if o.ID == identityOrg || strings.EqualFold(o.Name, identityOrg) {
				return o.ID, nil
			}
		}
		return "", fmt.Errorf("no organization matching %q — list them with: dina identity federation list", identityOrg)
	}

	if len(orgs) == 1 {
		return orgs[0].ID, nil
	}

	names := make([]string, len(orgs))
	for i, o := range orgs {
		names[i] = o.Name
	}
	return "", fmt.Errorf("multiple organizations — pass --org with one of: %s", strings.Join(names, ", "))
}

func init() {
	identityCmd.PersistentFlags().StringVar(&identityOrg, "org", "", "Organization id or name (defaults to your only organization)")
	identityCmd.AddCommand(identityFederationCmd)
	identityCmd.AddCommand(identityMappingCmd)
	rootCmd.AddCommand(identityCmd)
}
