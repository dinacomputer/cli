package skills

import (
	"os"
	"path/filepath"

	"github.com/dinacomputer/cli/internal/config"
)

// RegisterDiscoveredSkills picks up any user-scope SKILL.md files installed
// before config tracking existed (or by a previous CLI version) and records
// them in config.
//
// It is a no-op after the first run on a given CLI version: the stored
// cli_version is compared against the running binary's and nothing happens
// when they match.
//
// Project-scope installs cannot be discovered retroactively — we don't know
// the user's projects.
//
// Errors are swallowed; the caller runs this best-effort.
func RegisterDiscoveredSkills(cliVersion string) {
	cfg, err := config.Load()
	if err != nil {
		return
	}
	if cfg.CLIVersion == cliVersion {
		return
	}

	cfg.Skills = mergeSkills(cfg.Skills, discoverUserSkills())
	cfg.CLIVersion = cliVersion
	_ = config.Save(cfg)
}

// discoverUserSkills walks known user-scope installation paths and reports
// any SKILL.md files already in place.
func discoverUserSkills() []config.InstalledSkill {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	var found []config.InstalledSkill
	for _, a := range SupportedAgents {
		if a.UserDir == "" {
			continue
		}
		path := filepath.Join(home, a.UserDir, skillDirName, "SKILL.md")
		if _, err := os.Stat(path); err == nil {
			found = append(found, config.InstalledSkill{
				Agent: a.Slug,
				Scope: "user",
				Path:  path,
			})
		}
	}
	return found
}

func mergeSkills(a, b []config.InstalledSkill) []config.InstalledSkill {
	seen := make(map[string]bool, len(a))
	out := make([]config.InstalledSkill, 0, len(a)+len(b))
	for _, s := range a {
		seen[s.Path] = true
		out = append(out, s)
	}
	for _, s := range b {
		if !seen[s.Path] {
			out = append(out, s)
		}
	}
	return out
}
