package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/dinacomputer/cli/internal/api"
	"github.com/dinacomputer/cli/internal/config"
	"github.com/spf13/cobra"
)

// orgEnvVar overrides the active organization for a single shell or CI job.
const orgEnvVar = "DINA_ORG"

const (
	orgSourceFlag   = "flag"
	orgSourceEnv    = "env"
	orgSourceActive = "active"
)

// orgSelection returns the organization the caller asked for and where it came
// from: --org, then DINA_ORG, then the active organization from `dina orgs use`.
// An empty value means nothing was selected.
func orgSelection(flag string) (value, source string) {
	if flag != "" {
		return flag, orgSourceFlag
	}
	if v := os.Getenv(orgEnvVar); v != "" {
		return v, orgSourceEnv
	}
	if cfg, err := config.Load(); err == nil && cfg.ActiveOrg != nil {
		return cfg.ActiveOrg.ID, orgSourceActive
	}
	return "", ""
}

// resolveOrgID turns an --org flag value (id or name) into an organization id.
// With an empty flag it falls back to DINA_ORG, then the active organization,
// then the caller's organization when they have exactly one, erroring
// otherwise so the command never guesses.
func resolveOrgID(client *api.Client, flag string) (string, error) {
	orgs, err := client.ListOrganizations()
	if err != nil {
		return "", err
	}
	if len(orgs) == 0 {
		return "", fmt.Errorf("no organizations found for this account")
	}

	org, source := orgSelection(flag)
	if org != "" {
		if o := matchOrg(orgs, org); o != nil {
			return o.ID, nil
		}
		switch source {
		case orgSourceEnv:
			return "", fmt.Errorf("%s=%q matches no organization — yours: %s", orgEnvVar, org, orgNames(orgs))
		case orgSourceActive:
			return "", fmt.Errorf("the active organization is no longer available — run `dina orgs use` with one of: %s", orgNames(orgs))
		}
		return "", fmt.Errorf("no organization matching %q — yours: %s", org, orgNames(orgs))
	}

	if len(orgs) == 1 {
		return orgs[0].ID, nil
	}
	return "", fmt.Errorf("multiple organizations — pass --org or run `dina orgs use` with one of: %s", orgNames(orgs))
}

func matchOrg(orgs []api.Organization, org string) *api.Organization {
	for i, o := range orgs {
		if o.ID == org || strings.EqualFold(o.Name, org) {
			return &orgs[i]
		}
	}
	return nil
}

func orgNames(orgs []api.Organization) string {
	names := make([]string, len(orgs))
	for i, o := range orgs {
		names[i] = o.Name
	}
	return strings.Join(names, ", ")
}

const orgFlagUsage = "Organization id or name (defaults to the active organization, or your only one)"

var orgsCmd = &cobra.Command{
	Use:     "orgs",
	Aliases: []string{"org", "organizations"},
	Short:   "List organizations and choose the active one",
	Long: `List your organizations and choose which one commands act on.

Commands that take --org pick the organization in this order: --org, the
` + orgEnvVar + ` environment variable, the active organization set with
` + "`dina orgs use`" + `, and finally your only organization if you belong to exactly one.`,
	GroupID: groupAccount,
}

var orgsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List your organizations",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := ValidateOutput(); err != nil {
			return err
		}
		client, err := api.NewClient()
		if err != nil {
			return err
		}
		orgs, err := client.ListOrganizations()
		if err != nil {
			return err
		}
		active := ""
		if cfg, err := config.Load(); err == nil && cfg.ActiveOrg != nil {
			active = cfg.ActiveOrg.ID
		}

		if JSONOutput() {
			type orgEntry struct {
				api.Organization
				Active bool `json:"active"`
			}
			out := make([]orgEntry, len(orgs))
			for i, o := range orgs {
				out[i] = orgEntry{Organization: o, Active: o.ID == active}
			}
			return writeJSON(out)
		}
		if len(orgs) == 0 {
			fmt.Println("No organizations found.")
			return nil
		}
		for _, o := range orgs {
			marker := " "
			if o.ID == active {
				marker = "*"
			}
			fmt.Printf("%s %-30s  %s\n", marker, o.Name, o.ID)
		}
		return nil
	},
}

var orgsUseCmd = &cobra.Command{
	Use:     "use <org>",
	Aliases: []string{"switch"},
	Short:   "Set the active organization",
	Long:    "Set the organization commands act on when --org is not passed. Accepts an id or name.",
	Example: `  dina orgs use acme`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := api.NewClient()
		if err != nil {
			return err
		}
		orgs, err := client.ListOrganizations()
		if err != nil {
			return err
		}
		o := matchOrg(orgs, args[0])
		if o == nil {
			return fmt.Errorf("no organization matching %q — yours: %s", args[0], orgNames(orgs))
		}
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		cfg.ActiveOrg = &config.ActiveOrg{ID: o.ID, Name: o.Name}
		if err := config.Save(cfg); err != nil {
			return err
		}
		fmt.Printf("Active organization: %s\n", o.Name)
		return nil
	},
}

var orgsCurrentCmd = &cobra.Command{
	Use:   "current",
	Short: "Show the organization commands act on",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := ValidateOutput(); err != nil {
			return err
		}
		client, err := api.NewClient()
		if err != nil {
			return err
		}
		id, err := resolveOrgID(client, "")
		if err != nil {
			return err
		}
		orgs, err := client.ListOrganizations()
		if err != nil {
			return err
		}
		o := matchOrg(orgs, id)
		_, source := orgSelection("")
		if source == "" {
			source = "only"
		}

		if JSONOutput() {
			return writeJSON(struct {
				api.Organization
				Source string `json:"source"`
			}{*o, source})
		}
		switch source {
		case orgSourceEnv:
			fmt.Printf("%s (from %s)\n", o.Name, orgEnvVar)
		case "only":
			fmt.Printf("%s (your only organization)\n", o.Name)
		default:
			fmt.Println(o.Name)
		}
		return nil
	},
}

var orgsUnsetCmd = &cobra.Command{
	Use:   "unset",
	Short: "Clear the active organization",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := clearActiveOrg(); err != nil {
			return err
		}
		Infoln("Active organization cleared.")
		return nil
	},
}

func clearActiveOrg() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if cfg.ActiveOrg == nil {
		return nil
	}
	cfg.ActiveOrg = nil
	return config.Save(cfg)
}

func init() {
	orgsCmd.AddCommand(orgsListCmd)
	orgsCmd.AddCommand(orgsUseCmd)
	orgsCmd.AddCommand(orgsCurrentCmd)
	orgsCmd.AddCommand(orgsUnsetCmd)
	rootCmd.AddCommand(orgsCmd)
}
