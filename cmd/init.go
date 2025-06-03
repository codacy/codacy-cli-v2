package cmd

import (
	"codacy/cli-v2/cmd/configsetup"
	"codacy/cli-v2/config"
	"codacy/cli-v2/domain"
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
	Long:  "Bootstraps project configuration, creates codacy configuration file",
	Run: func(cmd *cobra.Command, args []string) {
		// Create local codacy directory first
		if err := config.Config.CreateLocalCodacyDir(); err != nil {
			log.Fatalf("Failed to create local codacy directory: %v", err)
		}

		// Create tools-configs directory
		toolsConfigDir := config.Config.ToolsConfigDirectory()
		if err := os.MkdirAll(toolsConfigDir, 0777); err != nil {
			log.Fatalf("Failed to create tools-configs directory: %v", err)
		}

		cliLocalMode := len(initFlags.ApiToken) == 0

		if cliLocalMode {
			fmt.Println()
			fmt.Println("ℹ️  No project token was specified, fetching codacy default configurations")
			noTools := []domain.Tool{}
			err := configsetup.CreateConfigurationFiles(noTools, cliLocalMode)
			if err != nil {
				log.Fatal(err)
			}
			// Create default configuration files
			if err := configsetup.BuildDefaultConfigurationFiles(toolsConfigDir, initFlags); err != nil {
				log.Fatal(err)
			}
			if err := configsetup.CreateLanguagesConfigFileLocal(toolsConfigDir); err != nil {
				log.Fatal(err)
			}
		} else {
			err := configsetup.BuildRepositoryConfigurationFiles(initFlags)
			if err != nil {
				log.Fatal(err)
			}
		}
		configsetup.CreateGitIgnoreFile()
		fmt.Println()
		fmt.Println("✅ Successfully initialized Codacy configuration!")
		fmt.Println()
		fmt.Println("🔧 Next steps:")
		fmt.Println("  1. Run 'codacy-cli install' to install all dependencies")
		fmt.Println("  2. Run 'codacy-cli analyze' to start analyzing your code")
		fmt.Println()
	},
}
