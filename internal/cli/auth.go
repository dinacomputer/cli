package cli

import (
	"fmt"

	"github.com/dinacomputer/cli/internal/auth"
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication",
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with the Dina platform",
	RunE: func(cmd *cobra.Command, args []string) error {
		creds, err := auth.Login()
		if err != nil {
			return err
		}
		fmt.Printf("Authenticated! Token expires at %s\n", creds.ExpiresAt.Format("2006-01-02 15:04:05"))
		return nil
	},
}

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Log out and clear stored credentials",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := auth.ClearCredentials(); err != nil {
			return err
		}
		fmt.Println("Logged out.")
		return nil
	},
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current authentication status",
	RunE: func(cmd *cobra.Command, args []string) error {
		creds, err := auth.LoadCredentials()
		if err != nil {
			return err
		}
		if creds == nil || creds.AccessToken == "" {
			fmt.Println("Not authenticated. Run: dina auth login")
			return nil
		}
		if creds.Expired() {
			fmt.Println("Authenticated (token expired — will refresh on next API call)")
		} else {
			fmt.Printf("Authenticated. Token expires at %s\n", creds.ExpiresAt.Format("2006-01-02 15:04:05"))
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
