package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/dinacomputer/cli/internal/term"
)

// confirmYesNo prompts for a y/N confirmation. Returns nil if confirmed, an
// error otherwise. If force is true, the prompt is skipped and nil is returned.
// Respects --no-input: if input is not a TTY or --no-input is set, the call
// fails with a message pointing to --force.
func confirmYesNo(prompt string, force bool) error {
	if force {
		return nil
	}
	if noInput || !term.IsStdinTTY() {
		return fmt.Errorf("refusing to proceed without confirmation — re-run with --force to skip this prompt")
	}

	fmt.Fprintf(os.Stderr, "%s [y/N]: ", prompt)
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("reading confirmation: %w", err)
	}
	switch strings.ToLower(strings.TrimSpace(line)) {
	case "y", "yes":
		return nil
	default:
		return fmt.Errorf("aborted")
	}
}

// confirmByName prompts the user to type the resource's name verbatim to
// confirm a severe action. If force is true, the prompt is skipped.
func confirmByName(resource, name string, force bool) error {
	if force {
		return nil
	}
	if noInput || !term.IsStdinTTY() {
		return fmt.Errorf("refusing to proceed without confirmation — re-run with --force to skip this prompt")
	}

	fmt.Fprintf(os.Stderr, "This will permanently delete %s %q. Type the name to confirm: ", resource, name)
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("reading confirmation: %w", err)
	}
	if strings.TrimSpace(line) != name {
		return fmt.Errorf("aborted: name did not match")
	}
	return nil
}
