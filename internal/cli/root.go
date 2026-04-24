package cli

import (
	"fmt"
	"os"

	"github.com/dinacomputer/cli/internal/skills"
	"github.com/spf13/cobra"
)

var (
	noInput bool
	quiet   bool
	debug   bool
	noColor bool
)

var rootCmd = &cobra.Command{
	Use:   "dina",
	Short: "Dina CLI",
	Long: `Dina CLI – deploy applications, manage apps and environment variables, view logs, configure custom hostnames, and install AI agent skills for the Dina platform.

Report bugs or send feedback with ` + "`dina feedback bug`" + ` / ` + "`dina feedback`" + ` —
or file issues at https://github.com/dinacomputer/cli/issues.`,
	PersistentPreRun: runSkillCheck,
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&noInput, "no-input", false, "Disable interactive prompts (fail if input is required)")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Suppress informational progress messages on stderr")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Print debug output (also honors DEBUG=1)")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable colorized output (also honors NO_COLOR and TERM=dumb)")
}

// runSkillCheck surfaces a warning on stderr when installed skills are
// outdated. Skipped for commands where the warning would be noise (install,
// version) and for the skill-update command itself once we add one.
func runSkillCheck(cmd *cobra.Command, _ []string) {
	switch cmd.Name() {
	case "install", "version", "help", "doctor":
		return
	}
	if Quiet() {
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
func NoInput() bool { return noInput }

// Quiet reports whether informational progress messages should be suppressed.
func Quiet() bool { return quiet }

// Debug reports whether verbose debug output should be emitted. Honors the
// --debug flag and a truthy DEBUG env var.
func Debug() bool {
	if debug {
		return true
	}
	switch os.Getenv("DEBUG") {
	case "", "0", "false", "False", "FALSE":
		return false
	}
	return true
}

// NoColor reports whether the --no-color flag was passed. (The color helper
// also consults NO_COLOR and TERM=dumb for the final decision.)
func NoColor() bool { return noColor }

// Infof writes an informational progress message to stderr, suppressed by
// --quiet. Use for "Fetching...", "Uploading...", etc. — not for results,
// warnings, or errors.
func Infof(format string, args ...any) {
	if quiet {
		return
	}
	fmt.Fprintf(os.Stderr, format, args...)
}

// Infoln is Infof with an appended newline.
func Infoln(args ...any) {
	if quiet {
		return
	}
	fmt.Fprintln(os.Stderr, args...)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		if msg := err.Error(); msg != "" {
			fmt.Fprintln(os.Stderr, msg)
		}
		os.Exit(1)
	}
}
