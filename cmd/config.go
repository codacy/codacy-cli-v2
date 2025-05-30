package cmd

import (
	"fmt"
	"log"
	"os"

	"codacy/cli-v2/cmd/configsetup"
	"codacy/cli-v2/config"
	"codacy/cli-v2/domain"
	"codacy/cli-v2/utils"

	"github.com/spf13/cobra"
)

// configUpdateInitFlags holds the flags for the config update command, similar to init command flags.
var configUpdateInitFlags domain.InitFlags

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage Codacy configuration",
}

var configUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update Codacy configuration, or initialize if it doesn't exist",
	Long:  "Updates the Codacy configuration files. If .codacy/codacy.yaml does not exist, it runs the initialization process.",
	Run: func(cmd *cobra.Command, args []string) {
		codacyConfigFile := config.Config.ProjectConfigFile()
		// Check if the main configuration file exists
		if _, err := os.Stat(codacyConfigFile); os.IsNotExist(err) {
			fmt.Println("Configuration file (.codacy/codacy.yaml) not found, running initialization logic...")
			runConfigUpdateLogic(cmd, args, configUpdateInitFlags)
		} else {
			fmt.Println("Updating existing Codacy configuration...")
			runConfigUpdateLogic(cmd, args, configUpdateInitFlags)
		}
	},
}

// runConfigUpdateLogic contains the core logic for updating or initializing the configuration.
// It mirrors the behavior of the original init command but uses shared functions from the configsetup package.
func runConfigUpdateLogic(cmd *cobra.Command, args []string, flags domain.InitFlags) {
	// Create local .codacy directory first
	if err := config.Config.CreateLocalCodacyDir(); err != nil {
		log.Fatalf("Failed to create local codacy directory: %v", err)
	}

	// Create .codacy/tools-configs directory
	toolsConfigDir := config.Config.ToolsConfigDirectory()
	if err := os.MkdirAll(toolsConfigDir, utils.DefaultDirPerms); err != nil {
		log.Fatalf("Failed to create tools-configs directory: %v", err)
	}

	// Determine if running in local mode (no API token)
	cliLocalMode := len(flags.ApiToken) == 0

	if cliLocalMode {
		fmt.Println()
		fmt.Println("ℹ️  No API token was specified. Proceeding with local default configurations.")
		noTools := []domain.Tool{} // Empty slice for tools as we are in local mode without specific toolset from API initially
		if err := configsetup.CreateConfigurationFiles(noTools, cliLocalMode); err != nil {
			log.Fatalf("Failed to create base configuration files: %v", err)
		}
		// Create default configuration files for tools
		if err := configsetup.BuildDefaultConfigurationFiles(toolsConfigDir, flags); err != nil {
			log.Fatalf("Failed to build default tool configuration files: %v", err)
		}
		// Create the languages configuration file for local mode
		if err := configsetup.CreateLanguagesConfigFileLocal(toolsConfigDir); err != nil {
			log.Fatalf("Failed to create local languages configuration file: %v", err)
		}
	} else {
		// API token provided, fetch configuration from Codacy
		fmt.Println("API token specified. Fetching repository-specific configurations from Codacy...")
		if err := configsetup.BuildRepositoryConfigurationFiles(flags); err != nil {
			log.Fatalf("Failed to build repository-specific configuration files: %v", err)
		}
	}

	// Create or update .gitignore file in .codacy directory
	if err := configsetup.CreateGitIgnoreFile(); err != nil {
		log.Printf("Warning: Failed to create or update .codacy/.gitignore: %v", err) // Log as warning, not fatal
	}

	fmt.Println()
	fmt.Println("✅ Successfully initialized/updated Codacy configuration!")
	fmt.Println()
	fmt.Println("🔧 Next steps:")
	fmt.Println("  1. Run 'codacy-cli install' to install all dependencies based on the new/updated configuration.")
	fmt.Println("  2. Run 'codacy-cli analyze' to start analyzing your code.")
	fmt.Println()
}

func init() {
	// Define flags for the config update command. These are the same flags used by the init command.
	configUpdateCmd.Flags().StringVar(&configUpdateInitFlags.ApiToken, "api-token", "", "Optional Codacy API token. If defined, configurations will be fetched from Codacy.")
	configUpdateCmd.Flags().StringVar(&configUpdateInitFlags.Provider, "provider", "", "Provider (e.g., gh, bb, gl) to fetch configurations from Codacy. Required when api-token is provided.")
	configUpdateCmd.Flags().StringVar(&configUpdateInitFlags.Organization, "organization", "", "Remote organization name to fetch configurations from Codacy. Required when api-token is provided.")
	configUpdateCmd.Flags().StringVar(&configUpdateInitFlags.Repository, "repository", "", "Remote repository name to fetch configurations from Codacy. Required when api-token is provided.")

	// Add the update subcommand to the config command
	configCmd.AddCommand(configUpdateCmd)
	// Add the config command to the root command
	rootCmd.AddCommand(configCmd)
}
