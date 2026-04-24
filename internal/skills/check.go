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

// recheckInterval is how often we re-hash files to see if an outdated skill
// was fixed manually. Fresh installs via `dina install --skills` update the
// config directly and don't rely on this interval.
const recheckInterval = 24 * time.Hour

// CheckResult summarizes an update check.
type CheckResult struct {
	Outdated []config.InstalledSkill
}

// CheckIfDue returns any skills recorded as outdated in the config. It uses
// the cached per-skill state on every call (cheap — no file I/O), and only
// re-hashes when one of the following is true:
//
//   - The CLI version differs from the value in config (binary was upgraded).
//   - At least one skill is cached as outdated and the last re-check was more
//     than recheckInterval ago (probing to see if the user fixed it).
//
// Per spec: the check is skipped entirely when no skills are tracked and the
// CLI version in config matches the running binary — new users who never
// installed a skill are never bothered.
//
// All I/O errors are swallowed; the check is best-effort and must never block
// the user's real command.
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

	outdatedCached := anyOutdated(cfg.Skills)
	shouldRecheck := versionStale ||
		(outdatedCached && time.Since(cfg.LastSkillCheck) >= recheckInterval)

	if shouldRecheck {
		if versionStale {
			cfg.Skills = mergeSkills(cfg.Skills, discoverUserSkills())
		}
		cfg.Skills = refreshSkillState(cfg.Skills)
		cfg.CLIVersion = cliVersion
		cfg.LastSkillCheck = time.Now()
		_ = config.Save(cfg)
	}

	var outdated []config.InstalledSkill
	for _, s := range cfg.Skills {
		if s.Outdated {
			outdated = append(outdated, s)
		}
	}
	if len(outdated) == 0 {
		return nil
	}
	return &CheckResult{Outdated: outdated}
}

// refreshSkillState hashes each skill file and updates its Outdated flag.
// Missing files are dropped. Unreadable files retain their previous state.
func refreshSkillState(skills []config.InstalledSkill) []config.InstalledSkill {
	current := currentHash()
	out := skills[:0]
	for _, s := range skills {
		h, err := fileHash(s.Path)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			out = append(out, s)
			continue
		}
		s.Outdated = h != current
		out = append(out, s)
	}
	return out
}

func anyOutdated(skills []config.InstalledSkill) bool {
	for _, s := range skills {
		if s.Outdated {
			return true
		}
	}
	return false
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
