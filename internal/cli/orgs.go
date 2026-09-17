package cli

import (
	"fmt"
	"strings"

	"github.com/dinacomputer/cli/internal/api"
)

// resolveOrgID turns an --org flag value (id or name) into an organization id.
// With an empty flag it returns the caller's organization when they have
// exactly one, erroring otherwise so the command never guesses.
func resolveOrgID(client *api.Client, org string) (string, error) {
	orgs, err := client.ListOrganizations()
	if err != nil {
		return "", err
	}
	if len(orgs) == 0 {
		return "", fmt.Errorf("no organizations found for this account")
	}

	if org != "" {
		for _, o := range orgs {
			if o.ID == org || strings.EqualFold(o.Name, org) {
				return o.ID, nil
			}
		}
		names := make([]string, len(orgs))
		for i, o := range orgs {
			names[i] = o.Name
		}
		return "", fmt.Errorf("no organization matching %q — yours: %s", org, strings.Join(names, ", "))
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
