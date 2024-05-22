package cmd

import (
	cfg "codacy/cli-v2/config"
	toolutils "codacy/cli-v2/tool-utils"
	"codacy/cli-v2/utils"
	"github.com/spf13/cobra"
	"log"
	"os"
	"path/filepath"
)

func init() {
	rootCmd.AddCommand(installCmd)
}

var installCmd = &cobra.Command{
	Use: "install",
	Short: "Installs the tools specified in the project's config.",
	Long: "Installs all runtimes and tools specified in the project's config file.",
	Run: func(cmd *cobra.Command, args []string) {
		// install runtimes
		fetchRuntimes(cfg.Config.Runtimes(), cfg.Config.RuntimesDirectory())
		// install tools
		for _, r := range cfg.Config.Runtimes() {
			fetchTools(r, cfg.Config.RuntimesDirectory(), cfg.Config.ToolsDirectory())
		}
	},
}

func fetchRuntimes(runtimes map[string]*cfg.Runtime, runtimesDirectory string) {
	for _, r := range runtimes {
		switch r.Name() {
		case "node":
			// TODO should delete downloaded archive
			// TODO check for deflated archive
			log.Println("Fetching node...")
			downloadNodeURL := cfg.GetNodeDownloadURL(r)
			nodeTar, err := utils.DownloadFile(downloadNodeURL, runtimesDirectory)
			if err != nil {
				log.Fatal(err)
			}

			// deflate node archive
			t, err := os.Open(nodeTar)
			defer t.Close()
			if err != nil {
				log.Fatal(err)
			}
			err = utils.ExtractTarGz(t, runtimesDirectory)
			if err != nil {
				log.Fatal(err)
			}
		default:
			log.Fatal("Unknown runtime:", r.Name())
		}
	}
}

func fetchTools(runtime *cfg.Runtime, runtimesDirectory string, toolsDirectory string) {
	for _, tool := range runtime.Tools() {
		switch tool.Name() {
		case "eslint":
			npmPath := filepath.Join(runtimesDirectory, cfg.GetNodeFileName(runtime),
				"bin", "npm")
			toolutils.InstallESLint(npmPath, "eslint@" + tool.Version(), toolsDirectory)
		default:
			log.Fatal("Unknown tool:", tool.Name())
		}
	}
}


