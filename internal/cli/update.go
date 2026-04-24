package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/dinacomputer/cli/internal/api"
	"github.com/spf13/cobra"
)

const (
	githubRepo    = "dinacomputer/cli"
	githubAPIBase = "https://api.github.com/repos/" + githubRepo
)

type githubRelease struct {
	TagName string        `json:"tag_name"`
	Assets  []githubAsset `json:"assets"`
}

type githubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

var updateCheck bool

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update dina to the latest version",
	Long: `Download the latest release binary from GitHub and replace the running dina binary in place.

Releases are fetched from https://github.com/dinacomputer/cli/releases. The
new binary is extracted to a temp dir and atomically renamed over the current
binary (on Unix) or swapped via .old/.new shuffling (on Windows).

Pass --check to see whether an update is available without installing anything.`,
	Example: `  # check for and install the latest version
  dina update

  # only check — don't install
  dina update --check`,
	GroupID: groupCLI,
	RunE: func(cmd *cobra.Command, args []string) error {
		Infoln("Checking for updates...")
		latest, err := fetchLatestRelease()
		if err != nil {
			return fmt.Errorf("failed to check for updates: %w", err)
		}

		latestVersion := strings.TrimPrefix(latest.TagName, "v")
		currentVersion := strings.TrimPrefix(Version, "v")

		if latestVersion == currentVersion {
			fmt.Printf("Up to date (v%s).\n", currentVersion)
			return nil
		}

		if updateCheck {
			fmt.Printf("Update available: v%s → v%s\n", currentVersion, latestVersion)
			fmt.Println("Run 'dina update' to install the latest version.")
			return nil
		}

		Infof("New version available: v%s (current: v%s)\n", latestVersion, currentVersion)

		assetName := fmt.Sprintf("dina_%s_%s_%s.tar.gz", latestVersion, runtime.GOOS, runtime.GOARCH)
		if runtime.GOOS == "windows" {
			assetName = fmt.Sprintf("dina_%s_%s_%s.zip", latestVersion, runtime.GOOS, runtime.GOARCH)
		}

		var downloadURL string
		for _, a := range latest.Assets {
			if a.Name == assetName {
				downloadURL = a.BrowserDownloadURL
				break
			}
		}
		if downloadURL == "" {
			return fmt.Errorf("no release asset found for %s/%s", runtime.GOOS, runtime.GOARCH)
		}

		Infof("Downloading %s...\n", assetName)

		tmpDir, err := os.MkdirTemp("", "dina-update-*")
		if err != nil {
			return err
		}
		defer os.RemoveAll(tmpDir)

		archivePath := tmpDir + "/" + assetName
		if err := downloadFile(downloadURL, archivePath); err != nil {
			return fmt.Errorf("download failed: %w", err)
		}

		// Extract archive
		if runtime.GOOS == "windows" {
			// Use PowerShell to extract zip on Windows
			extractCmd := exec.Command("powershell", "-Command",
				fmt.Sprintf("Expand-Archive -Path '%s' -DestinationPath '%s'", archivePath, tmpDir))
			if out, err := extractCmd.CombinedOutput(); err != nil {
				return fmt.Errorf("extract failed: %s: %w", string(out), err)
			}
		} else {
			extractCmd := exec.Command("tar", "-xzf", archivePath, "-C", tmpDir)
			if out, err := extractCmd.CombinedOutput(); err != nil {
				return fmt.Errorf("extract failed: %s: %w", string(out), err)
			}
		}

		// Find current binary path
		binPath, err := os.Executable()
		if err != nil {
			return fmt.Errorf("cannot determine binary path: %w", err)
		}

		newBin := tmpDir + "/dina"
		if runtime.GOOS == "windows" {
			newBin += ".exe"
		}

		// Replace the current binary
		if err := replaceBinary(binPath, newBin); err != nil {
			return fmt.Errorf("failed to replace binary: %w", err)
		}

		Infof("Updated to v%s successfully!\n", latestVersion)
		return nil
	},
}


func init() {
	updateCmd.Flags().BoolVar(&updateCheck, "check", false, "Only check if an update is available; don't install")
	rootCmd.AddCommand(updateCmd)
}

func fetchLatestRelease() (*githubRelease, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", githubAPIBase+"/releases/latest", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", api.UserAgent)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}
	return &release, nil
}

func downloadFile(url, dest string) error {
	client := &http.Client{Timeout: 120 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", api.UserAgent)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download returned %d", resp.StatusCode)
	}

	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return err
}

func replaceBinary(target, source string) error {
	newBin, err := os.ReadFile(source)
	if err != nil {
		return err
	}

	// Get original file permissions
	info, err := os.Stat(target)
	if err != nil {
		return err
	}

	// Write new binary to a temp file next to the target, then rename (atomic on most OS)
	tmpPath := target + ".new"
	if err := os.WriteFile(tmpPath, newBin, info.Mode()); err != nil {
		return err
	}

	// Rename is atomic on Unix; on Windows we need to remove first
	if runtime.GOOS == "windows" {
		oldPath := target + ".old"
		_ = os.Remove(oldPath)
		if err := os.Rename(target, oldPath); err != nil {
			return err
		}
		if err := os.Rename(tmpPath, target); err != nil {
			return err
		}
		_ = os.Remove(oldPath)
		return nil
	}

	return os.Rename(tmpPath, target)
}
