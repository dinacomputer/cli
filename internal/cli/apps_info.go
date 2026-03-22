package cli

import (
	"fmt"

	"github.com/dinacomputer/cli/internal/api"
	"github.com/spf13/cobra"
)

var appsInfoApp string

var appsInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show application details",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := api.NewClient()
		if err != nil {
			return err
		}
		out, err := client.GetApp(appsInfoApp)
		if err != nil {
			return err
		}

		a := out.App
		fmt.Printf("Name:       %s\n", a.Name)
		fmt.Printf("ID:         %s\n", a.ID)
		fmt.Printf("URL:        %s\n", a.URL)
		fmt.Printf("Namespace:  %s\n", a.Namespace)
		fmt.Printf("Owner:      %s\n", a.OwnerID)
		fmt.Printf("Created:    %s\n", a.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("Updated:    %s\n", a.UpdatedAt.Format("2006-01-02 15:04:05"))

		if len(a.Hostnames) > 0 {
			fmt.Println("Hostnames:")
			for _, h := range a.Hostnames {
				fmt.Printf("  - %s\n", h.Hostname)
			}
		}

		if out.LatestDeployment != nil {
			d := out.LatestDeployment
			fmt.Printf("\nLatest deployment:\n")
			fmt.Printf("  ID:       %s\n", d.ID)
			fmt.Printf("  Status:   %s\n", d.Status)
			fmt.Printf("  Port:     %d\n", d.Port)
			fmt.Printf("  Replicas: %d\n", d.Replicas)
			if d.Image != "" {
				fmt.Printf("  Image:    %s\n", d.Image)
			}
		}
		return nil
	},
}

func init() {
	appsInfoCmd.Flags().StringVarP(&appsInfoApp, "app", "a", "", "Application name")
	appsInfoCmd.MarkFlagRequired("app")
	appsCmd.AddCommand(appsInfoCmd)
}
