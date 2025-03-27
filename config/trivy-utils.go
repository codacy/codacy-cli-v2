package config

import (
	"codacy/cli-v2/utils"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

func genInfoTrivy(r *Runtime) map[string]string {
	trivyFolder := fmt.Sprintf("%s@%s", r.Name(), r.Version())
	installDir := path.Join(Config.ToolsDirectory(), trivyFolder)

	// Path to the binary depends on OS
	binaryName := "trivy"
	if runtime.GOOS == "windows" {
		binaryName = "trivy.exe"
	}

	return map[string]string{
		"installDir": installDir,
		"trivy":      path.Join(installDir, binaryName),
	}
}

// InstallTrivy downloads Trivy binary based on the specified version
func InstallTrivy(trivyRuntime *Runtime) error {
	log.Println("Installing Trivy...")

	// Create installation directory if it doesn't exist
	installDir := trivyRuntime.Info()["installDir"]
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return fmt.Errorf("failed to create installation directory: %w", err)
	}

	// Determine OS and architecture
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	// Map Go architecture to Trivy architecture
	var trivyArch string
	switch goarch {
	case "386":
		trivyArch = "32bit"
	case "amd64":
		trivyArch = "64bit"
	case "arm":
		trivyArch = "ARM"
	case "arm64":
		trivyArch = "ARM64"
	default:
		trivyArch = goarch
	}

	// Determine file extension and platform
	extension := "tar.gz"
	var platform string

	switch goos {
	case "linux":
		platform = "Linux"
	case "darwin":
		platform = "macOS"
	case "windows":
		platform = "Windows"
		extension = "zip"
	default:
		return fmt.Errorf("unsupported OS: %s", goos)
	}

	// Construct the download URL
	version := trivyRuntime.Version()
	fileName := fmt.Sprintf("trivy_%s_%s-%s.%s", version, platform, trivyArch, extension)
	downloadURL := fmt.Sprintf("https://github.com/aquasecurity/trivy/releases/download/v%s/%s", version, fileName)

	log.Printf("Downloading Trivy from: %s", downloadURL)

	// Download the archive to a temporary directory
	tempDir, err := os.MkdirTemp("", "trivy-download")
	if err != nil {
		return fmt.Errorf("failed to create temporary directory: %w", err)
	}
	defer os.RemoveAll(tempDir) // Clean up temp dir when done

	archivePath, err := utils.DownloadFile(downloadURL, tempDir)
	if err != nil {
		return fmt.Errorf("failed to download Trivy: %w", err)
	}

	// Use system commands for extraction to handle directory structure correctly
	if extension == "tar.gz" {
		// For Unix-like systems, use tar command
		cmd := exec.Command("tar", "-xzf", archivePath, "-C", installDir)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to extract Trivy archive: %w", err)
		}
	} else {
		// For Windows, use the ExtractZip function but we need to handle the directory structure
		err = utils.ExtractZip(archivePath, tempDir)
		if err != nil {
			return fmt.Errorf("failed to extract Trivy archive: %w", err)
		}

		// Find the trivy binary in the extracted content and copy it to the install directory
		err = filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() && (info.Name() == "trivy" || info.Name() == "trivy.exe") {
				// Copy the binary to the installation directory
				srcFile, err := os.Open(path)
				if err != nil {
					return err
				}
				defer srcFile.Close()

				destPath := filepath.Join(installDir, info.Name())
				destFile, err := os.Create(destPath)
				if err != nil {
					return err
				}
				defer destFile.Close()

				_, err = io.Copy(destFile, srcFile)
				if err != nil {
					return err
				}

				// Make it executable
				if err := os.Chmod(destPath, 0755); err != nil {
					return err
				}

				return nil
			}

			return nil
		})

		if err != nil {
			return fmt.Errorf("failed to locate and copy Trivy binary: %w", err)
		}
	}

	// Verify trivy binary is available and executable
	trivyBinaryPath := trivyRuntime.Info()["trivy"]
	if _, err := os.Stat(trivyBinaryPath); os.IsNotExist(err) {
		// If not found in root of install dir, try to find it
		err := findAndMoveTrivyBinary(installDir, trivyBinaryPath)
		if err != nil {
			return fmt.Errorf("trivy binary not found after extraction: %w", err)
		}
	}

	// Make the binary executable (for Unix-like systems)
	if goos != "windows" {
		if err := os.Chmod(trivyBinaryPath, 0755); err != nil {
			return fmt.Errorf("failed to make Trivy binary executable: %w", err)
		}
	}

	log.Println("Trivy installed successfully.")
	return nil
}

// findAndMoveTrivyBinary searches for the trivy binary in the directory structure and moves it to the target path
func findAndMoveTrivyBinary(rootDir, targetPath string) error {
	var foundBinaryPath string

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && (strings.HasSuffix(info.Name(), "trivy") || strings.HasSuffix(info.Name(), "trivy.exe")) {
			foundBinaryPath = path
			return filepath.SkipDir // Stop walking once found
		}

		return nil
	})

	if err != nil {
		return err
	}

	if foundBinaryPath == "" {
		return fmt.Errorf("trivy binary not found in extracted archive")
	}

	// If binary is already at target location, we're done
	if foundBinaryPath == targetPath {
		return nil
	}

	// Copy the binary to the target location
	if err := utils.CopyFile(foundBinaryPath, targetPath); err != nil {
		return fmt.Errorf("failed to move trivy binary: %w", err)
	}

	return nil
}
