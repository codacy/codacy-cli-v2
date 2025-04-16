package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update to the latest version",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		// Get the latest version
		fmt.Println("Checking for latest version...")
		latestVersion, err := getLatestVersion()
		if err != nil {
			fmt.Printf("Failed to get latest version: %v\n", err)
			os.Exit(1)
		}

		// Get the cache directory based on OS
		var cacheDir string
		switch runtime.GOOS {
		case "linux":
			cacheDir = filepath.Join(os.Getenv("HOME"), ".cache", "codacy", "codacy-cli-v2")
		case "darwin":
			cacheDir = filepath.Join(os.Getenv("HOME"), "Library", "Caches", "Codacy", "codacy-cli-v2")
		default:
			cacheDir = ".codacy-cli-v2"
		}

		// Check if version is already installed
		versionDir := filepath.Join(cacheDir, latestVersion)
		cachedBinary := filepath.Join(versionDir, "codacy-cli-v2")
		if _, err := os.Stat(cachedBinary); err == nil {
			fmt.Printf("Version %s is already installed locally\n", latestVersion)
		} else {
			// Create version-specific directory
			if err := os.MkdirAll(versionDir, 0755); err != nil {
				fmt.Printf("Failed to create cache directory: %v\n", err)
				os.Exit(1)
			}

			// Download and extract the latest version to cache
			fmt.Printf("Downloading version %s...\n", latestVersion)
			if err := downloadAndExtract(latestVersion, versionDir); err != nil {
				fmt.Printf("Failed to download and extract version %s: %v\n", latestVersion, err)
				os.Exit(1)
			}
		}

		// Get the current executable path
		executable, err := os.Executable()
		if err != nil {
			fmt.Printf("Failed to get executable path: %v\n", err)
			os.Exit(1)
		}

		// Copy the binary from cache to the executable location
		if err := copyFile(cachedBinary, executable); err != nil {
			fmt.Printf("Failed to copy binary: %v\n", err)
			os.Exit(1)
		}

		// Update version.json in .codacy directory
		versionFile := filepath.Join(".codacy", "version.json")
		if err := os.MkdirAll(filepath.Dir(versionFile), 0755); err != nil {
			fmt.Printf("Failed to create .codacy directory: %v\n", err)
			os.Exit(1)
		}

		versionInfo := struct {
			Version string `json:"version"`
		}{
			Version: latestVersion,
		}
		versionJSON, err := json.Marshal(versionInfo)
		if err != nil {
			fmt.Printf("Failed to create version.json: %v\n", err)
			os.Exit(1)
		}
		if err := os.WriteFile(versionFile, versionJSON, 0644); err != nil {
			fmt.Printf("Failed to write version.json: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully updated to version %s\n", latestVersion)
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}

func getLatestVersion() (string, error) {
	resp, err := http.Get("https://api.github.com/repos/codacy/codacy-cli-v2/releases/latest")
	if err != nil {
		return "", fmt.Errorf("failed to fetch latest version: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch latest version: status code %d", resp.StatusCode)
	}

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	return release.TagName, nil
}

func downloadAndExtract(version, targetDir string) error {
	// Determine the appropriate asset name based on OS and architecture
	var assetName string
	switch runtime.GOOS {
	case "darwin":
		assetName = "codacy-cli-v2_darwin_amd64.tar.gz"
	case "linux":
		assetName = "codacy-cli-v2_linux_amd64.tar.gz"
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	// Download the release asset
	url := fmt.Sprintf("https://github.com/codacy/codacy-cli-v2/releases/download/%s/%s", version, assetName)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download release: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download release: status code %d", resp.StatusCode)
	}

	// Create a temporary file for the downloaded asset
	tmpFile, err := os.CreateTemp("", "codacy-cli-v2-*.tar.gz")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Copy the downloaded content to the temporary file
	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		return fmt.Errorf("failed to save downloaded file: %v", err)
	}
	tmpFile.Close()

	// Extract the tar.gz file
	cmd := exec.Command("tar", "xzf", tmpFile.Name(), "-C", targetDir)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to extract archive: %v", err)
	}

	return nil
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %v", err)
	}
	defer sourceFile.Close()

	// Create the destination file with the same permissions as the source
	sourceInfo, err := sourceFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to get source file info: %v", err)
	}

	destFile, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, sourceInfo.Mode())
	if err != nil {
		return fmt.Errorf("failed to create destination file: %v", err)
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return fmt.Errorf("failed to copy file: %v", err)
	}

	return nil
}
