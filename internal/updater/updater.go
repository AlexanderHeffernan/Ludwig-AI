package updater

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"archive/tar"
	"compress/gzip"
	"archive/zip"
)

type Release struct {
	TagName string `json:"tag_name"`
	Assets  []Asset `json:"assets"`
}

type Asset struct {
	Name string `json:"name"`
	DownloadURL string `json:"browser_download_url"`
}

const (
	repoOwner = "AlexanderHeffernan"
	repoName  = "Ludwig-AI"
	apiURL    = "https://api.github.com/repos/" + repoOwner + "/" + repoName + "/releases/latest"
)

// GetLatestVersion fetches the latest release version from GitHub
func GetLatestVersion() (string, error) {
	resp, err := http.Get(apiURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch latest release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", fmt.Errorf("failed to parse release: %w", err)
	}

	return release.TagName, nil
}

// CheckForUpdate returns true if a newer version is available
func CheckForUpdate(currentVersion string) (bool, string, error) {
	latestVersion, err := GetLatestVersion()
	if err != nil {
		return false, "", err
	}

	// Simple version comparison (strip 'v' prefix)
	current := strings.TrimPrefix(currentVersion, "v")
	latest := strings.TrimPrefix(latestVersion, "v")

	// If versions are equal or current is newer, no update needed
	if compareVersions(current, latest) >= 0 {
		return false, "", nil
	}

	return true, latestVersion, nil
}

// DownloadAndInstall downloads the latest release and replaces the current binary
func DownloadAndInstall(currentVersion string) error {
	latestVersion, err := GetLatestVersion()
	if err != nil {
		return err
	}

	// Check if update is actually needed
	isNewer, _, err := CheckForUpdate(currentVersion)
	if err != nil {
		return err
	}
	if !isNewer {
		fmt.Println("Already on the latest version (" + currentVersion + ")")
		return nil
	}

	// Get current executable path
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Fetch release info to find the right asset
	resp, err := http.Get(apiURL)
	if err != nil {
		return fmt.Errorf("failed to fetch release: %w", err)
	}
	defer resp.Body.Close()

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return fmt.Errorf("failed to parse release: %w", err)
	}

	// Find the right asset for current OS/arch
	osName, archName := getOSAndArch()
	var downloadURL, assetName string

	for _, asset := range release.Assets {
		if matchesAsset(asset.Name, osName, archName) {
			downloadURL = asset.DownloadURL
			assetName = asset.Name
			break
		}
	}

	if downloadURL == "" {
		return fmt.Errorf("no binary found for %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	// Download the binary
	fmt.Println("Downloading " + latestVersion + "...")
	newBinary, err := downloadFile(downloadURL)
	if err != nil {
		return err
	}
	defer os.Remove(newBinary)

	// Extract if needed (tar.gz or zip)
	extractedBinary := newBinary
	if strings.HasSuffix(assetName, ".tar.gz") {
		var err2 error
		extractedBinary, err2 = extractTarGz(newBinary)
		if err2 != nil {
			return err2
		}
		defer os.Remove(extractedBinary)
	} else if strings.HasSuffix(assetName, ".zip") {
		var err2 error
		extractedBinary, err2 = extractZip(newBinary)
		if err2 != nil {
			return err2
		}
		defer os.Remove(extractedBinary)
	}

	// Make executable
	if err := os.Chmod(extractedBinary, 0755); err != nil {
		return fmt.Errorf("failed to make binary executable: %w", err)
	}

	// Move to .new location (will be applied on next startup)
	newPath := exePath + ".new"
	if err := os.Rename(extractedBinary, newPath); err != nil {
		return fmt.Errorf("failed to prepare update: %w", err)
	}

	fmt.Println("Update ready! Please restart Ludwig to apply.")
	return nil
}

func downloadFile(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	tmpFile, err := os.CreateTemp("", "ludwig-update-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}

	_, err = io.Copy(tmpFile, resp.Body)
	tmpFile.Close()
	if err != nil {
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("failed to download: %w", err)
	}

	return tmpFile.Name(), nil
}

func extractTarGz(tarGzPath string) (string, error) {
	file, err := os.Open(tarGzPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	gr, err := gzip.NewReader(file)
	if err != nil {
		return "", err
	}
	defer gr.Close()

	tr := tar.NewReader(gr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		if header.Typeflag == tar.TypeReg && strings.Contains(header.Name, "ludwig") {
			tmpFile, err := os.CreateTemp("", "ludwig-bin-*")
			if err != nil {
				return "", err
			}
			defer tmpFile.Close()

			_, err = io.Copy(tmpFile, tr)
			if err != nil {
				return "", err
			}

			return tmpFile.Name(), nil
		}
	}

	return "", fmt.Errorf("ludwig binary not found in archive")
}

func extractZip(zipPath string) (string, error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", err
	}
	defer r.Close()

	for _, f := range r.File {
		if !f.FileInfo().IsDir() && strings.Contains(f.Name, "ludwig") {
			rc, err := f.Open()
			if err != nil {
				return "", err
			}
			defer rc.Close()

			tmpFile, err := os.CreateTemp("", "ludwig-bin-*")
			if err != nil {
				return "", err
			}
			defer tmpFile.Close()

			_, err = io.Copy(tmpFile, rc)
			if err != nil {
				return "", err
			}

			return tmpFile.Name(), nil
		}
	}

	return "", fmt.Errorf("ludwig binary not found in archive")
}

func getOSAndArch() (string, string) {
	var os, arch string

	switch runtime.GOOS {
	case "darwin":
		os = "Darwin"
	case "linux":
		os = "Linux"
	default:
		os = runtime.GOOS
	}

	switch runtime.GOARCH {
	case "amd64":
		arch = "x86_64"
	case "arm64":
		arch = "arm64"
	default:
		arch = runtime.GOARCH
	}

	return os, arch
}

func matchesAsset(assetName, os, arch string) bool {
	// Asset names are like: ludwig_v0.0.5_Darwin_arm64.tar.gz
	return strings.Contains(assetName, "_"+os+"_") &&
		strings.Contains(assetName, "_"+arch) &&
		(strings.HasSuffix(assetName, ".tar.gz") || strings.HasSuffix(assetName, ".zip"))
}

// ApplyPendingUpdate checks if there's a pending .new binary and applies it.
// Returns true if an update was applied (requiring a restart).
func ApplyPendingUpdate() (bool, error) {
	exePath, err := os.Executable()
	if err != nil {
		return false, nil // Can't apply update if we can't determine executable path
	}

	newPath := exePath + ".new"

	// Check if pending update exists
	if _, err := os.Stat(newPath); os.IsNotExist(err) {
		return false, nil // No pending update
	}

	// Replace the old binary with the new one
	if err := os.Rename(newPath, exePath); err != nil {
		return false, fmt.Errorf("failed to apply pending update: %w", err)
	}

	return true, nil
}

// compareVersions returns -1 if v1 < v2, 0 if equal, 1 if v1 > v2
// Simple comparison: split by dots and compare numeric parts
func compareVersions(v1, v2 string) int {
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	for i := 0; i < len(parts1) && i < len(parts2); i++ {
		// Simple numeric comparison (good enough for semver)
		var n1, n2 int
		fmt.Sscanf(parts1[i], "%d", &n1)
		fmt.Sscanf(parts2[i], "%d", &n2)

		if n1 < n2 {
			return -1
		} else if n1 > n2 {
			return 1
		}
	}

	if len(parts1) < len(parts2) {
		return -1
	} else if len(parts1) > len(parts2) {
		return 1
	}

	return 0
}
