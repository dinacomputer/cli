// Package config manages the Dina CLI's persistent non-secret configuration:
// CLI version last seen, list of installed skills, and check timestamps.
//
// Secrets (OAuth credentials) live separately in internal/auth.
package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"
)

const configDirName = "dina"

// Config holds the persisted CLI configuration.
type Config struct {
	// CLIVersion records the CLI version that last wrote this file. An empty
	// or stale value triggers a skill re-discovery on next run.
	CLIVersion string `json:"cli_version,omitempty"`

	// LastSkillCheck is the timestamp of the last skill-update check. Used
	// to throttle subsequent checks.
	LastSkillCheck time.Time `json:"last_skill_check,omitempty"`

	// Skills lists the installations recorded by `dina install --skills`.
	Skills []InstalledSkill `json:"skills,omitempty"`
}

// InstalledSkill records where a SKILL.md was written.
type InstalledSkill struct {
	Agent string `json:"agent"` // agent slug (e.g. "claude-code")
	Scope string `json:"scope"` // "project" or "user"
	Path  string `json:"path"`  // absolute path to SKILL.md
}

func configDir() (string, error) {
	dir := os.Getenv("XDG_CONFIG_HOME")
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		dir = filepath.Join(home, ".config")
	}
	return filepath.Join(dir, configDirName), nil
}

func configPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

// Load reads the config file. If it doesn't exist, returns an empty Config.
func Load() (*Config, error) {
	p, err := configPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &Config{}, nil
		}
		return nil, err
	}
	var c Config
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

// Save writes the config file, creating parent directories as needed.
func Save(c *Config) error {
	p, err := configPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0o600)
}
