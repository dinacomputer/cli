package cli

import (
	"fmt"
	"os"

	"github.com/dinacomputer/cli/internal/color"
	"github.com/dinacomputer/cli/internal/doctor"
	"github.com/spf13/cobra"
)

var doctorFix bool

var doctorCmd = &cobra.Command{
	Use:           "doctor",
	Short:         "Run diagnostic checks",
	Long:          "Run diagnostic checks on the CLI's local state: authentication, installed skills, and CLI version. Pass --fix to auto-repair any fixable issues.",
	GroupID:       groupCLI,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		results := doctor.RunAll(doctor.DefaultChecks(Version))

		printResults(results, "")

		issues := countIssues(results)
		if issues == 0 {
			fmt.Fprintln(os.Stderr, "\nAll checks passed.")
			return nil
		}

		if !doctorFix {
			fmt.Fprintf(os.Stderr, "\n%d issue(s) found. Run `dina doctor --fix` to repair where possible.\n", issues)
			return silentExitError{}
		}

		fmt.Fprintln(os.Stderr, "\nFixing...")
		for i := range results {
			r := &results[i]
			if r.Status == doctor.StatusOK || !r.Fixable() {
				continue
			}
			fmt.Fprintf(os.Stderr, "  %-20s ", r.Name)
			if err := r.Fix(); err != nil {
				fmt.Fprintf(os.Stderr, "fix failed: %s\n", err)
				continue
			}
			fmt.Fprintln(os.Stderr, "fixed")
		}

		fmt.Fprintln(os.Stderr, "\nRe-checking...")
		final := doctor.RunAll(doctor.DefaultChecks(Version))
		printResults(final, "  ")
		remaining := countIssues(final)
		if remaining == 0 {
			fmt.Fprintln(os.Stderr, "\nAll checks passed.")
			return nil
		}
		fmt.Fprintf(os.Stderr, "\n%d issue(s) remain. Follow the hints above to resolve.\n", remaining)
		return silentExitError{}
	},
}

// silentExitError conveys a non-zero exit without triggering cobra/root's
// default error print, since doctor has already rendered the user-facing
// output itself.
type silentExitError struct{}

func (silentExitError) Error() string { return "" }

func printResults(results []doctor.Result, indent string) {
	for _, r := range results {
		tag := fmt.Sprintf("[%s]", r.Status)
		switch r.Status {
		case doctor.StatusOK:
			tag = color.Green(tag)
		case doctor.StatusWarn:
			tag = color.Yellow(tag)
		case doctor.StatusFail:
			tag = color.Red(tag)
		}
		// %-7s on a colored string counts ANSI bytes; render the raw tag
		// into a width-aware slot manually.
		rawTag := fmt.Sprintf("[%s]", r.Status)
		pad := 7 - len(rawTag)
		if pad < 1 {
			pad = 1
		}
		fmt.Fprintf(os.Stderr, "%s%s%s%-20s %s", indent, tag, spaces(pad), r.Name, r.Summary)
		if r.Status != doctor.StatusOK && r.FixHint != "" && !r.Fixable() {
			fmt.Fprintf(os.Stderr, "  →  run: %s", color.Bold(r.FixHint))
		}
		fmt.Fprintln(os.Stderr)
		for _, d := range r.Details {
			fmt.Fprintf(os.Stderr, "%s        - %s\n", indent, d)
		}
	}
}

func spaces(n int) string {
	if n <= 0 {
		return ""
	}
	s := make([]byte, n)
	for i := range s {
		s[i] = ' '
	}
	return string(s)
}

func countIssues(results []doctor.Result) int {
	n := 0
	for _, r := range results {
		if r.Status != doctor.StatusOK {
			n++
		}
	}
	return n
}

func init() {
	doctorCmd.Flags().BoolVar(&doctorFix, "fix", false, "Attempt to auto-repair any fixable issues")
	rootCmd.AddCommand(doctorCmd)
}
