package cli

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/dinacomputer/cli/internal/api"
	"github.com/spf13/cobra"
)

// obsOrg, obsSince and obsUntil are shared by every observability
// subcommand. --since takes a duration (30m, 24h, 7d) or an RFC 3339 time.
var (
	obsOrg   string
	obsSince string
	obsUntil string
)

var observabilityCmd = &cobra.Command{
	Use:     "observability",
	Aliases: []string{"obs"},
	Short:   "Inspect logs, traces, exceptions, metrics and alerts",
	Long: `Read the telemetry your organization sends to its SigNoz instance: service
health, logs, traces, exceptions, metrics, infrastructure, and firing alerts.

Observability is enabled per organization by a platform operator. Run
` + "`dina obs status`" + ` to see whether it is on and where the full UI lives. Alert
rules, channels and dashboards are managed in that UI; the CLI only reads.

Time ranges default to the last hour. --since takes a duration (30m, 24h, 7d)
or an RFC 3339 time; --until takes an RFC 3339 time.

Filters (--filter) are written like a SQL WHERE clause without the WHERE:
  service.name = 'api' AND http.response.status_code >= 500
  body CONTAINS 'timeout' OR severity_text IN ('ERROR', 'FATAL')
Supported: = != < <= > >=, AND, OR, NOT, parentheses, IN (...), LIKE, ILIKE,
CONTAINS, EXISTS, IS [NOT] NULL. List field names with ` + "`dina obs fields`" + `.`,
	GroupID: groupDeploy,
}

var obsStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show whether observability is enabled and where its UI is",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, orgID, err := obsClient()
		if err != nil {
			return err
		}
		st, err := client.ObservabilityStatus(orgID)
		if err != nil {
			return err
		}
		if JSONOutput() {
			return writeJSON(st)
		}
		if !st.Enabled {
			fmt.Println("Observability is not enabled for this organization. Ask a platform operator to enable it.")
			return nil
		}
		fmt.Printf("Enabled (%s)\nUI: %s\n", st.Kind, st.UIURL)
		return nil
	},
}

func init() {
	observabilityCmd.PersistentFlags().StringVar(&obsOrg, "org", "", "Organization id or name (defaults to your only organization)")
	observabilityCmd.AddCommand(obsStatusCmd)
	rootCmd.AddCommand(observabilityCmd)
}

// addWindowFlags gives a command --since and --until.
func addWindowFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&obsSince, "since", "", "Start: duration before --until (30m, 24h, 7d) or RFC 3339 time (default 1h)")
	cmd.Flags().StringVar(&obsUntil, "until", "", "End: RFC 3339 time (default now)")
}

func obsWindow() api.Window {
	return api.Window{Since: obsSince, Until: obsUntil}
}

// obsClient validates --output and resolves the organization, which every
// observability command needs before anything else.
func obsClient() (*api.Client, string, error) {
	if err := ValidateOutput(); err != nil {
		return nil, "", err
	}
	client, err := api.NewClient()
	if err != nil {
		return nil, "", err
	}
	orgID, err := resolveOrgID(client, obsOrg)
	if err != nil {
		return nil, "", err
	}
	return client, orgID, nil
}

// --- formatting ---

const timeLayout = "2006-01-02 15:04:05"

func formatTime(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	return t.Local().Format(timeLayout)
}

// formatAgo renders how long ago t was, coarsely: 45s, 12m, 3h, 2d.
func formatAgo(t time.Time, now time.Time) string {
	d := now.Sub(t)
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 48*time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	}
}

// formatMs renders a duration in milliseconds with a unit that keeps it short.
func formatMs(ms float64) string {
	switch {
	case ms >= 60_000:
		return strconv.FormatFloat(ms/60_000, 'f', 1, 64) + "m"
	case ms >= 1000:
		return strconv.FormatFloat(ms/1000, 'f', 2, 64) + "s"
	case ms >= 10:
		return strconv.FormatFloat(ms, 'f', 0, 64) + "ms"
	default:
		return strconv.FormatFloat(ms, 'f', 1, 64) + "ms"
	}
}

func formatPercent(fraction float64) string {
	return strconv.FormatFloat(fraction*100, 'f', 1, 64) + "%"
}

// oneLine collapses a multi-line value so a list stays one row per entry.
func oneLine(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n-1]) + "…"
}

// sortedKeys returns m's keys in order, for stable label output.
func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// moreHint tells the reader how to fetch the next page, on stderr so piping
// stdout stays clean.
func moreHint(cursor string) {
	if cursor != "" {
		Infof("More results: rerun with --cursor %s\n", cursor)
	}
}
