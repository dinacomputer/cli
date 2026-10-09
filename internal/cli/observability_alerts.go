package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/dinacomputer/cli/internal/api"
	"github.com/spf13/cobra"
)

var obsAlertsCmd = &cobra.Command{
	Use:     "alerts",
	Aliases: []string{"alert"},
	Short:   "List firing alerts, most severe first",
	Long: `List alerts that are firing or suppressed, most severe first. Explain one
with: dina obs alerts show <fingerprint>

Alert rules and notification channels are managed in the SigNoz UI
(dina obs status prints its address).`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, orgID, err := obsClient()
		if err != nil {
			return err
		}
		alerts, err := client.ListAlerts(orgID)
		if err != nil {
			return err
		}
		if JSONOutput() {
			return writeJSON(alerts)
		}
		if len(alerts) == 0 {
			fmt.Println("No alerts firing.")
			return nil
		}
		now := time.Now()
		fmt.Printf("%-16s  %-9s  %-10s  %-7s  %-40s  %s\n", "FINGERPRINT", "SEVERITY", "STATE", "FOR", "ALERT", "VALUE")
		for _, a := range alerts {
			fmt.Printf("%-16s  %-9s  %-10s  %-7s  %-40s  %s\n",
				a.Fingerprint, orDash(a.Severity), a.State, formatAgo(a.StartsAt, now), truncate(a.Name, 40), breach(a))
		}
		return nil
	},
}

var obsAlertShowCmd = &cobra.Command{
	Use:   "show <fingerprint>",
	Short: "Explain one alert: its rule, history, and related exceptions and error logs",
	Long: `Explain one alert: what fired and by how much, how often the rule fired in the
last day compared with the day before, and the exceptions and error logs of the
affected service since shortly before the alert started.

A unique prefix of the fingerprint is enough.`,
	Example: `  dina obs alerts show 3f9a`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, orgID, err := obsClient()
		if err != nil {
			return err
		}
		fingerprint, err := resolveFingerprint(client, orgID, args[0])
		if err != nil {
			return err
		}
		ac, err := client.GetAlert(orgID, fingerprint)
		if err != nil {
			return err
		}
		if JSONOutput() {
			return writeJSON(ac)
		}
		printAlertContext(ac, time.Now())
		return nil
	},
}

func init() {
	obsAlertsCmd.AddCommand(obsAlertShowCmd)
	observabilityCmd.AddCommand(obsAlertsCmd)
}

// fingerprintLen is the length of an Alertmanager fingerprint. Anything
// shorter is treated as a prefix and matched against the firing alerts.
const fingerprintLen = 16

func resolveFingerprint(client *api.Client, orgID, arg string) (string, error) {
	if len(arg) >= fingerprintLen {
		return arg, nil
	}
	alerts, err := client.ListAlerts(orgID)
	if err != nil {
		return "", err
	}
	return matchFingerprint(alerts, arg)
}

func matchFingerprint(alerts []api.Alert, prefix string) (string, error) {
	var matches []string
	for _, a := range alerts {
		if strings.HasPrefix(a.Fingerprint, prefix) {
			matches = append(matches, a.Fingerprint)
		}
	}
	switch len(matches) {
	case 0:
		return "", fmt.Errorf("no firing alert matches %q — it may have resolved; list alerts with: dina obs alerts", prefix)
	case 1:
		return matches[0], nil
	default:
		return "", fmt.Errorf("%q matches %d alerts (%s) — use more characters", prefix, len(matches), strings.Join(matches, ", "))
	}
}

// breach renders "12 above 5" from an alert's value, comparison and
// threshold, or just the value when the rule did not say more.
func breach(a api.Alert) string {
	switch {
	case a.Value == "":
		return "-"
	case a.Threshold == "":
		return a.Value
	default:
		op := a.CompareOp
		if op == "" {
			op = "vs"
		}
		return fmt.Sprintf("%s %s %s", a.Value, op, a.Threshold)
	}
}

func orDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

func printAlertContext(ac *api.AlertContext, now time.Time) {
	a := ac.Alert
	fmt.Printf("%s  [%s, %s]\n", a.Name, orDash(a.Severity), a.State)
	fmt.Printf("Firing for %s, since %s\n", formatAgo(a.StartsAt, now), formatTime(a.StartsAt))
	if b := breach(a); b != "-" {
		fmt.Printf("Value:   %s\n", b)
	}
	if a.Summary != "" {
		fmt.Printf("Summary: %s\n", oneLine(a.Summary))
	}
	if a.Description != "" {
		fmt.Printf("\n%s\n", strings.TrimSpace(a.Description))
	}

	if ac.Rule != nil {
		fmt.Printf("\nRule:    %s (%s, %s)\n", ac.Rule.Name, ac.Rule.Type, ac.Rule.State)
	}
	if ac.History != nil {
		fmt.Printf("History: fired %d times in the last 24h, %d in the 24h before\n", ac.History.Triggers, ac.History.PreviousTriggers)
	}
	if len(a.Labels) > 0 {
		var labels []string
		for _, k := range sortedKeys(a.Labels) {
			if internalAlertLabels[k] {
				continue
			}
			labels = append(labels, k+"="+a.Labels[k])
		}
		if len(labels) > 0 {
			fmt.Printf("Labels:  %s\n", strings.Join(labels, " "))
		}
	}

	if ac.Service != "" {
		fmt.Printf("\nExceptions in %s since shortly before it started:\n", ac.Service)
		if len(ac.Exceptions) == 0 {
			fmt.Println("  none")
		}
		for _, e := range ac.Exceptions {
			fmt.Printf("  %5dx  %s  (dina obs exceptions show %s --since 24h)\n", e.Count, truncate(oneLine(e.Type+": "+e.Message), 80), e.GroupID)
		}

		fmt.Printf("\nRecent error logs from %s:\n", ac.Service)
		if len(ac.ErrorLogs) == 0 {
			fmt.Println("  none")
		}
		for _, l := range ac.ErrorLogs {
			fmt.Printf("  %s  %s\n", formatTime(l.Timestamp), truncate(oneLine(l.Body), 120))
		}
	} else {
		fmt.Println("\nThe alert's labels do not name a service, so no related exceptions or logs were fetched.")
	}

	if a.UIURL != "" || a.RelatedLogsURL != "" || a.RelatedTracesURL != "" {
		fmt.Println()
	}
	if a.UIURL != "" {
		fmt.Printf("Rule in SigNoz:  %s\n", a.UIURL)
	}
	if a.RelatedLogsURL != "" {
		fmt.Printf("Related logs:    %s\n", a.RelatedLogsURL)
	}
	if a.RelatedTracesURL != "" {
		fmt.Printf("Related traces:  %s\n", a.RelatedTracesURL)
	}
	for _, w := range ac.Warnings {
		Infof("Could not fetch %s\n", w)
	}
}

// internalAlertLabels are set by SigNoz for its own bookkeeping and say
// nothing a reader needs; the rest are the rule's own labels and group-by
// values.
var internalAlertLabels = map[string]bool{
	"alertname":      true,
	"ruleId":         true,
	"ruleSource":     true,
	"severity":       true,
	"threshold.name": true,
}
