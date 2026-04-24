package skills

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/dinacomputer/cli/internal/config"
)

// checkInterval is how often we re-check for outdated skills.
const checkInterval = 24 * time.Hour

// CheckResult summarizes an update check.
type CheckResult struct {
	Outdated []config.InstalledSkill
}

// CheckIfDue runs a skill freshness check when warranted and returns any
// outdated installations. The caller decides how to surface the result.
//
// The check is skipped when all of these are true: the config's cli_version
// matches the current version, there are no tracked skill installations, and
// the last check was less than checkInterval ago. On a version bump, user-scope
// installation paths are scanned so that pre-tracking installs are picked up.
//
// Any I/O error is swallowed; the check is best-effort and must never block
// the user's actual command.
func CheckIfDue(cliVersion string) *CheckResult {
	cfg, err := config.Load()
	if err != nil {
		return nil
	}

	versionStale := cfg.CLIVersion != cliVersion
	hasInstalls := len(cfg.Skills) > 0

	if !versionStale && !hasInstalls {
		return nil
	}
	if !cfg.LastSkillCheck.IsZero() && time.Since(cfg.LastSkillCheck) < checkInterval {
		return nil
	}

	if versionStale {
		cfg.Skills = mergeSkills(cfg.Skills, discoverUserSkills())
	}

	current := currentHash()
	var outdated, present []config.InstalledSkill
	for _, s := range cfg.Skills {
		h, err := fileHash(s.Path)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			present = append(present, s)
			continue
		}
		present = append(present, s)
		if h != current {
			outdated = append(outdated, s)
		}
	}

	cfg.Skills = present
	cfg.CLIVersion = cliVersion
	cfg.LastSkillCheck = time.Now()
	_ = config.Save(cfg)

	if len(outdated) == 0 {
		return nil
	}
	return &CheckResult{Outdated: outdated}
}

func currentHash() string {
	sum := sha256.Sum256([]byte(SkillMD()))
	return hex.EncodeToString(sum[:])
}

func fileHash(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

// discoverUserSkills walks known user-scope installation paths and reports any
// SKILL.md files already in place. Project-scope installs cannot be discovered
// retroactively (we don't know the user's projects).
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
