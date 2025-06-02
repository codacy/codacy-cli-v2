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
	// Added import for YAML parsing
)

// configResetInitFlags holds the flags for the config reset command.
var configResetInitFlags domain.InitFlags

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage Codacy configuration",
}

// cliConfigYaml defines the structure for parsing .codacy/cli-config.yaml
type cliConfigYaml struct {
	Mode string `yaml:"mode"`
}

var configResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset Codacy configuration to default or repository-specific settings",
	Long: "Resets the Codacy configuration files and tool-specific configurations. " +
		"This command will overwrite an existing configuration with local default configurations " +
		"if no API token is provided (and current mode is not 'remote'). If an API token is provided, it will fetch and apply " +
		"repository-specific configurations from the Codacy API, effectively resetting to those.",
	Run: func(cmd *cobra.Command, args []string) {
		// Get current CLI mode from config
		currentCliMode, err := config.Config.GetCliMode()
		if err != nil {
			// Log the error for debugging purposes
			log.Printf("Warning: Could not determine CLI mode from cli-config.yaml: %v. Defaulting to 'local' mode.", err)
			// Show a user-facing warning on stdout
			fmt.Println("⚠️  Warning: Could not read or parse .codacy/cli-config.yaml. Defaulting to 'local' CLI mode.")
			fmt.Println("   You might want to run 'codacy-cli init' or 'codacy-cli config reset --api-token ...' to correctly set up your configuration.")
			fmt.Println()
			currentCliMode = "local" // Default to local as per existing logic
		}

		apiTokenFlagProvided := len(configResetInitFlags.ApiToken) > 0

		// If current mode is 'remote', prevent resetting to local without explicit API token for a remote reset.
		if currentCliMode == "remote" && !apiTokenFlagProvided {
			fmt.Println("Error: Your Codacy CLI is currently configured in 'remote' (cloud) mode.")
			fmt.Println("To reset your configuration using remote settings, you must provide the --api-token, --provider, --organization, and --repository flags.")
			fmt.Println("Running 'config reset' without these flags is not permitted while configured for 'remote' mode.")
			fmt.Println("This prevents an accidental switch to a local default configuration.")
			fmt.Println()
			if errHelp := cmd.Help(); errHelp != nil {
				log.Printf("Warning: Failed to display command help: %v\n", errHelp)
			}
			os.Exit(1)
		}

		// Validate flags: if API token is provided, other related flags must also be provided.
		if apiTokenFlagProvided {
			if configResetInitFlags.Provider == "" || configResetInitFlags.Organization == "" || configResetInitFlags.Repository == "" {
				fmt.Println("Error: When using --api-token, you must also provide --provider, --organization, and --repository flags.")
				fmt.Println("Please provide all required flags and try again.")
				fmt.Println()
				if errHelp := cmd.Help(); errHelp != nil {
					log.Fatalf("Failed to display command help: %v", errHelp)
				}
				os.Exit(1)
			}
		}

		codacyConfigFile := config.Config.ProjectConfigFile()
		// Check if the main configuration file exists
		if _, err := os.Stat(codacyConfigFile); os.IsNotExist(err) {
			fmt.Println("Configuration file (.codacy/codacy.yaml) not found, running initialization logic...")
			runConfigResetLogic(cmd, args, configResetInitFlags)
		} else {
			fmt.Println("Resetting existing Codacy configuration...")
			runConfigResetLogic(cmd, args, configResetInitFlags)
		}
	},
}

// runConfigResetLogic contains the core logic for resetting or initializing the configuration.
// It mirrors the behavior of the original init command but uses shared functions from the configsetup package.
func runConfigResetLogic(cmd *cobra.Command, args []string, flags domain.InitFlags) {
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
		fmt.Println("ℹ️  Resetting to local default configurations.")
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
		fmt.Println("API token specified. Fetching and applying repository-specific configurations from Codacy...")
		if err := configsetup.BuildRepositoryConfigurationFiles(flags); err != nil {
			log.Fatalf("Failed to build repository-specific configuration files: %v", err)
		}
	}

	// Create or update .gitignore file in .codacy directory
	if err := configsetup.CreateGitIgnoreFile(); err != nil {
		log.Printf("Warning: Failed to create or update .codacy/.gitignore: %v", err) // Log as warning, not fatal
	}

	fmt.Println()
	fmt.Println("✅ Successfully reset Codacy configuration!")
	fmt.Println()
	fmt.Println("🔧 Next steps:")
	fmt.Println("  1. Run 'codacy-cli install' to install all dependencies based on the new/updated configuration.")
	fmt.Println("  2. Run 'codacy-cli analyze' to start analyzing your code.")
	fmt.Println()
}

func init() {
	// Define flags for the config reset command. These are the same flags used by the init command.
	configResetCmd.Flags().StringVar(&configResetInitFlags.ApiToken, "api-token", "", "Optional Codacy API token. If defined, configurations will be fetched from Codacy.")
	configResetCmd.Flags().StringVar(&configResetInitFlags.Provider, "provider", "", "Provider (e.g., gh, bb, gl) to fetch configurations from Codacy. Required when api-token is provided.")
	configResetCmd.Flags().StringVar(&configResetInitFlags.Organization, "organization", "", "Remote organization name to fetch configurations from Codacy. Required when api-token is provided.")
	configResetCmd.Flags().StringVar(&configResetInitFlags.Repository, "repository", "", "Remote repository name to fetch configurations from Codacy. Required when api-token is provided.")

	// Add the reset subcommand to the config command
	configCmd.AddCommand(configResetCmd)
	// Add the config command to the root command
	rootCmd.AddCommand(configCmd)
}
