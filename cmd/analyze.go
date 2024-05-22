package cmd

import (
	"codacy/cli-v2/tools"
	"codacy/cli-v2/config"
	"fmt"
	"log"
	"os"
	"github.com/spf13/cobra"
)

var outputFolder string

func init() {
	analyzeCmd.Flags().StringVarP(&outputFolder, "output", "o", ".codacy/out/eslint.sarif", "where to output the results")
	rootCmd.AddCommand(analyzeCmd)
}

var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Runs all linters.",
	Long:  "Runs all tools for all runtimes.",
	Run: func(cmd *cobra.Command, args []string) {

		fmt.Println(outputFolder)

		eslintRunInfo, err := config.GetToolRunInfo("eslint")
		if err != nil {
			log.Fatal(err)
		}

		workDirectory, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}

		eslintInstallationDirectory := eslintRunInfo["eslintInstallationDirectory"]
		nodeBinary := eslintRunInfo["nodeBinary"]

		if len(args) == 0 {
			log.Fatal("You need to specify the tool you want to run for now! ;D")
		}

		msg := fmt.Sprintf("Running the tool %s. Output will be available at %s", args[0], outputFolder)
		fmt.Println(msg)
		err = tools.RunEslintToFile(workDirectory, eslintInstallationDirectory, nodeBinary, outputFolder)
		if err != nil {
			log.Fatal(err)
		}
	},
}
