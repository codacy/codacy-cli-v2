package config

import (
	"codacy/cli-v2/plugins"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

/*
 * This installs Trivy based on the plugin.yaml configuration
 */
func InstallTrivy(trivyConfig *Runtime, registry string) error {
	log.Println("Installing Trivy")

	// Create Trivy installation directory
	trivyFolder := fmt.Sprintf("%s@%s", trivyConfig.Name(), trivyConfig.Version())
	installDir := filepath.Join(Config.ToolsDirectory(), trivyFolder)

	// Create installation directory
	err := os.MkdirAll(installDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create installation directory: %w", err)
	}

	// Load the plugin configuration
	pluginPath := filepath.Join("plugins", "tools", "trivy", "plugin.yaml")
	pluginConfig, err := plugins.LoadPluginConfig(pluginPath)
	if err != nil {
		return fmt.Errorf("failed to load Trivy plugin configuration: %w", err)
	}

	// Find the download configuration for the current OS
	var downloadConfig *plugins.DownloadConfig
	currentOS := runtime.GOOS
	if currentOS == "darwin" {
		currentOS = "macos"
	}

	for i := range pluginConfig.Downloads {
		for _, os := range pluginConfig.Downloads[i].OS {
			if os == currentOS {
				downloadConfig = &pluginConfig.Downloads[i]
				break
			}
		}
		if downloadConfig != nil {
			break
		}
	}

	if downloadConfig == nil {
		return fmt.Errorf("no download configuration found for OS %s", runtime.GOOS)
	}

	// Get the mapped architecture
	arch := runtime.GOARCH
	if mappedArch, ok := pluginConfig.ArchMapping[arch]; ok {
		arch = mappedArch
	}

	// Get the appropriate extension
	extension := downloadConfig.Extension.Default
	if runtime.GOOS == "windows" && downloadConfig.Extension.Windows != "" {
		extension = downloadConfig.Extension.Windows
	}

	// Template substitution for URL
	version := trivyConfig.Version()
	versionWithPrefix := fmt.Sprintf("v%s", version)

	// Handle template substitution properly
	fileName := strings.ReplaceAll(downloadConfig.FileNameTemplate, "{{.Version}}", version)
	fileName = strings.ReplaceAll(fileName, "{{.OS}}", currentOS)
	fileName = strings.ReplaceAll(fileName, "{{.Arch}}", arch)

	// Generate URL, making sure to handle the version format
	url := strings.ReplaceAll(downloadConfig.URLTemplate, "{{.Version}}", versionWithPrefix)
	// If URL template already has v prefix, we need to use version without prefix
	if strings.Contains(downloadConfig.URLTemplate, "/v{{.Version}}/") {
		url = strings.ReplaceAll(downloadConfig.URLTemplate, "{{.Version}}", version)
	}

	url = strings.ReplaceAll(url, "{{.FileName}}", fileName)
	url = strings.ReplaceAll(url, "{{.Extension}}", extension)
	url = strings.ReplaceAll(url, "{{.OS}}", currentOS)
	url = strings.ReplaceAll(url, "{{.Arch}}", arch)

	log.Printf("Using download URL from plugin configuration: %s\n", url)
	log.Printf("Using filename: %s.%s\n", fileName, extension)

	// Create a temporary directory for the download
	tempDir := filepath.Join(installDir, "temp")
	err = os.MkdirAll(tempDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create temporary directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Download and extract
	downloadPath := filepath.Join(tempDir, fmt.Sprintf("%s.%s", fileName, extension))
	cmd := exec.Command("curl", "-L", "-o", downloadPath, url)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to download Trivy: %w\nOutput: %s", err, string(output))
	}

	log.Printf("Downloaded Trivy to: %s\n", downloadPath)

	// Extract the archive
	extractDir := filepath.Join(tempDir, "extracted")
	err = os.MkdirAll(extractDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create extraction directory: %w", err)
	}

	if extension == "zip" {
		cmd = exec.Command("unzip", "-q", downloadPath, "-d", extractDir)
	} else {
		cmd = exec.Command("tar", "-xzf", downloadPath, "-C", extractDir)
	}
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to extract Trivy: %w\nOutput: %s", err, string(output))
	}

	log.Printf("Extracted Trivy to: %s\n", extractDir)

	// Find the binary from the plugin configuration
	var binaryConfig *plugins.BinaryConfig
	if len(pluginConfig.Binaries) > 0 {
		binaryConfig = &pluginConfig.Binaries[0]
	} else {
		return fmt.Errorf("no binary configuration found in the plugin")
	}

	// Find the binary in the extracted directory
	var binaryPath string
	binaryName := binaryConfig.Name
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}

	err = filepath.Walk(extractDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Base(path) == binaryName {
			binaryPath = path
			return filepath.SkipDir
		}
		return nil
	})

	if binaryPath == "" {
		return fmt.Errorf("could not find %s binary in extracted files", binaryName)
	}

	log.Printf("Found Trivy binary at: %s\n", binaryPath)

	// Copy the binary to the final location
	destPath := filepath.Join(installDir, binaryName)
	source, err := os.Open(binaryPath)
	if err != nil {
		return fmt.Errorf("failed to open source binary: %w", err)
	}
	defer source.Close()

	destination, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create destination binary: %w", err)
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	if err != nil {
		return fmt.Errorf("failed to copy binary: %w", err)
	}

	// Make the binary executable
	err = os.Chmod(destPath, 0755)
	if err != nil {
		return fmt.Errorf("failed to make binary executable: %w", err)
	}

	log.Printf("Successfully installed Trivy %s\n", trivyConfig.Version())
	return nil
}

// isTrivyInstalled checks if Trivy is already installed
func isTrivyInstalled(trivyConfig *Runtime) bool {
	trivyFolder := fmt.Sprintf("%s@%s", trivyConfig.Name(), trivyConfig.Version())
	installDir := filepath.Join(Config.ToolsDirectory(), trivyFolder)
	binaryPath := filepath.Join(installDir, "trivy")

	// Add .exe extension for Windows
	if runtime.GOOS == "windows" {
		binaryPath += ".exe"
	}

	_, err := os.Stat(binaryPath)
	return err == nil
}
