package cmd

import (
	cfg "codacy/cli-v2/config"
	"codacy/cli-v2/plugins"
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
	err := plugins.InstallRuntimes(config.Runtimes())
	if err != nil {
		log.Fatal(err)
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
		default:
			log.Fatal("Unknown tool:", tool.Name())
		}
	}
}
