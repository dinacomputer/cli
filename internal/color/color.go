// Package color provides minimal ANSI colorization helpers that no-op when
// color shouldn't be used.
//
// Color is disabled when:
//   - the target stream is not a TTY,
//   - the NO_COLOR environment variable is set to any non-empty value,
//   - TERM is "dumb",
//   - the CLI's --no-color flag is set (signalled via Disable()).
//
// Callers use Green("foo") etc. to wrap a string with ANSI escape codes; when
// color is disabled, the string is returned unchanged.
package color

import (
	"os"
	"sync/atomic"

	"github.com/dinacomputer/cli/internal/term"
)

const (
	reset  = "\x1b[0m"
	bold   = "\x1b[1m"
	dim    = "\x1b[2m"
	red    = "\x1b[31m"
	green  = "\x1b[32m"
	yellow = "\x1b[33m"
)

// disabled tracks the --no-color flag state (set once at startup).
var disabled atomic.Bool

// Disable turns off all color output, regardless of TTY or env vars. Called
// from the cli package when --no-color is passed.
func Disable() { disabled.Store(true) }

// enabledFor reports whether color should be written to the given stream.
func enabledFor(stream *os.File) bool {
	if disabled.Load() {
		return false
	}
	if v := os.Getenv("NO_COLOR"); v != "" {
		return false
	}
	if os.Getenv("TERM") == "dumb" {
		return false
	}
	// Check TTY-ness of the target stream.
	switch stream {
	case os.Stdout:
		return term.IsStdoutTTY()
	case os.Stderr:
		return term.IsStderrTTY()
	}
	return false
}

// StdoutEnabled reports whether color should be written to stdout.
func StdoutEnabled() bool { return enabledFor(os.Stdout) }

// StderrEnabled reports whether color should be written to stderr.
func StderrEnabled() bool { return enabledFor(os.Stderr) }

// wrapFor returns s wrapped in the given ANSI sequence iff color is enabled
// for the given stream.
func wrapFor(stream *os.File, seq, s string) string {
	if !enabledFor(stream) {
		return s
	}
	return seq + s + reset
}

// Green/Red/Yellow/Bold/Dim wrap s for stderr output. Stderr is the typical
// target for status output, warnings, and errors.
func Green(s string) string  { return wrapFor(os.Stderr, green, s) }
func Red(s string) string    { return wrapFor(os.Stderr, red, s) }
func Yellow(s string) string { return wrapFor(os.Stderr, yellow, s) }
func Bold(s string) string   { return wrapFor(os.Stderr, bold, s) }
func Dim(s string) string    { return wrapFor(os.Stderr, dim, s) }
