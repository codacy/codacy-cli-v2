package config

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

/*
 * This installs Trivy using the official install script
 */
func InstallTrivy(trivyConfig *Runtime, registry string) error {
	log.Println("Installing Trivy")

	// Create Trivy installation directory
	trivyFolder := fmt.Sprintf("%s@%s", trivyConfig.Name(), trivyConfig.Version())
	installDir := filepath.Join(Config.ToolsDirectory(), trivyFolder)

	// Check if already installed
	if isTrivyInstalled(trivyConfig) {
		fmt.Printf("Trivy %s is already installed\n", trivyConfig.Version())
		return nil
	}

	// Create installation directory
	err := os.MkdirAll(installDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create installation directory: %w", err)
	}

	// Use the official install script to download and install Trivy
	version := fmt.Sprintf("v%s", trivyConfig.Version())
	installScriptURL := "https://raw.githubusercontent.com/aquasecurity/trivy/main/contrib/install.sh"

	log.Printf("Installing Trivy %s using the official install script\n", version)

	// Create a temporary directory for the installation
	tempDir := filepath.Join(installDir, "temp")
	err = os.MkdirAll(tempDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create temporary directory: %w", err)
	}

	// Download and run the install script
	cmd := exec.Command("sh", "-c", fmt.Sprintf("curl -sfL %s | sh -s -- -b %s %s",
		installScriptURL, tempDir, version))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to run Trivy install script: %w\nOutput: %s", err, string(output))
	}

	log.Printf("Install script output: %s\n", string(output))

	// Copy the Trivy binary to the final location
	sourcePath := filepath.Join(tempDir, "trivy")
	if runtime.GOOS == "windows" {
		sourcePath += ".exe"
	}

	binaryPath := filepath.Join(installDir, "trivy")
	if runtime.GOOS == "windows" {
		binaryPath += ".exe"
	}

	// Check if the source binary exists
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return fmt.Errorf("trivy binary not found at %s after installation", sourcePath)
	}

	// Copy the binary to the final location
	source, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to open source binary: %w", err)
	}
	defer source.Close()

	destination, err := os.Create(binaryPath)
	if err != nil {
		return fmt.Errorf("failed to create destination binary: %w", err)
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	if err != nil {
		return fmt.Errorf("failed to copy binary: %w", err)
	}

	// Make the copied binary executable
	err = os.Chmod(binaryPath, 0755)
	if err != nil {
		return fmt.Errorf("failed to make binary executable: %w", err)
	}

	// Clean up the temporary directory
	os.RemoveAll(tempDir)

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
