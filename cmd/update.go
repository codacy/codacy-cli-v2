package cmd

import (
	"fmt"
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
		// Get the cache directory based on OS or environment variable
		var cacheDir string
		envDir := os.Getenv("CODACY_CLI_V2_TMP_FOLDER")
		if envDir != "" {
			cacheDir = envDir
		} else {
			homePath, err := os.UserHomeDir()
			if err != nil {
				fmt.Printf("Failed to get home directory: %v\n", err)
				os.Exit(1)
			}
			switch runtime.GOOS {
			case "linux":
				cacheDir = filepath.Join(homePath, ".cache", "codacy", "codacy-cli-v2")
			case "darwin":
				cacheDir = filepath.Join(homePath, "Library", "Caches", "Codacy", "codacy-cli-v2")
			default:
				cacheDir = ".codacy-cli-v2"
			}
		}

		// Read version from yaml
		versionFile := filepath.Join(cacheDir, "version.yaml")
		versionData, err := os.ReadFile(versionFile)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Println("Could not read version file. Make sure you have the latest version of the script at:")
				fmt.Println("https://github.com/codacy/codacy-cli-v2/blob/main/README.md#download")
			} else {
				fmt.Printf("Failed to read version.yaml: %v\n", err)
			}
			os.Exit(1)
		}

		var versionInfo struct {
			Version string `yaml:"version"`
		}
		if err := yaml.Unmarshal(versionData, &versionInfo); err != nil {
			fmt.Printf("Failed to parse version.yaml: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✓ Updated to version %s\n", versionInfo.Version)
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
