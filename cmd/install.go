package cmd

import (
	cfg "codacy/cli-v2/config"
	"github.com/spf13/cobra"
	"log"
)

func init() {
	rootCmd.AddCommand(installCmd)
}

var installCmd = &cobra.Command{
	Use: "install",
	Short: "Installs the tools specified in the project's config-file.",
	Long: "Installs all runtimes and tools specified in the project's config-file file.",
	Run: func(cmd *cobra.Command, args []string) {
		// install runtimes
		fetchRuntimes(&cfg.Config)
		// install tools
		fetchTools(&cfg.Config)
	},
}

func fetchRuntimes(config *cfg.ConfigType) {
	for _, r := range config.Runtimes() {
		switch r.Name() {
		case "node":
			err := cfg.InstallNode(r)
			if err != nil {
				log.Fatal(err)
			}
		default:
			log.Fatal("Unknown runtime:", r.Name())
		}
	}
}

func fetchTools(config *cfg.ConfigType) {
	for _, tool := range config.Tools() {
		switch tool.Name() {
		case "eslint":
			// eslint needs node runtime
			nodeRuntime := config.Runtimes()["node"]
			err := cfg.InstallEslint(nodeRuntime, tool)
			if err != nil {
				log.Fatal(err)
			}
		default:
			log.Fatal("Unknown tool:", tool.Name())
		}
	}
}


