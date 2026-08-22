package cli

import (
	"fmt"
	"strings"

	"github.com/dinacomputer/cli/internal/api"
	"github.com/spf13/cobra"
)

// mappingFederation is the --federation flag shared by mapping subcommands.
var mappingFederation string

var identityMappingListCmd = &cobra.Command{
	Use:   "list",
	Short: "List mappings in a federation",
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
		Infoln("Fetching mappings...")
		mappings, err := client.ListMappings(orgID, mappingFederation)
		if err != nil {
			return err
		}
		if JSONOutput() {
			return writeJSON(mappings)
		}
		if len(mappings) == 0 {
			fmt.Println("No mappings found.")
			return nil
		}
		for _, m := range mappings {
			fmt.Printf("%-24s  %-16s  %-32s  %s\n", m.ID, m.Name, formatClaims(m.MatchClaims), strings.Join(m.Scopes, ","))
		}
		return nil
	},
}

var (
	mappingCreateName   string
	mappingCreateMatch  []string
	mappingCreateScopes []string
)

var identityMappingCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a claim-to-scope mapping",
	Long: `Create a mapping that grants scopes when a federation's token claims match.

All --match conditions must match (glob-aware) for the scopes to apply.`,
	Example: `  dina identity mapping create --federation fed_123 \
    --name deploy-main \
    --match repository=dinacomputer/cli \
    --match ref=refs/heads/main \
    --scope deploy`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := ValidateOutput(); err != nil {
			return err
		}
		claims, err := parseClaims(mappingCreateMatch)
		if err != nil {
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
		m, err := client.CreateMapping(orgID, mappingFederation, api.CreateMappingInput{
			Name:        mappingCreateName,
			MatchClaims: claims,
			Scopes:      mappingCreateScopes,
		})
		if err != nil {
			return err
		}
		if JSONOutput() {
			return writeJSON(m)
		}
		fmt.Printf("Mapping created: %s (%s)\n", m.Name, m.ID)
		return nil
	},
}

var mappingDeleteForce bool

var identityMappingDeleteCmd = &cobra.Command{
	Use:   "delete <mapping-id>",
	Short: "Delete a mapping",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := confirmYesNo(fmt.Sprintf("Delete mapping %q?", args[0]), mappingDeleteForce); err != nil {
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
		if err := client.DeleteMapping(orgID, mappingFederation, args[0]); err != nil {
			return err
		}
		Infof("Mapping %q deleted.\n", args[0])
		return nil
	},
}

// parseClaims turns repeated key=value flags into a claim→glob map.
func parseClaims(pairs []string) (map[string]string, error) {
	if len(pairs) == 0 {
		return nil, nil
	}
	claims := make(map[string]string, len(pairs))
	for _, p := range pairs {
		k, v, ok := strings.Cut(p, "=")
		if !ok || k == "" {
			return nil, fmt.Errorf("invalid --match %q: expected claim=value", p)
		}
		claims[k] = v
	}
	return claims, nil
}

// formatClaims renders a claim map compactly for list output.
func formatClaims(claims map[string]string) string {
	if len(claims) == 0 {
		return "-"
	}
	parts := make([]string, 0, len(claims))
	for k, v := range claims {
		parts = append(parts, k+"="+v)
	}
	return strings.Join(parts, ",")
}

func init() {
	identityMappingListCmd.Flags().StringVar(&mappingFederation, "federation", "", "Federation id")
	identityMappingListCmd.MarkFlagRequired("federation")

	identityMappingCreateCmd.Flags().StringVar(&mappingFederation, "federation", "", "Federation id")
	identityMappingCreateCmd.Flags().StringVar(&mappingCreateName, "name", "", "Optional label for the mapping")
	identityMappingCreateCmd.Flags().StringArrayVar(&mappingCreateMatch, "match", nil, "Claim condition claim=value (repeatable; all must match)")
	identityMappingCreateCmd.Flags().StringArrayVar(&mappingCreateScopes, "scope", nil, "Scope to grant (repeatable)")
	identityMappingCreateCmd.MarkFlagRequired("federation")

	identityMappingDeleteCmd.Flags().StringVar(&mappingFederation, "federation", "", "Federation id")
	identityMappingDeleteCmd.Flags().BoolVarP(&mappingDeleteForce, "force", "f", false, "Skip confirmation prompt")
	identityMappingDeleteCmd.MarkFlagRequired("federation")

	identityMappingCmd.AddCommand(identityMappingListCmd)
	identityMappingCmd.AddCommand(identityMappingCreateCmd)
	identityMappingCmd.AddCommand(identityMappingDeleteCmd)
}
