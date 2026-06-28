package plugin

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

const (
	maxPluginArchiveBytes   int64 = 128 << 20
	maxPluginExtractedBytes int64 = 256 << 20
	maxPluginArchiveFiles         = 2048
)

// PluginInstallOptions controls how an uploaded plugin package is installed.
type PluginInstallOptions struct {
	Replace bool `json:"replace"`
	Enable  bool `json:"enable"`
}

// PluginInstallResult describes the result of installing an uploaded plugin.
type PluginInstallResult struct {
	Name       string           `json:"name"`
	Replaced   bool             `json:"replaced"`
	Enabled    bool             `json:"enabled"`
	Status     PluginStatus     `json:"status"`
	Plugin     PluginInfo       `json:"plugin"`
	Validation PluginValidation `json:"validation"`
}

// InstallPluginArchive installs a zip package that contains plugin.yaml and the WASM module.
func (m *Manager) InstallPluginArchive(ctx context.Context, r io.Reader, opts PluginInstallOptions) (PluginInstallResult, error) {
	var result PluginInstallResult
	if r == nil {
		return result, fmt.Errorf("plugin package is required")
	}

	pluginDir := m.pluginDirectory()
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		return result, fmt.Errorf("create plugin directory %q: %w", pluginDir, err)
	}

	tempDir, err := os.MkdirTemp("", "ppanel-plugin-upload-*")
	if err != nil {
		return result, fmt.Errorf("create upload temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	archivePath := filepath.Join(tempDir, "plugin.zip")
	if err := writeLimitedFile(archivePath, r, maxPluginArchiveBytes); err != nil {
		return result, err
	}

	stagingDir := filepath.Join(tempDir, "staging")
	if err := extractPluginZip(archivePath, stagingDir); err != nil {
		return result, err
	}

	pluginRoot, err := findPluginRoot(stagingDir)
	if err != nil {
		return result, err
	}

	manifest, err := ParseManifest(pluginRoot)
	if err != nil {
		return result, err
	}
	if err := ValidatePluginName(manifest.Name); err != nil {
		return result, err
	}
	if !m.isAllowed(manifest.Name) {
		return result, fmt.Errorf("plugin %q is blocked by allowlist/blocklist", manifest.Name)
	}

	wasmPath, err := ResolvePluginFile(pluginRoot, manifest.Main)
	if err != nil {
		return result, err
	}
	wasmInfo, err := os.Stat(wasmPath)
	if err != nil {
		return result, fmt.Errorf("stat wasm file: %w", err)
	}
	if wasmInfo.IsDir() {
		return result, fmt.Errorf("wasm path is a directory")
	}

	destDir, err := m.pluginPath(manifest.Name)
	if err != nil {
		return result, err
	}
	exists := pathExists(destDir)
	if exists && !opts.Replace {
		return result, fmt.Errorf("plugin %q already exists", manifest.Name)
	}

	if existing := m.GetPlugin(manifest.Name); existing != nil {
		if !opts.Replace {
			return result, fmt.Errorf("plugin %q is already loaded", manifest.Name)
		}
		if err := m.DisablePlugin(manifest.Name); err != nil {
			return result, fmt.Errorf("disable existing plugin: %w", err)
		}
	}

	installTmp := filepath.Join(pluginDir, "."+manifest.Name+".install-"+fmt.Sprint(time.Now().UnixNano()))
	_ = os.RemoveAll(installTmp)
	if err := copyDir(pluginRoot, installTmp); err != nil {
		return result, err
	}
	defer os.RemoveAll(installTmp)

	if exists {
		if err := os.RemoveAll(destDir); err != nil {
			return result, fmt.Errorf("remove existing plugin directory: %w", err)
		}
	}
	if err := os.Rename(installTmp, destDir); err != nil {
		return result, fmt.Errorf("install plugin directory: %w", err)
	}

	if opts.Enable {
		if err := m.EnablePlugin(manifest.Name); err != nil {
			return result, err
		}
	}

	info, _ := m.GetInstalledPluginInfo(manifest.Name)
	result = PluginInstallResult{
		Name:       manifest.Name,
		Replaced:   exists,
		Enabled:    opts.Enable,
		Status:     info.Status,
		Plugin:     info,
		Validation: m.ValidateInstalledPlugin(manifest.Name),
	}
	return result, nil
}

func writeLimitedFile(filename string, r io.Reader, limit int64) error {
	out, err := os.OpenFile(filename, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("create upload archive: %w", err)
	}
	defer out.Close()

	limited := &io.LimitedReader{R: r, N: limit + 1}
	written, err := io.Copy(out, limited)
	if err != nil {
		return fmt.Errorf("write upload archive: %w", err)
	}
	if written > limit {
		return fmt.Errorf("plugin package exceeds %d bytes", limit)
	}
	return nil
}

func extractPluginZip(archivePath, destDir string) error {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return fmt.Errorf("open plugin zip: %w", err)
	}
	defer reader.Close()

	if len(reader.File) == 0 {
		return fmt.Errorf("plugin package is empty")
	}
	if len(reader.File) > maxPluginArchiveFiles {
		return fmt.Errorf("plugin package contains too many files")
	}

	var extractedBytes int64
	for _, item := range reader.File {
		name, err := cleanArchivePath(item.Name)
		if err != nil {
			return err
		}
		info := item.FileInfo()
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("plugin package must not contain symlink: %s", item.Name)
		}

		target := filepath.Join(destDir, filepath.FromSlash(name))
		if err := ensureChildPath(destDir, target); err != nil {
			return err
		}

		if info.IsDir() {
			if err := os.MkdirAll(target, 0755); err != nil {
				return fmt.Errorf("create plugin package directory: %w", err)
			}
			continue
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("plugin package contains unsupported file type: %s", item.Name)
		}

		remainingBytes := maxPluginExtractedBytes - extractedBytes
		if remainingBytes <= 0 {
			return fmt.Errorf("plugin package extracted size exceeds %d bytes", maxPluginExtractedBytes)
		}

		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return fmt.Errorf("create plugin package parent directory: %w", err)
		}
		src, err := item.Open()
		if err != nil {
			return fmt.Errorf("open plugin package file: %w", err)
		}
		written, err := writeExtractedFile(target, &io.LimitedReader{R: src, N: remainingBytes + 1}, info.Mode().Perm())
		closeErr := src.Close()
		if err != nil {
			return err
		}
		extractedBytes += written
		if extractedBytes > maxPluginExtractedBytes {
			return fmt.Errorf("plugin package extracted size exceeds %d bytes", maxPluginExtractedBytes)
		}
		if closeErr != nil {
			return fmt.Errorf("close plugin package file: %w", closeErr)
		}
	}
	return nil
}

func writeExtractedFile(filename string, r io.Reader, mode fs.FileMode) (int64, error) {
	if mode == 0 {
		mode = 0644
	}
	out, err := os.OpenFile(filename, os.O_CREATE|os.O_EXCL|os.O_WRONLY, mode)
	if err != nil {
		return 0, fmt.Errorf("create extracted plugin file: %w", err)
	}
	defer out.Close()
	written, err := io.Copy(out, r)
	if err != nil {
		return written, fmt.Errorf("write extracted plugin file: %w", err)
	}
	return written, nil
}

func cleanArchivePath(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", fmt.Errorf("plugin package contains empty file name")
	}
	if strings.Contains(name, "\\") {
		return "", fmt.Errorf("plugin package path must use slash separators: %s", name)
	}
	cleaned := path.Clean(name)
	if cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, "../") || path.IsAbs(cleaned) {
		return "", fmt.Errorf("plugin package path escapes package root: %s", name)
	}
	return cleaned, nil
}

func findPluginRoot(stagingDir string) (string, error) {
	manifests := make([]string, 0, 1)
	if err := filepath.WalkDir(stagingDir, func(itemPath string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		if entry.Name() == "plugin.yaml" {
			manifests = append(manifests, itemPath)
		}
		return nil
	}); err != nil {
		return "", fmt.Errorf("scan plugin package: %w", err)
	}
	if len(manifests) == 0 {
		return "", fmt.Errorf("plugin package must contain plugin.yaml")
	}
	if len(manifests) > 1 {
		return "", fmt.Errorf("plugin package contains multiple plugin.yaml files")
	}
	return filepath.Dir(manifests[0]), nil
}

func copyDir(srcDir, destDir string) error {
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("create install directory: %w", err)
	}
	return filepath.WalkDir(srcDir, func(itemPath string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(srcDir, itemPath)
		if err != nil {
			return fmt.Errorf("resolve install path: %w", err)
		}
		if rel == "." {
			return nil
		}

		target := filepath.Join(destDir, rel)
		if err := ensureChildPath(destDir, target); err != nil {
			return err
		}
		info, err := entry.Info()
		if err != nil {
			return fmt.Errorf("stat plugin file: %w", err)
		}
		if entry.IsDir() {
			return os.MkdirAll(target, 0755)
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("plugin directory contains unsupported file type: %s", itemPath)
		}

		src, err := os.Open(itemPath)
		if err != nil {
			return fmt.Errorf("open plugin file: %w", err)
		}
		_, err = writeExtractedFile(target, src, info.Mode().Perm())
		closeErr := src.Close()
		if err != nil {
			return err
		}
		if closeErr != nil {
			return fmt.Errorf("close plugin file: %w", closeErr)
		}
		return nil
	})
}

func ensureChildPath(baseDir, target string) error {
	baseAbs, err := filepath.Abs(baseDir)
	if err != nil {
		return fmt.Errorf("resolve base path: %w", err)
	}
	targetAbs, err := filepath.Abs(target)
	if err != nil {
		return fmt.Errorf("resolve target path: %w", err)
	}
	if targetAbs != baseAbs && !strings.HasPrefix(targetAbs, baseAbs+string(os.PathSeparator)) {
		return fmt.Errorf("path escapes base directory: %s", target)
	}
	return nil
}

func pathExists(name string) bool {
	_, err := os.Stat(name)
	return err == nil
}
