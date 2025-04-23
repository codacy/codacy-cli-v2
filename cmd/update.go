package cmd

import (
	"codacy/cli-v2/config"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update to the latest version",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		// Read version from yaml
		versionFile := filepath.Join(config.Config.CodacyDirectory(), "version.yaml")
		versionData, err := os.ReadFile(versionFile)
		if err != nil {
			fmt.Printf("Failed to read version.yaml: %v\n", err)
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
