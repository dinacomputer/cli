// Package term provides small helpers for detecting whether the standard
// streams are attached to a terminal.
package term

import (
	"os"

	"github.com/mattn/go-isatty"
)

// IsStdinTTY reports whether stdin is an interactive terminal.
func IsStdinTTY() bool {
	return isTTY(os.Stdin.Fd())
}

// IsStdoutTTY reports whether stdout is an interactive terminal.
func IsStdoutTTY() bool {
	return isTTY(os.Stdout.Fd())
}

// IsStderrTTY reports whether stderr is an interactive terminal.
func IsStderrTTY() bool {
	return isTTY(os.Stderr.Fd())
}

func isTTY(fd uintptr) bool {
	return isatty.IsTerminal(fd) || isatty.IsCygwinTerminal(fd)
}
