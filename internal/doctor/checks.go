package doctor

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dinacomputer/cli/internal/api"
	"github.com/dinacomputer/cli/internal/auth"
	"github.com/dinacomputer/cli/internal/config"
	"github.com/dinacomputer/cli/internal/skills"
)

// DefaultChecks returns the standard set of diagnostics, bound to the running
// CLI version.
func DefaultChecks(cliVersion string) []Check {
	return []Check{
		{Name: "authentication", Diagnose: diagnoseAuth},
		{Name: "installed skills", Diagnose: diagnoseSkills},
		{Name: "CLI version", Diagnose: func() Result { return diagnoseUpdate(cliVersion) }},
	}
}

// --- authentication ---

func diagnoseAuth() Result {
	creds, err := auth.LoadCredentials()
	if err != nil {
		return Result{Status: StatusFail, Summary: "could not read credentials", Details: []string{err.Error()}, FixHint: "dina auth login"}
	}
	if creds == nil || creds.AccessToken == "" {
		if tok := os.Getenv("DINA_API_TOKEN"); tok != "" {
			return Result{Status: StatusOK, Summary: "using DINA_API_TOKEN from environment"}
		}
		return Result{Status: StatusWarn, Summary: "not authenticated", FixHint: "dina auth login"}
	}
	if creds.Expired() {
		return Result{Status: StatusWarn, Summary: "access token expired (will refresh on next API call)"}
	}
	return Result{Status: StatusOK, Summary: fmt.Sprintf("logged in, token valid until %s", creds.ExpiresAt.Format("2006-01-02 15:04"))}
}

// --- installed skills ---

func diagnoseSkills() Result {
	cfg, err := config.Load()
	if err != nil {
		return Result{Status: StatusWarn, Summary: "could not read config", Details: []string{err.Error()}}
	}
	if len(cfg.Skills) == 0 {
		return Result{Status: StatusOK, Summary: "no skills installed"}
	}

	var outdated, missing []config.InstalledSkill
	for _, s := range cfg.Skills {
		data, err := os.ReadFile(s.Path)
		if os.IsNotExist(err) {
			missing = append(missing, s)
			continue
		}
		if err != nil {
			continue
		}
		sum := sha256.Sum256(data)
		if hex.EncodeToString(sum[:]) != canonicalSkillHash() {
			outdated = append(outdated, s)
		}
	}

	if len(outdated) == 0 && len(missing) == 0 {
		return Result{Status: StatusOK, Summary: fmt.Sprintf("%d installed, all current", len(cfg.Skills))}
	}

	var details []string
	for _, s := range outdated {
		details = append(details, fmt.Sprintf("outdated: %s (%s) %s", s.Agent, s.Scope, s.Path))
	}
	for _, s := range missing {
		details = append(details, fmt.Sprintf("missing file: %s (%s) %s", s.Agent, s.Scope, s.Path))
	}

	summary := strings.TrimSpace(fmt.Sprintf("%d outdated, %d missing", len(outdated), len(missing)))

	fix := func() error { return fixSkills(outdated, missing) }

	return Result{
		Status:  StatusWarn,
		Summary: summary,
		Details: details,
		fix:     fix,
	}
}

func canonicalSkillHash() string {
	sum := sha256.Sum256([]byte(skills.SkillMD()))
	return hex.EncodeToString(sum[:])
}

// fixSkills rewrites outdated skill files with the current canonical content
// and prunes missing entries from the config.
func fixSkills(outdated, missing []config.InstalledSkill) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	outdatedPaths := make(map[string]bool, len(outdated))
	for _, s := range outdated {
		outdatedPaths[s.Path] = true
	}
	missingPaths := make(map[string]bool, len(missing))
	for _, s := range missing {
		missingPaths[s.Path] = true
	}

	content := []byte(skills.SkillMD())
	for p := range outdatedPaths {
		if err := os.WriteFile(p, content, 0o644); err != nil {
			return fmt.Errorf("writing %s: %w", p, err)
		}
	}

	updated := cfg.Skills[:0]
	for _, s := range cfg.Skills {
		if missingPaths[s.Path] {
			continue
		}
		if outdatedPaths[s.Path] {
			s.Outdated = false
		}
		updated = append(updated, s)
	}
	cfg.Skills = updated
	cfg.LastSkillCheck = time.Now()
	return config.Save(cfg)
}

// --- CLI version ---

func diagnoseUpdate(cliVersion string) Result {
	latest, err := fetchLatestTag()
	if err != nil {
		return Result{Status: StatusWarn, Summary: "could not reach GitHub", Details: []string{err.Error()}}
	}
	current := strings.TrimPrefix(cliVersion, "v")
	latest = strings.TrimPrefix(latest, "v")
	if latest == current {
		return Result{Status: StatusOK, Summary: fmt.Sprintf("up to date (v%s)", current)}
	}
	return Result{
		Status:  StatusWarn,
		Summary: fmt.Sprintf("v%s available (current: v%s)", latest, current),
		FixHint: "dina update",
	}
}

func fetchLatestTag() (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", "https://api.github.com/repos/dinacomputer/cli/releases/latest", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", api.UserAgent)
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub returned %d", resp.StatusCode)
	}
	var body struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", err
	}
	return body.TagName, nil
}
