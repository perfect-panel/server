package updater

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/perfect-panel/server/pkg/constant"
)

const (
	githubAPIURL = "https://api.github.com/repos/OmnTeam/server/releases/latest"
	githubRelURL = "https://github.com/OmnTeam/server/releases"
)

// Release represents a GitHub release
type Release struct {
	TagName    string  `json:"tag_name"`
	Name       string  `json:"name"`
	Body       string  `json:"body"`
	Draft      bool    `json:"draft"`
	Prerelease bool    `json:"prerelease"`
	Assets     []Asset `json:"assets"`
	HTMLURL    string  `json:"html_url"`
}

// Asset represents a release asset
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

// Updater handles auto-update functionality
type Updater struct {
	CurrentVersion string
	Owner          string
	Repo           string
	HTTPClient     *http.Client
}

// NewUpdater creates a new updater instance
func NewUpdater() *Updater {
	return &Updater{
		CurrentVersion: constant.Version,
		Owner:          "OmnTeam",
		Repo:           "server",
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CheckForUpdates checks if a new version is available
func (u *Updater) CheckForUpdates() (*Release, bool, error) {
	req, err := http.NewRequest("GET", githubAPIURL, nil)
	if err != nil {
		return nil, false, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := u.HTTPClient.Do(req)
	if err != nil {
		return nil, false, fmt.Errorf("failed to fetch release info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, false, fmt.Errorf("failed to decode response: %w", err)
	}

	// Skip draft and prerelease versions
	if release.Draft || release.Prerelease {
		return nil, false, nil
	}

	// Compare versions
	hasUpdate := u.compareVersions(release.TagName, u.CurrentVersion)
	return &release, hasUpdate, nil
}

// compareVersions compares two version strings
// Returns true if newVersion is newer than currentVersion
func (u *Updater) compareVersions(newVersion, currentVersion string) bool {
	// Remove 'v' prefix if present
	newVersion = strings.TrimPrefix(newVersion, "v")
	currentVersion = strings.TrimPrefix(currentVersion, "v")

	// Handle "unknown version" case
	if currentVersion == "unknown version" || currentVersion == "" {
		return true
	}

	return newVersion != currentVersion
}

// Download downloads the appropriate binary for the current platform
func (u *Updater) Download(release *Release) (string, error) {
	assetName := u.getAssetName()

	var targetAsset *Asset
	for _, asset := range release.Assets {
		if asset.Name == assetName {
			targetAsset = &asset
			break
		}
	}

	if targetAsset == nil {
		return "", fmt.Errorf("no suitable asset found for %s", assetName)
	}

	// Create temp directory
	tempDir, err := os.MkdirTemp("", "ppanel-update-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Download the file
	resp, err := u.HTTPClient.Get(targetAsset.BrowserDownloadURL)
	if err != nil {
		os.RemoveAll(tempDir)
		return "", fmt.Errorf("failed to download asset: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		os.RemoveAll(tempDir)
		return "", fmt.Errorf("failed to download: status code %d", resp.StatusCode)
	}

	// Read the entire file into memory
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		os.RemoveAll(tempDir)
		return "", fmt.Errorf("failed to read download: %w", err)
	}

	// Extract the binary
	binaryPath, err := u.extractBinary(data, tempDir, assetName)
	if err != nil {
		os.RemoveAll(tempDir)
		return "", fmt.Errorf("failed to extract binary: %w", err)
	}

	return binaryPath, nil
}

// getAssetName returns the expected asset name for the current platform
func (u *Updater) getAssetName() string {
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	// Capitalize first letter of OS
	osName := strings.Title(goos)

	// Map architecture names to match goreleaser output
	archName := goarch
	switch goarch {
	case "amd64":
		archName = "x86_64"
	case "386":
		archName = "i386"
	}

	// Format: ppanel-server-{Version}-{Os}-{Arch}.{ext}
	ext := "tar.gz"
	if goos == "windows" {
		ext = "zip"
	}

	return fmt.Sprintf("ppanel-server-%s-%s-%s.%s", u.CurrentVersion, osName, archName, ext)
}

// extractBinary extracts the binary from the downloaded archive
func (u *Updater) extractBinary(data []byte, destDir, assetName string) (string, error) {
	if strings.HasSuffix(assetName, ".zip") {
		return u.extractZip(data, destDir)
	}
	return u.extractTarGz(data, destDir)
}

// extractZip extracts a zip archive
func (u *Updater) extractZip(data []byte, destDir string) (string, error) {
	reader := bytes.NewReader(data)
	zipReader, err := zip.NewReader(reader, int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("failed to create zip reader: %w", err)
	}

	var binaryPath string
	for _, file := range zipReader.File {
		// Look for the binary file
		if strings.Contains(file.Name, "ppanel-server") && !strings.Contains(file.Name, "/") {
			binaryPath = filepath.Join(destDir, filepath.Base(file.Name))

			rc, err := file.Open()
			if err != nil {
				return "", fmt.Errorf("failed to open file in zip: %w", err)
			}
			defer rc.Close()

			outFile, err := os.OpenFile(binaryPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
			if err != nil {
				return "", fmt.Errorf("failed to create output file: %w", err)
			}
			defer outFile.Close()

			if _, err := io.Copy(outFile, rc); err != nil {
				return "", fmt.Errorf("failed to write file: %w", err)
			}

			return binaryPath, nil
		}
	}

	return "", fmt.Errorf("binary not found in archive")
}

// extractTarGz extracts a tar.gz archive
func (u *Updater) extractTarGz(data []byte, destDir string) (string, error) {
	reader := bytes.NewReader(data)
	gzReader, err := gzip.NewReader(reader)
	if err != nil {
		return "", fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	var binaryPath string
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("failed to read tar: %w", err)
		}

		// Look for the binary file
		if strings.Contains(header.Name, "ppanel-server") && !strings.Contains(header.Name, "/") {
			binaryPath = filepath.Join(destDir, filepath.Base(header.Name))

			outFile, err := os.OpenFile(binaryPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
			if err != nil {
				return "", fmt.Errorf("failed to create output file: %w", err)
			}
			defer outFile.Close()

			if _, err := io.Copy(outFile, tarReader); err != nil {
				return "", fmt.Errorf("failed to write file: %w", err)
			}

			return binaryPath, nil
		}
	}

	return "", fmt.Errorf("binary not found in archive")
}

// Apply applies the update by replacing the current binary
func (u *Updater) Apply(newBinaryPath string) error {
	// Get current executable path
	currentPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get current executable path: %w", err)
	}

	// Resolve symlinks
	currentPath, err = filepath.EvalSymlinks(currentPath)
	if err != nil {
		return fmt.Errorf("failed to resolve symlinks: %w", err)
	}

	// Create backup
	backupPath := currentPath + ".backup"
	if err := u.copyFile(currentPath, backupPath); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	// Replace the binary
	if err := u.replaceFile(newBinaryPath, currentPath); err != nil {
		// Restore backup on failure
		u.copyFile(backupPath, currentPath)
		return fmt.Errorf("failed to replace binary: %w", err)
	}

	// Remove backup on success
	os.Remove(backupPath)

	return nil
}

// copyFile copies a file from src to dst
func (u *Updater) copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	return dstFile.Sync()
}

// replaceFile replaces dst with src
func (u *Updater) replaceFile(src, dst string) error {
	// On Windows, we need to rename the old file first
	if runtime.GOOS == "windows" {
		oldPath := dst + ".old"
		if err := os.Rename(dst, oldPath); err != nil {
			return err
		}
		defer os.Remove(oldPath)
	}

	// Copy the new file
	if err := u.copyFile(src, dst); err != nil {
		return err
	}

	return nil
}

// Update performs the complete update process
func (u *Updater) Update() error {
	// Check for updates
	release, hasUpdate, err := u.CheckForUpdates()
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	if !hasUpdate {
		return fmt.Errorf("already running the latest version")
	}

	fmt.Printf("New version available: %s\n", release.TagName)
	fmt.Printf("Downloading update...\n")

	// Download the update
	binaryPath, err := u.Download(release)
	if err != nil {
		return fmt.Errorf("failed to download update: %w", err)
	}
	defer os.RemoveAll(filepath.Dir(binaryPath))

	fmt.Printf("Applying update...\n")

	// Apply the update
	if err := u.Apply(binaryPath); err != nil {
		return fmt.Errorf("failed to apply update: %w", err)
	}

	fmt.Printf("Update completed successfully! Please restart the application.\n")
	return nil
}
