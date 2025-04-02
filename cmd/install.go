package cmd

import (
	cfg "codacy/cli-v2/config"
	config_file "codacy/cli-v2/config-file"
	"fmt"
	"log"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var registry string

func init() {
	installCmd.Flags().StringVarP(&registry, "registry", "r", "", "Registry to use for installing tools")
	rootCmd.AddCommand(installCmd)
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Installs the tools specified in the project's config-file.",
	Long:  "Installs all runtimes and tools specified in the project's config-file file.",
	Run: func(cmd *cobra.Command, args []string) {
		cyan := color.New(color.FgCyan)
		bold := color.New(color.Bold)

		// Initialize config
		cfg.Init()

		// Load config file
		if err := config_file.ReadConfigFile(cfg.Config.ProjectConfigFile()); err != nil {
			log.Fatalf("Failed to load config file: %v", err)
		}

		fmt.Println()
		bold.Println("🚀 Starting installation process...")
		fmt.Println()

		cyan.Println("Installing runtimes...")
		installRuntimes(&cfg.Config)

		cyan.Println("\nInstalling tools...")
		installTools(&cfg.Config)

		fmt.Println()
		bold.Println("✅ Installation completed successfully!")
	},
}

func installRuntimes(config *cfg.ConfigType) {
	err := cfg.InstallRuntimes()
	if err != nil {
		log.Fatal(err)
	}
}

func installTools(config *cfg.ConfigType) {
	// Use the new tools-installer instead of manual installation
	err := cfg.InstallTools()
	if err != nil {
		log.Fatal(err)
	}
}
