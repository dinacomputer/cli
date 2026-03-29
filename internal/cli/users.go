package cli

import (
	"fmt"

	"github.com/dinacomputer/cli/internal/api"
	"github.com/spf13/cobra"
)

var usersCmd = &cobra.Command{
	Use:   "users",
	Short: "Manage users (admin only)",
	Long:  "List and manage platform users. These commands require admin privileges.",
}

var usersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all users",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := api.NewClient()
		if err != nil {
			return err
		}
		users, err := client.ListUsers()
		if err != nil {
			return err
		}
		if len(users) == 0 {
			fmt.Println("No users found.")
			return nil
		}
		for _, u := range users {
			admin := ""
			if u.IsAdmin {
				admin = " (admin)"
			}
			fmt.Printf("%-36s  %-30s  %s%s\n", u.ID, u.Email, u.Status, admin)
		}
		return nil
	},
}

var usersActivateCmd = &cobra.Command{
	Use:   "activate <user-id>",
	Short: "Activate a user",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := api.NewClient()
		if err != nil {
			return err
		}
		user, err := client.ActivateUser(args[0])
		if err != nil {
			return err
		}
		fmt.Printf("User %s (%s) activated.\n", user.Email, user.ID)
		return nil
	},
}

func init() {
	usersCmd.AddCommand(usersListCmd)
	usersCmd.AddCommand(usersActivateCmd)
	rootCmd.AddCommand(usersCmd)
}
