package cmd

import (
	"codacy/cli-v2/config"
	"codacy/cli-v2/tools"
	"codacy/cli-v2/utils"
	"log"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var outputFile string
var toolToAnalyze string
var autoFix bool
var doNewPr bool

func init() {
	analyzeCmd.Flags().StringVarP(&outputFile, "output", "o", "", "output file for the results")
	analyzeCmd.Flags().StringVarP(&toolToAnalyze, "tool", "t", "", "Which tool to run analysis with")
	analyzeCmd.Flags().BoolVarP(&autoFix, "fix", "f", false, "Apply auto fix to your issues when available")
	analyzeCmd.Flags().BoolVar(&doNewPr, "new-pr", false, "Create a new PR on GitHub containing the fixed issues")
	rootCmd.AddCommand(analyzeCmd)
}

var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Runs all linters.",
	Long:  "Runs all tools for all runtimes.",
	Run: func(cmd *cobra.Command, args []string) {
		workDirectory, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}

		// TODO add more tools here
		switch toolToAnalyze {
		case "eslint":
			// nothing
		case "":
			log.Fatal("You need to specify a tool to run analysis with, e.g., '--tool eslint'", toolToAnalyze)
		default:
			log.Fatal("Trying to run unsupported tool: ", toolToAnalyze)
		}

		// can't create a new PR if there will be no changes/fixed issues
		if doNewPr && !autoFix {
			log.Fatal("Can't create a new PR with fixes without fixing issues. Use the '--fix' option.")
		} else if doNewPr {
			failIfThereArePendingChanges()
		}

		eslint := config.Config.Tools()["eslint"]
		eslintInstallationDirectory := eslint.Info()["installDir"]
		nodeRuntime := config.Config.Runtimes()["node"]
		nodeBinary := nodeRuntime.Info()["node"]

		log.Printf("Running %s...\n", toolToAnalyze)
		if outputFile != "" {
			log.Println("Output will be available at", outputFile)
		}

		tools.RunEslint(workDirectory, eslintInstallationDirectory, nodeBinary, args, autoFix, outputFile)

		if doNewPr {
			utils.CreatePr(false)
		}
	},
}

func failIfThereArePendingChanges() {
	cmd := exec.Command("git", "status", "--porcelain")
	out, _ := cmd.Output()

	if string(out) != "" {
		log.Fatal("There are pending changes, cannot proceed. Commit your pending changes.")
	}
}
