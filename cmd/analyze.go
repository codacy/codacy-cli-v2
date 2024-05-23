package cmd

import (
	"codacy/cli-v2/config"
	"codacy/cli-v2/tools"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
)

var outputFile string

func init() {
	analyzeCmd.Flags().StringVarP(&outputFile, "output", "o", "", "output file for the results")
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

		if len(args) == 0 {
			log.Fatal("You need to specify the tool you want to run for now! ;D")
		}

		eslint := config.Config.Tools()["eslint"]
		if eslint == nil {
			fmt.Println("Eslint is not installed, make sure that it is before running analyze command")
			return
		}
		eslintInstallationDirectory := eslint.Info()["installDir"]
		nodeRuntime := config.Config.Runtimes()["node"]
		nodeBinary := nodeRuntime.Info()["node"]

		log.Printf("Running %s...\n", args[0])
		if outputFile != "" {
			log.Printf("Output will be available at %s\n", outputFile)
		}

		tools.RunEslint(workDirectory, eslintInstallationDirectory, nodeBinary, outputFile)
	},
}
