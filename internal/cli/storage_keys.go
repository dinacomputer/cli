package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/dinacomputer/cli/internal/api"
	"github.com/dinacomputer/cli/internal/color"
	"github.com/spf13/cobra"
)

var storageKeysListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List storage access keys",
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
		Infoln("Fetching storage keys...")
		keys, err := client.ListStorageKeys(orgID)
		if err != nil {
			return err
		}
		if JSONOutput() {
			return writeJSON(keys)
		}
		if len(keys) == 0 {
			fmt.Println("No storage keys found.")
			return nil
		}
		for _, k := range keys {
			fmt.Printf("%-24s  %-24s  %s\n", k.Name, k.AccessKeyID, summarizeGrants(k.Grants))
		}
		return nil
	},
}

var keyCreateBackend string

var storageKeysCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a storage access key",
	Long: `Mint an S3 access key pair for the organization.

The secret is shown once and cannot be retrieved afterwards — store it before
moving on. A new key has no access to any bucket; grant it with
dina storage keys grant.`,
	Example: `  dina storage keys create my-app
  dina storage keys grant my-app uploads --read --write`,
	Args: cobra.ExactArgs(1),
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
		issued, err := client.CreateStorageKey(orgID, api.CreateStorageKeyInput{
			Name:    args[0],
			Backend: keyCreateBackend,
		})
		if err != nil {
			return err
		}
		if JSONOutput() {
			return writeJSON(issued)
		}
		fmt.Printf("Storage key created: %s\n\n", issued.Key.Name)
		printIssuedKey(issued)
		return nil
	},
}

var storageKeysGetCmd = &cobra.Command{
	Use:     "get <name>",
	Aliases: []string{"info"},
	Short:   "Show a key's details and bucket grants",
	Args:    cobra.ExactArgs(1),
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
		k, err := client.GetStorageKey(orgID, args[0])
		if err != nil {
			return err
		}
		if JSONOutput() {
			return writeJSON(k)
		}
		fmt.Printf("Name:           %s\n", k.Name)
		fmt.Printf("Access key ID:  %s\n", k.AccessKeyID)
		fmt.Printf("Created:        %s\n", k.CreatedAt.Format("2006-01-02 15:04"))
		if k.ExpiresAt != nil {
			fmt.Printf("Expires:        %s\n", k.ExpiresAt.Format("2006-01-02 15:04"))
		}
		fmt.Println("\nGrants:")
		if len(k.Grants) == 0 {
			fmt.Println("  (none — this key cannot reach any bucket)")
			return nil
		}
		for _, g := range k.Grants {
			fmt.Printf("  %-28s  %s\n", g.BucketName, permissions(g))
		}
		return nil
	},
}

var keyRotateForce bool

var storageKeysRotateCmd = &cobra.Command{
	Use:   "rotate <name>",
	Short: "Rotate a storage access key",
	Long: `Issue a fresh secret for a key, invalidating the previous one.

Anything still using the old credentials stops working the moment this
completes, so roll the new secret out first. Grants are preserved.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := ValidateOutput(); err != nil {
			return err
		}
		prompt := fmt.Sprintf("Rotating key %q invalidates its current secret immediately. Continue?", args[0])
		if err := confirmYesNo(prompt, keyRotateForce); err != nil {
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
		issued, err := client.RotateStorageKey(orgID, args[0])
		if err != nil {
			return err
		}
		if JSONOutput() {
			return writeJSON(issued)
		}
		fmt.Printf("Storage key rotated: %s\n\n", issued.Key.Name)
		printIssuedKey(issued)
		return nil
	},
}

var keyDeleteForce bool

var storageKeysDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a storage access key",
	Long: `Permanently delete a key and every grant it holds. Anything authenticating
with it stops working immediately.

By default you will be prompted to type the key's name to confirm. Pass --force
to skip the prompt in scripts.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := confirmByName("storage key", args[0], keyDeleteForce); err != nil {
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
		if err := client.DeleteStorageKey(orgID, args[0]); err != nil {
			return err
		}
		Infof("Storage key %q deleted.\n", args[0])
		return nil
	},
}

var (
	grantRead  bool
	grantWrite bool
	grantOwner bool
)

var storageKeysGrantCmd = &cobra.Command{
	Use:   "grant <key> <bucket>",
	Short: "Grant a key access to a bucket",
	Long: `Set what a key may do on one bucket. The permissions you pass replace the
key's existing permissions on that bucket rather than adding to them.

--owner implies full control, including bucket configuration.`,
	Example: `  # read-only
  dina storage keys grant my-app assets --read

  # read and write
  dina storage keys grant my-app uploads --read --write`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := ValidateOutput(); err != nil {
			return err
		}
		if !grantRead && !grantWrite && !grantOwner {
			return fmt.Errorf("no permissions given — pass at least one of --read, --write, or --owner")
		}
		client, err := api.NewClient()
		if err != nil {
			return err
		}
		orgID, err := resolveOrgID(client, storageOrg)
		if err != nil {
			return err
		}
		g, err := client.SetStorageKeyGrant(orgID, args[0], args[1], api.SetGrantInput{
			Read:  grantRead,
			Write: grantWrite,
			Owner: grantOwner,
		})
		if err != nil {
			return err
		}
		if JSONOutput() {
			return writeJSON(g)
		}
		fmt.Printf("Key %q granted %s on bucket %q.\n", args[0], permissions(*g), args[1])
		return nil
	},
}

var revokeForce bool

var storageKeysRevokeCmd = &cobra.Command{
	Use:   "revoke <key> <bucket>",
	Short: "Revoke a key's access to a bucket",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		prompt := fmt.Sprintf("Revoke key %q's access to bucket %q?", args[0], args[1])
		if err := confirmYesNo(prompt, revokeForce); err != nil {
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
		if err := client.DeleteStorageKeyGrant(orgID, args[0], args[1]); err != nil {
			return err
		}
		Infof("Key %q no longer has access to bucket %q.\n", args[0], args[1])
		return nil
	},
}

// permissions renders a grant's flags as a compact "read,write" string.
func permissions(g api.StorageGrant) string {
	var perms []string
	if g.Read {
		perms = append(perms, "read")
	}
	if g.Write {
		perms = append(perms, "write")
	}
	if g.Owner {
		perms = append(perms, "owner")
	}
	if len(perms) == 0 {
		return "no access"
	}
	return strings.Join(perms, ",")
}

func summarizeGrants(grants []api.StorageGrant) string {
	if len(grants) == 0 {
		return "(no grants)"
	}
	parts := make([]string, len(grants))
	for i, g := range grants {
		parts[i] = fmt.Sprintf("%s:%s", g.BucketName, permissions(g))
	}
	return strings.Join(parts, " ")
}

// printIssuedKey writes the one-time secret. The credential goes to stdout so
// redirecting the command captures it; the warning goes to stderr, and unlike
// Infof it is not silenced by --quiet.
func printIssuedKey(issued *api.IssuedStorageKey) {
	fmt.Printf("  AWS_ACCESS_KEY_ID=%s\n", issued.Key.AccessKeyID)
	fmt.Printf("  AWS_SECRET_ACCESS_KEY=%s\n", issued.SecretAccessKey)
	fmt.Fprintln(os.Stderr, color.Yellow("\nThe secret is shown once and cannot be retrieved later — store it now."))
	fmt.Fprintf(os.Stderr, "Grant it a bucket with: dina storage keys grant %s <bucket> --read --write\n", issued.Key.Name)
}

func init() {
	storageKeysCreateCmd.Flags().StringVar(&keyCreateBackend, "backend", "", "Backend to mint the credential on (omit while only one exists)")

	storageKeysRotateCmd.Flags().BoolVarP(&keyRotateForce, "force", "f", false, "Skip confirmation prompt")
	storageKeysDeleteCmd.Flags().BoolVarP(&keyDeleteForce, "force", "f", false, "Skip confirmation prompt")

	storageKeysGrantCmd.Flags().BoolVar(&grantRead, "read", false, "Allow reading objects")
	storageKeysGrantCmd.Flags().BoolVar(&grantWrite, "write", false, "Allow writing and deleting objects")
	storageKeysGrantCmd.Flags().BoolVar(&grantOwner, "owner", false, "Full control, including bucket configuration")

	storageKeysRevokeCmd.Flags().BoolVarP(&revokeForce, "force", "f", false, "Skip confirmation prompt")

	storageKeysCmd.AddCommand(storageKeysListCmd)
	storageKeysCmd.AddCommand(storageKeysCreateCmd)
	storageKeysCmd.AddCommand(storageKeysGetCmd)
	storageKeysCmd.AddCommand(storageKeysRotateCmd)
	storageKeysCmd.AddCommand(storageKeysDeleteCmd)
	storageKeysCmd.AddCommand(storageKeysGrantCmd)
	storageKeysCmd.AddCommand(storageKeysRevokeCmd)
}
