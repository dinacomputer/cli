package cli

import (
	"fmt"
	"strings"

	"github.com/dinacomputer/cli/internal/api"
	"github.com/spf13/cobra"
)

var appsEnvCmd = &cobra.Command{
	Use:   "env",
	Short: "Manage environment variables for an application",
}

var appsEnvSetApp string

var appsEnvSetCmd = &cobra.Command{
	Use:   "set KEY=VALUE [KEY=VALUE ...]",
	Short: "Set environment variables",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		vars := make(map[string]string, len(args))
		for _, arg := range args {
			k, v, ok := strings.Cut(arg, "=")
			if !ok {
				return fmt.Errorf("invalid format %q, expected KEY=VALUE", arg)
			}
			vars[k] = v
		}

		client, err := api.NewClient()
		if err != nil {
			return err
		}
		result, err := client.SetEnvVars(appsEnvSetApp, vars)
		if err != nil {
			return err
		}
		for _, ev := range result {
			fmt.Printf("%s=%s\n", ev.Key, ev.Value)
		}
		return nil
	},
}

func init() {
	appsEnvSetCmd.Flags().StringVarP(&appsEnvSetApp, "app", "a", "", "Application name")
	appsEnvSetCmd.MarkFlagRequired("app")
	appsEnvCmd.AddCommand(appsEnvSetCmd)
	appsCmd.AddCommand(appsEnvCmd)
}
