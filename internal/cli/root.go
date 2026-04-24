package cli

import (
	"fmt"
	"os"

	"github.com/dinacomputer/cli/internal/skills"
	"github.com/spf13/cobra"
)

var noInput bool

var rootCmd = &cobra.Command{
	Use:              "dina",
	Short:            "Dina CLI",
	Long:             "Dina CLI – deploy applications, manage apps and environment variables, view logs, configure custom hostnames, and install AI agent skills for the Dina platform.",
	PersistentPreRun: runSkillCheck,
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&noInput, "no-input", false, "Disable interactive prompts (fail if input is required)")
}

// runSkillCheck surfaces a warning on stderr when installed skills are
// outdated. Skipped for commands where the warning would be noise (install,
// version) and for the skill-update command itself once we add one.
func runSkillCheck(cmd *cobra.Command, _ []string) {
	switch cmd.Name() {
	case "install", "version", "help", "doctor":
		return
	}
	res := skills.CheckIfDue(Version)
	if res == nil || len(res.Outdated) == 0 {
		return
	}
	fmt.Fprintln(os.Stderr, "! Alert: installed skills are outdated. Run `dina doctor --fix` to update.")
	for _, s := range res.Outdated {
		fmt.Fprintf(os.Stderr, "!   - %s (%s): %s\n", s.Agent, s.Scope, s.Path)
	}
	fmt.Fprintln(os.Stderr)
}

// NoInput reports whether the --no-input flag was passed.
func NoInput() bool {
	return noInput
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		if msg := err.Error(); msg != "" {
			fmt.Fprintln(os.Stderr, msg)
		}
		os.Exit(1)
	}
}
