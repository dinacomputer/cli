package auth

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"
)

const configDirName = "dina"

// Credentials holds the persisted OAuth state.
type Credentials struct {
	ClientID     string    `json:"client_id,omitempty"`
	ClientSecret string    `json:"client_secret,omitempty"`
	Issuer       string    `json:"issuer,omitempty"`
	APIBaseURL   string    `json:"api_base_url,omitempty"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// Expired returns true if the access token has expired (with a 30s buffer).
func (c *Credentials) Expired() bool {
	return time.Now().After(c.ExpiresAt.Add(-30 * time.Second))
}

// configDir returns ~/.config/dina (XDG-style).
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

func credentialsPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "auth.json"), nil
}

// LoadCredentials reads stored credentials from disk.
func LoadCredentials() (*Credentials, error) {
	p, err := credentialsPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var creds Credentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, err
	}
	return &creds, nil
}

// SaveCredentials writes credentials to disk.
func SaveCredentials(creds *Credentials) error {
	p, err := credentialsPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0o600)
}

// ClearCredentials removes stored credentials.
func ClearCredentials() error {
	p, err := credentialsPath()
	if err != nil {
		return err
	}
	err = os.Remove(p)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}
