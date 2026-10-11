package cli

import (
	"fmt"

	"github.com/dinacomputer/cli/internal/api"
	"github.com/spf13/cobra"
)

var storageRegionsCmd = &cobra.Command{
	Use:     "regions",
	Aliases: []string{"region"},
	Short:   "List storage regions opened for the organization",
	Long: `List the storage regions the organization can place buckets and keys in,
besides the default one used when --region is omitted.

Regions are opened per organization by Sokkel. Pass a region's name to
dina storage buckets create --region and dina storage keys create --region.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := ValidateOutput(); err != nil {
			return err
		}
		client, err := api.NewClient()
		if err != nil {
			return err
		}
		orgID, err := resolveOrgID(client, storageOrg)
		if err != nil {
			return err
		}
		regions, err := client.ListStorageRegions(orgID)
		if err != nil {
			return err
		}
		if JSONOutput() {
			return writeJSON(regions)
		}
		if len(regions) == 0 {
			fmt.Println("No extra storage regions; buckets and keys use the default region.")
			return nil
		}
		for _, r := range regions {
			fmt.Printf("%-16s  %s\n", r.Name, r.EndpointURL)
		}
		return nil
	},
}
