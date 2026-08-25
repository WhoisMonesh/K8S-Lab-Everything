package update

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

var (
	Version   = "dev"
	GitCommit = "unknown"
)

const (
	githubOwner = "WhoisMonesh"
	githubRepo  = "K8S-Lab-Everything"
	checkURL    = "https://api.github.com/repos/" + githubOwner + "/" + githubRepo + "/releases/latest"
)

type githubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

type cacheData struct {
	LastCheck time.Time `json:"last_check"`
	Version   string    `json:"version"`
}

func cachePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cka-lab-runner-update-cache")
}

func isCacheValid() bool {
	data, err := os.ReadFile(cachePath())
	if err != nil {
		return false
	}
	var c cacheData
	if json.Unmarshal(data, &c) != nil {
		return false
	}
	return time.Since(c.LastCheck) < 24*time.Hour
}

func saveCache(version string) {
	c := cacheData{LastCheck: time.Now(), Version: version}
	data, _ := json.Marshal(c)
	os.WriteFile(cachePath(), data, 0644)
}

func printUpdateMessage(latest string) {
	green := "\033[32m"
	cyan := "\033[36m"
	bold := "\033[1m"
	reset := "\033[0m"
	fmt.Println()
	fmt.Printf("%s%s=== UPDATE AVAILABLE ===%s\n", bold, green, reset)
	fmt.Printf("  Current version: %s%s%s\n", cyan, Version, reset)
	fmt.Printf("  Latest version:  %s%s%s\n", green, latest, reset)
	fmt.Println()
	fmt.Printf("  Run %scka-lab-runner update%s to install the latest version.\n", bold, reset)
	fmt.Println()
}

func isNewer(latest, current string) bool {
	lParts := strings.Split(latest, ".")
	cParts := strings.Split(current, ".")
	for i := 0; i < len(lParts) && i < len(cParts); i++ {
		if lParts[i] > cParts[i] {
			return true
		}
		if lParts[i] < cParts[i] {
			return false
		}
	}
	return len(lParts) > len(cParts)
}

func fetchLatestVersion() (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(checkURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}
	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}
	return strings.TrimPrefix(release.TagName, "v"), nil
}

func getDownloadURL(version string) (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(checkURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}
	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}
	arch := "amd64"
	if runtime.GOARCH == "arm64" {
		arch = "arm64"
	}
	expectedName := fmt.Sprintf("cka-lab-runner-%s-%s%s", runtime.GOOS, arch, ext)
	for _, asset := range release.Assets {
		if asset.Name == expectedName {
			return asset.BrowserDownloadURL, nil
		}
	}
	return "", fmt.Errorf("no binary found for %s/%s", runtime.GOOS, arch)
}

func downloadFile(url, dest string) error {
	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	return err
}

// CheckForUpdate checks GitHub for a newer version in the background.
func CheckForUpdate() {
	if Version == "dev" || os.Getenv("CKA_LAB_SKIP_UPDATE_CHECK") != "" {
		return
	}
	if isCacheValid() {
		return
	}
	go func() {
		latestVersion, err := fetchLatestVersion()
		if err != nil {
			return
		}
		if isNewer(latestVersion, Version) {
			printUpdateMessage(latestVersion)
		}
		saveCache(latestVersion)
	}()
}

// SelfUpdate downloads and installs the latest release binary.
func SelfUpdate() error {
	fmt.Println("Checking for latest release...")
	latestVersion, err := fetchLatestVersion()
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}
	if !isNewer(latestVersion, Version) {
		fmt.Printf("Already on latest version (%s)\n", Version)
		return nil
	}
	fmt.Printf("New version available: %s (current: %s)\n", latestVersion, Version)
	fmt.Println("Downloading...")
	downloadURL, err := getDownloadURL(latestVersion)
	if err != nil {
		return fmt.Errorf("failed to find download URL: %w", err)
	}
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("failed to resolve executable path: %w", err)
	}
	tmpFile := execPath + ".tmp"
	if err := downloadFile(downloadURL, tmpFile); err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("failed to download: %w", err)
	}
	if err := os.Chmod(tmpFile, 0755); err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("failed to set permissions: %w", err)
	}
	if err := os.Rename(tmpFile, execPath); err != nil {
		cmd := exec.Command("mv", tmpFile, execPath)
		if output, err := cmd.CombinedOutput(); err != nil {
			os.Remove(tmpFile)
			return fmt.Errorf("failed to replace binary: %s: %w", string(output), err)
		}
	}
	fmt.Printf("Updated successfully! (%s -> %s)\n", Version, latestVersion)
	fmt.Println("Run 'cka-lab-runner --version' to verify.")
	return nil
}

// GetVersion returns the current version string.
func GetVersion() string {
	return Version
}
