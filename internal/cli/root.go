package cli

import (
	"fmt"
	"os"

	"github.com/dinacomputer/cli/internal/api"
	"github.com/dinacomputer/cli/internal/color"
	"github.com/dinacomputer/cli/internal/skills"
	"github.com/spf13/cobra"
)

var (
	noInput    bool
	quiet      bool
	debug      bool
	noColor    bool
	output     string
	validOutputs = map[string]bool{"text": true, "json": true}
)

var rootCmd = &cobra.Command{
	Use:   "dina",
	Short: "Dina CLI",
	Long: `Dina CLI – deploy applications, manage apps and environment variables, view logs, configure custom hostnames, and install AI agent skills for the Dina platform.

Report bugs or send feedback with ` + "`dina feedback bug`" + ` / ` + "`dina feedback`" + ` —
or file issues at https://github.com/dinacomputer/cli/issues.`,
	// SilenceUsage suppresses the usage dump when a RunE returns an error.
	// Cobra still shows usage for flag-parsing errors, which is the right
	// behavior: the user needs help fixing syntax, not for API errors.
	SilenceUsage: true,
	// SilenceErrors lets us own the "error:" prefix via Execute() instead
	// of getting cobra's default "Error: ..." line plus our own duplicate.
	SilenceErrors:    true,
	PersistentPreRun: persistentPreRun,
}

func persistentPreRun(cmd *cobra.Command, args []string) {
	if noColor {
		color.Disable()
	}
	if Debug() {
		api.Debug = true
	}
	runSkillCheck(cmd, args)
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&noInput, "no-input", false, "Disable interactive prompts (fail if input is required)")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Suppress informational progress messages on stderr")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Print debug output (also honors DEBUG=1)")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable colorized output (also honors NO_COLOR and TERM=dumb)")
	rootCmd.PersistentFlags().StringVarP(&output, "output", "o", "text", "Output format: text or json")
}

// JSONOutput reports whether -o json was passed. Commands that have a JSON
// renderer should check this and emit JSON on stdout instead of the
// human-readable form.
func JSONOutput() bool { return output == "json" }

// ValidateOutput returns an error if -o is set to an unsupported value. Call
// from RunE at the top of commands that respect --output.
func ValidateOutput() error {
	if !validOutputs[output] {
		return fmt.Errorf("invalid --output %q: expected 'text' or 'json'", output)
	}
	return nil
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
			fmt.Fprintf(os.Stderr, "%s %s\n", color.Red("error:"), msg)
		}
		os.Exit(1)
	}
}
