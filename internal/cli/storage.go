package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// storageOrg is the --org flag shared across storage subcommands. It accepts an
// organization id or name; when empty the caller's sole organization is used.
var storageOrg string

var storageCmd = &cobra.Command{
	Use:   "storage",
	Short: "Manage object storage buckets and access keys",
	Long: `Manage S3-compatible object storage: buckets that hold objects, and access
keys your applications authenticate with.

Buckets and keys are both scoped to an organization. A key can reach no bucket
until you grant it: dina storage keys grant <key> <bucket> --read --write.

The CLI manages buckets and credentials, not their contents — point any S3
client (aws, rclone, an SDK) at the endpoint from ` + "`dina storage buckets get`" + `
to read and write objects.`,
	GroupID: groupDeploy,
}

var storageBucketsCmd = &cobra.Command{
	Use:     "buckets",
	Aliases: []string{"bucket"},
	Short:   "Manage storage buckets",
}

var storageKeysCmd = &cobra.Command{
	Use:     "keys",
	Aliases: []string{"key"},
	Short:   "Manage storage access keys and their bucket grants",
}

func init() {
	storageCmd.PersistentFlags().StringVar(&storageOrg, "org", "", "Organization id or name (defaults to your only organization)")
	storageCmd.AddCommand(storageBucketsCmd)
	storageCmd.AddCommand(storageKeysCmd)
	rootCmd.AddCommand(storageCmd)
}

// sizeUnits are binary multiples, largest first so formatBytes picks the
// biggest unit that leaves a value >= 1.
var sizeUnits = []struct {
	suffix string
	factor int64
}{
	{"TB", 1 << 40},
	{"GB", 1 << 30},
	{"MB", 1 << 20},
	{"KB", 1 << 10},
}

// parseSize reads a byte quantity written either as a plain number or with a
// binary unit suffix: 500MB, 10GB, 2TB. Suffixes are case-insensitive, the "B"
// is optional and an "i" is tolerated, so 10g, 10GB and 10GiB all agree.
func parseSize(s string) (int64, error) {
	t := strings.ToUpper(strings.TrimSpace(s))
	if t == "" {
		return 0, fmt.Errorf("empty size")
	}
	t = strings.TrimSuffix(t, "B")
	t = strings.TrimSuffix(t, "I")

	factor := int64(1)
	for _, u := range sizeUnits {
		letter := u.suffix[:1]
		if strings.HasSuffix(t, letter) {
			factor = u.factor
			t = strings.TrimSpace(strings.TrimSuffix(t, letter))
			break
		}
	}

	n, err := strconv.ParseFloat(t, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid size %q — use a number optionally suffixed with KB, MB, GB, or TB", s)
	}
	if n < 0 {
		return 0, fmt.Errorf("invalid size %q — must not be negative", s)
	}
	return int64(n * float64(factor)), nil
}

// formatBytes renders a byte count for display. Zero reads as "unlimited",
// which is what a zero quota means on the API.
func formatBytes(n int64) string {
	if n == 0 {
		return "unlimited"
	}
	for _, u := range sizeUnits {
		if n >= u.factor {
			v := strconv.FormatFloat(float64(n)/float64(u.factor), 'f', 1, 64)
			return strings.TrimSuffix(v, ".0") + u.suffix
		}
	}
	return fmt.Sprintf("%dB", n)
}

// formatCount renders an object-count quota, where zero also means unlimited.
func formatCount(n int64) string {
	if n == 0 {
		return "unlimited"
	}
	return strconv.FormatInt(n, 10)
}
