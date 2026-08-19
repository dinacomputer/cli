package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/dinacomputer/cli/internal/api"
	"github.com/spf13/cobra"
)

// execCredential is the client.authentication.k8s.io/v1 ExecCredential that
// kubectl expects an exec plugin to print on stdout.
type execCredential struct {
	APIVersion string               `json:"apiVersion"`
	Kind       string               `json:"kind"`
	Status     execCredentialStatus `json:"status"`
}

type execCredentialStatus struct {
	Token               string `json:"token"`
	ExpirationTimestamp string `json:"expirationTimestamp,omitempty"`
}

// clustersCredentialsCmd is invoked by kubectl (not humans) as the exec
// credential plugin wired up by `dina clusters connect`. It returns the CLI's
// current Dina session token as an ExecCredential; Dina's cluster proxy
// authenticates with that same token and impersonates the caller.
var clustersCredentialsCmd = &cobra.Command{
	Use:    "credentials <cluster>",
	Short:  "Print a Kubernetes ExecCredential (used by kubectl)",
	Long:   "Prints the CLI's current Dina session token as a client.authentication.k8s.io ExecCredential. kubectl calls this automatically via the context created by `dina clusters connect`; you don't normally run it yourself.",
	Args:   cobra.ExactArgs(1),
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := api.NewClient()
		if err != nil {
			return err
		}
		// The cluster arg doesn't affect the token — the same session token
		// authorizes every cluster, and Dina scopes access server-side. It's
		// kept for CLI symmetry and future use.
		tok, err := client.AccessToken()
		if err != nil {
			return err
		}
		status := execCredentialStatus{Token: tok.Token}
		if !tok.Expiry.IsZero() {
			status.ExpirationTimestamp = tok.Expiry.Format(time.RFC3339)
		}
		cred := execCredential{
			APIVersion: "client.authentication.k8s.io/v1",
			Kind:       "ExecCredential",
			Status:     status,
		}
		enc := json.NewEncoder(os.Stdout)
		if err := enc.Encode(cred); err != nil {
			return fmt.Errorf("encoding ExecCredential: %w", err)
		}
		return nil
	},
}

func init() {
	clustersCmd.AddCommand(clustersCredentialsCmd)
}
