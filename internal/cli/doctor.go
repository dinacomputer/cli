package cli

import (
	"fmt"
	"os"

	"github.com/dinacomputer/cli/internal/doctor"
	"github.com/spf13/cobra"
)

var doctorFix bool

var doctorCmd = &cobra.Command{
	Use:           "doctor",
	Short:         "Run diagnostic checks",
	Long:          "Run diagnostic checks on the CLI's local state: authentication, installed skills, and CLI version. Pass --fix to auto-repair any fixable issues.",
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
		status := fmt.Sprintf("[%s]", r.Status)
		fmt.Fprintf(os.Stderr, "%s%-7s %-20s %s", indent, status, r.Name, r.Summary)
		if r.Status != doctor.StatusOK && r.FixHint != "" && !r.Fixable() {
			fmt.Fprintf(os.Stderr, "  →  run: %s", r.FixHint)
		}
		fmt.Fprintln(os.Stderr)
		for _, d := range r.Details {
			fmt.Fprintf(os.Stderr, "%s        - %s\n", indent, d)
		}
	}
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
