package cmd

import (
	"codacy/cli-v2/cmd/configsetup"
	"codacy/cli-v2/config"
	"codacy/cli-v2/domain"
	"codacy/cli-v2/utils"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
)

var initFlags domain.InitFlags

func init() {
	initCmd.Flags().StringVar(&initFlags.ApiToken, "api-token", "", "optional codacy api token, if defined configurations will be fetched from codacy")
	initCmd.Flags().StringVar(&initFlags.Provider, "provider", "", "provider (gh/bb/gl) to fetch configurations from codacy, required when api-token is provided")
	initCmd.Flags().StringVar(&initFlags.Organization, "organization", "", "remote organization name to fetch configurations from codacy, required when api-token is provided")
	initCmd.Flags().StringVar(&initFlags.Repository, "repository", "", "remote repository name to fetch configurations from codacy, required when api-token is provided")
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Bootstraps project configuration",
	Long:  "Bootstraps project configuration, creates codacy configuration file and necessary tool configurations.",
	Run: func(cmd *cobra.Command, args []string) {
		// Create local .codacy directory first
		if err := config.Config.CreateLocalCodacyDir(); err != nil {
			log.Fatalf("Failed to create local codacy directory: %v", err)
		}

		// Create .codacy/tools-configs directory
		toolsConfigDir := config.Config.ToolsConfigDirectory()
		if err := os.MkdirAll(toolsConfigDir, utils.DefaultDirPerms); err != nil {
			log.Fatalf("Failed to create tools-configs directory: %v", err)
		}

		cliLocalMode := len(initFlags.ApiToken) == 0

		if cliLocalMode {
			fmt.Println()
			fmt.Println("ℹ️  No API token was specified. Proceeding with local default configurations.")
			noTools := []domain.Tool{}
			if err := configsetup.CreateConfigurationFiles(noTools, cliLocalMode); err != nil {
				log.Fatalf("Failed to create base configuration files: %v", err)
			}
			// Create default configuration files for tools
			if err := configsetup.BuildDefaultConfigurationFiles(toolsConfigDir, initFlags); err != nil {
				log.Fatalf("Failed to build default tool configuration files: %v", err)
			}
			// Create the languages configuration file for local mode
			if err := configsetup.CreateLanguagesConfigFileLocal(toolsConfigDir); err != nil {
				log.Fatalf("Failed to create local languages configuration file: %v", err)
			}
		} else {
			// API token provided, fetch configuration from Codacy
			fmt.Println("API token specified. Fetching repository-specific configurations from Codacy...")
			if err := configsetup.BuildRepositoryConfigurationFiles(initFlags); err != nil {
				log.Fatalf("Failed to build repository-specific configuration files: %v", err)
			}
		}

		// Create or update .gitignore file in .codacy directory
		if err := configsetup.CreateGitIgnoreFile(); err != nil {
			log.Printf("Warning: Failed to create or update .codacy/.gitignore: %v", err)
		}

		fmt.Println()
		fmt.Println("✅ Successfully initialized Codacy configuration!")
		fmt.Println()
		fmt.Println("🔧 Next steps:")
		fmt.Println("  1. Run 'codacy-cli install' to install all dependencies.")
		fmt.Println("  2. Run 'codacy-cli analyze' to start analyzing your code.")
		fmt.Println()
	},
}

// KeepToolsWithLatestVersion filters the tools to keep only the latest version of each tool family.
func KeepToolsWithLatestVersion(tools []domain.Tool) (
	toolsWithLatestVersion []domain.Tool,
	uuidToName map[string]string,
	familyToVersions map[string][]string,
) {
	latestTools := map[string]domain.Tool{}
	uuidToName = map[string]string{}
	seen := map[string][]domain.Tool{}
	familyToVersions = map[string][]string{}

	for _, tool := range tools {
		meta, ok := domain.SupportedToolsMetadata[tool.Uuid]
		if !ok {
			continue
		}

		// Track all tools seen per family
		seen[meta.Name] = append(seen[meta.Name], tool)

		// Pick the best version
		current, exists := latestTools[meta.Name]
		if !exists || domain.SupportedToolsMetadata[current.Uuid].Priority > meta.Priority {
			latestTools[meta.Name] = tool
			uuidToName[tool.Uuid] = meta.Name
		}
	}

	// Populate final list and version map for logging
	for family, tools := range seen {
		var versions []string
		for _, t := range tools {
			v := t.Version
			if v == "" {
				v = "(unknown)"
			}
			versions = append(versions, v)
		}
		familyToVersions[family] = versions
	}

	for _, tool := range latestTools {
		toolsWithLatestVersion = append(toolsWithLatestVersion, tool)
	}

	return
}
