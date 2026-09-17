package cli

import (
	"fmt"

	"github.com/dinacomputer/cli/internal/api"
	"github.com/spf13/cobra"
)

var storageBucketsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List buckets",
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
		Infoln("Fetching buckets...")
		buckets, err := client.ListBuckets(orgID)
		if err != nil {
			return err
		}
		if JSONOutput() {
			return writeJSON(buckets)
		}
		if len(buckets) == 0 {
			fmt.Println("No buckets found.")
			return nil
		}
		for _, b := range buckets {
			used := "-"
			if b.Usage != nil {
				used = formatBytes(b.Usage.Bytes)
			}
			fmt.Printf("%-28s  %-10s  %-9s  %10s / %-10s\n", b.Name, b.Status, visibility(b.Public), used, formatBytes(b.QuotaBytes))
		}
		return nil
	},
}

var (
	bucketCreateBackend      string
	bucketCreatePublic       bool
	bucketCreateQuota        string
	bucketCreateQuotaObjects int64
)

var storageBucketsCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a bucket",
	Long: `Create a bucket in the organization.

A new bucket is private and reachable only by keys you grant access to. Pass
--public to serve every object in it anonymously over HTTP — that applies to the
whole bucket, not to individual objects, so don't mix private data into it.

Quotas default to the organization's unallocated remainder.`,
	Example: `  dina storage buckets create uploads
  dina storage buckets create assets --public --quota 10GB
  dina storage buckets create archive --quota 500GB --quota-objects 1000000`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := ValidateOutput(); err != nil {
			return err
		}
		input := api.CreateBucketInput{
			Name:         args[0],
			Backend:      bucketCreateBackend,
			Public:       bucketCreatePublic,
			QuotaObjects: bucketCreateQuotaObjects,
		}
		if bucketCreateQuota != "" {
			n, err := parseSize(bucketCreateQuota)
			if err != nil {
				return err
			}
			input.QuotaBytes = n
		}

		client, err := api.NewClient()
		if err != nil {
			return err
		}
		orgID, err := resolveOrgID(client, storageOrg)
		if err != nil {
			return err
		}
		b, err := client.CreateBucket(orgID, input)
		if err != nil {
			return err
		}
		if JSONOutput() {
			return writeJSON(b)
		}
		fmt.Printf("Bucket created: %s\n", b.Name)
		printBucket(b)
		return nil
	},
}

var storageBucketsGetCmd = &cobra.Command{
	Use:     "get <name>",
	Aliases: []string{"info"},
	Short:   "Show a bucket's details, endpoint, and usage",
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
		b, err := client.GetBucket(orgID, args[0])
		if err != nil {
			return err
		}
		if JSONOutput() {
			return writeJSON(b)
		}
		printBucket(b)
		return nil
	},
}

var (
	bucketUpdatePublic       bool
	bucketUpdateQuota        string
	bucketUpdateQuotaObjects int64
)

var storageBucketsUpdateCmd = &cobra.Command{
	Use:   "update <name>",
	Short: "Update a bucket's visibility or quotas",
	Long: `Change a bucket's public visibility or its quotas. Only the flags you pass are
sent; everything else is left as it is.`,
	Example: `  # make a bucket private again
  dina storage buckets update assets --public=false

  # raise the size ceiling
  dina storage buckets update uploads --quota 50GB`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := ValidateOutput(); err != nil {
			return err
		}
		input := api.UpdateBucketInput{}
		if cmd.Flags().Changed("public") {
			input.Public = &bucketUpdatePublic
		}
		if cmd.Flags().Changed("quota") {
			n, err := parseSize(bucketUpdateQuota)
			if err != nil {
				return err
			}
			input.QuotaBytes = &n
		}
		if cmd.Flags().Changed("quota-objects") {
			input.QuotaObjects = &bucketUpdateQuotaObjects
		}
		if input.Public == nil && input.QuotaBytes == nil && input.QuotaObjects == nil {
			return fmt.Errorf("nothing to update — pass --public, --quota, or --quota-objects")
		}

		client, err := api.NewClient()
		if err != nil {
			return err
		}
		orgID, err := resolveOrgID(client, storageOrg)
		if err != nil {
			return err
		}
		b, err := client.UpdateBucket(orgID, args[0], input)
		if err != nil {
			return err
		}
		if JSONOutput() {
			return writeJSON(b)
		}
		fmt.Printf("Bucket %q updated.\n", b.Name)
		printBucket(b)
		return nil
	},
}

var bucketDeleteForce bool

var storageBucketsDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a bucket",
	Long: `Permanently delete a bucket and everything stored in it.

By default you will be prompted to type the bucket's name to confirm. Pass
--force to skip the prompt in scripts.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := confirmByName("bucket", args[0], bucketDeleteForce); err != nil {
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
		if err := client.DeleteBucket(orgID, args[0]); err != nil {
			return err
		}
		Infof("Bucket %q deleted.\n", args[0])
		return nil
	},
}

func visibility(public bool) string {
	if public {
		return "public"
	}
	return "private"
}

func printBucket(b *api.Bucket) {
	fmt.Printf("Name:          %s\n", b.Name)
	fmt.Printf("Status:        %s\n", b.Status)
	fmt.Printf("Visibility:    %s\n", visibility(b.Public))
	if b.PublicURL != "" {
		fmt.Printf("Public URL:    %s\n", b.PublicURL)
	}
	fmt.Printf("Quota:         %s, %s objects\n", formatBytes(b.QuotaBytes), formatCount(b.QuotaObjects))
	if b.Usage != nil {
		fmt.Printf("Usage:         %s, %d objects (as of %s)\n", formatBytes(b.Usage.Bytes), b.Usage.Objects, b.Usage.ObservedAt.Format("2006-01-02 15:04"))
	}
	fmt.Println()
	// S3 clients address the physical name, not the display name, so the two
	// are printed together to avoid a confusing NoSuchBucket.
	fmt.Println("S3 endpoint:")
	fmt.Printf("  Endpoint:    %s\n", b.Endpoint.URL)
	fmt.Printf("  Region:      %s\n", b.Endpoint.Region)
	fmt.Printf("  Bucket:      %s\n", b.PhysicalName)
	if b.Endpoint.PathStyle {
		fmt.Printf("  Addressing:  path-style (required)\n")
	}
	fmt.Println()
	fmt.Printf("  aws --endpoint-url %s s3 ls s3://%s\n", b.Endpoint.URL, b.PhysicalName)
}

func init() {
	storageBucketsCreateCmd.Flags().StringVar(&bucketCreateBackend, "backend", "", "Backend to place the bucket on (omit while only one exists)")
	storageBucketsCreateCmd.Flags().BoolVar(&bucketCreatePublic, "public", false, "Serve every object in the bucket anonymously")
	storageBucketsCreateCmd.Flags().StringVar(&bucketCreateQuota, "quota", "", "Size ceiling, e.g. 500MB or 10GB (default: the org's unallocated remainder)")
	storageBucketsCreateCmd.Flags().Int64Var(&bucketCreateQuotaObjects, "quota-objects", 0, "Object-count ceiling (0 for unlimited)")

	storageBucketsUpdateCmd.Flags().BoolVar(&bucketUpdatePublic, "public", false, "Serve every object anonymously (--public=false to make private)")
	storageBucketsUpdateCmd.Flags().StringVar(&bucketUpdateQuota, "quota", "", "Size ceiling, e.g. 500MB or 10GB")
	storageBucketsUpdateCmd.Flags().Int64Var(&bucketUpdateQuotaObjects, "quota-objects", 0, "Object-count ceiling (0 for unlimited)")

	storageBucketsDeleteCmd.Flags().BoolVarP(&bucketDeleteForce, "force", "f", false, "Skip confirmation prompt")

	storageBucketsCmd.AddCommand(storageBucketsListCmd)
	storageBucketsCmd.AddCommand(storageBucketsCreateCmd)
	storageBucketsCmd.AddCommand(storageBucketsGetCmd)
	storageBucketsCmd.AddCommand(storageBucketsUpdateCmd)
	storageBucketsCmd.AddCommand(storageBucketsDeleteCmd)
}
