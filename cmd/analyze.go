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

		fmt.Printf("Running %s...\n", args[0])
		if outputFile != "" {
			fmt.Printf("Output will be available at %s\n", outputFile)
		}

		if outputFile != "" {
			err = tools.RunEslintToFile(workDirectory, eslintInstallationDirectory, nodeBinary, outputFile)
		} else {
			out, err2 := tools.RunEslintToString(workDirectory, eslintInstallationDirectory, nodeBinary)
			if err2 != nil {
				log.Fatal(err2)
			}

			fmt.Println(out)
		}

		if err != nil {
			log.Fatal(err)
		}
	},
}
