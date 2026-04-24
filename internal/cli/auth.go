package cli

import (
	"fmt"
	"time"

	"github.com/dinacomputer/cli/internal/auth"
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication",
	Long: `Log in, log out, and inspect the authentication state.

The CLI stores OAuth credentials at ~/.config/dina/auth.json with 0600 permissions.
Most commands require you to be authenticated — run ` + "`dina auth login`" + ` first.`,
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with the Dina platform",
	Long: `Authenticate with the Dina platform.

The CLI discovers the auth server via RFC 9728, then uses the OAuth device
code flow when available (works over SSH / headless) and falls back to PKCE
authorization code flow otherwise. Tokens are stored at ~/.config/dina/auth.json.`,
	Example: `  # interactive login (opens a browser or prompts for a device code)
  dina auth login`,
	RunE: func(cmd *cobra.Command, args []string) error {
		creds, err := auth.Login()
		if err != nil {
			return err
		}
		Infof("Authenticated! Token expires at %s\n", creds.ExpiresAt.Format("2006-01-02 15:04:05"))
		return nil
	},
}

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Log out and clear stored credentials",
	Long:  "Remove the stored OAuth credentials. Re-run `dina auth login` to authenticate again.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := auth.ClearCredentials(); err != nil {
			return err
		}
		Infoln("Logged out.")
		return nil
	},
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current authentication status",
	Long:  "Print whether the CLI is authenticated and when the access token expires.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := ValidateOutput(); err != nil {
			return err
		}
		creds, err := auth.LoadCredentials()
		if err != nil {
			return err
		}

		type statusPayload struct {
			Authenticated bool      `json:"authenticated"`
			Expired       bool      `json:"expired,omitempty"`
			ExpiresAt     time.Time `json:"expires_at,omitempty"`
			APIBaseURL    string    `json:"api_base_url,omitempty"`
			Issuer        string    `json:"issuer,omitempty"`
		}

		payload := statusPayload{}
		if creds != nil && creds.AccessToken != "" {
			payload.Authenticated = true
			payload.Expired = creds.Expired()
			payload.ExpiresAt = creds.ExpiresAt
			payload.APIBaseURL = creds.APIBaseURL
			payload.Issuer = creds.Issuer
		}

		if JSONOutput() {
			return writeJSON(payload)
		}

		if !payload.Authenticated {
			fmt.Println("Not authenticated. Run: dina auth login")
			return nil
		}
		if payload.Expired {
			fmt.Println("Authenticated (token expired — will refresh on next API call)")
		} else {
			fmt.Printf("Authenticated. Token expires at %s\n", payload.ExpiresAt.Format("2006-01-02 15:04:05"))
		}
		return nil
	},
}

func init() {
	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authLogoutCmd)
	authCmd.AddCommand(authStatusCmd)
	rootCmd.AddCommand(authCmd)
}
