package cmd

import (
	"codacy/cli-v2/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
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

		// Get the cache directory based on OS or environment variable
		var cacheDir string
		if envDir := os.Getenv("CODACY_CLI_V2_TMP_FOLDER"); envDir != "" {
			cacheDir = envDir
		} else {
			switch runtime.GOOS {
			case "linux":
				cacheDir = filepath.Join(os.Getenv("HOME"), ".cache", "codacy", "codacy-cli-v2")
			case "darwin":
				cacheDir = filepath.Join(os.Getenv("HOME"), "Library", "Caches", "Codacy", "codacy-cli-v2")
			default:
				cacheDir = ".codacy-cli-v2"
			}
		}

		// Update version.yaml
		versionFile := filepath.Join(cacheDir, "version.yaml")
		if err := os.MkdirAll(cacheDir, utils.DefaultDirPerms); err != nil {
			fmt.Printf("Failed to create cache directory: %v\n", err)
			os.Exit(1)
		}

		versionInfo := struct {
			Version string `yaml:"version"`
		}{
			Version: latestVersion,
		}
		versionYAML, err := yaml.Marshal(versionInfo)
		if err != nil {
			fmt.Printf("Failed to create version.yaml: %v\n", err)
			os.Exit(1)
		}
		if err := os.WriteFile(versionFile, versionYAML, utils.DefaultFilePerms); err != nil {
			fmt.Printf("Failed to write version.yaml: %v\n", err)
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
