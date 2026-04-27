// Package doctor runs diagnostic checks on the Dina CLI's local state and,
// where possible, repairs them.
package doctor

// Status is the outcome of a diagnostic check.
type Status int

const (
	StatusOK Status = iota
	StatusWarn
	StatusFail
)

func (s Status) String() string {
	switch s {
	case StatusOK:
		return "ok"
	case StatusWarn:
		return "warn"
	case StatusFail:
		return "fail"
	}
	return "?"
}

// Result is the outcome of running a single check.
type Result struct {
	Name    string
	Status  Status
	Summary string
	Details []string

	// FixHint is displayed when Status != OK and the check cannot auto-repair;
	// typically the command the user should run to resolve it.
	FixHint string

	// AfterFix is displayed after a successful auto-repair. Use this to tell
	// the user about any out-of-band step the fix can't do itself (e.g.,
	// "restart your editor").
	AfterFix string

	// fix repairs the issue in place, if possible. nil means the check cannot
	// auto-repair (the FixHint is what the user must do).
	fix func() error
}

// Fixable reports whether Fix() can attempt an auto-repair.
func (r *Result) Fixable() bool {
	return r.fix != nil
}

// Fix runs the attached repair routine. It is the caller's responsibility to
// re-run the diagnosis afterward if they need the updated status.
func (r *Result) Fix() error {
	if r.fix == nil {
		return nil
	}
	return r.fix()
}

// Check is one diagnostic probe.
type Check struct {
	Name     string
	Diagnose func() Result
}

// RunAll runs every check in order and returns the results.
func RunAll(checks []Check) []Result {
	out := make([]Result, 0, len(checks))
	for _, c := range checks {
		r := c.Diagnose()
		r.Name = c.Name
		out = append(out, r)
	}
	return out
}
