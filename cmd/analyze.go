package cmd

import (
	"codacy/cli-v2/config"
	"codacy/cli-v2/tools"
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"os"
	"path"
)

var outputFolder string

func init() {
	analyzeCmd.Flags().StringVarP(&outputFolder, "output", "o", path.Join(".codacy", "out"), "where to output the results")
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

		fmt.Printf("Running the tool %s. Output will be available at %s\n", args[0], outputFolder)
		err = tools.RunEslintToFile(workDirectory, eslintInstallationDirectory, nodeBinary, outputFolder)
		if err != nil {
			log.Fatal(err)
		}
	},
}
