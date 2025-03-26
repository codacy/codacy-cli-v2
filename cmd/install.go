package cmd

import (
	cfg "codacy/cli-v2/config"
	"fmt"
	"log"

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
		case "dart":
			err := cfg.InstallDart(r)
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
			err := cfg.InstallEslint(nodeRuntime, tool, registry)
			if err != nil {
				fmt.Println(err.Error())
				log.Fatal(err)
			}
		case "dartanalyzer":
			// dartanalyzer needs dart runtime
			dartRuntime := config.Runtimes()["dart"]
			err := cfg.InstallDartAnalyzer(dartRuntime, tool, registry)
			if err != nil {
				log.Fatal(err)
			}
		default:
			log.Fatal("Unknown tool:", tool.Name())
		}
	}
}
