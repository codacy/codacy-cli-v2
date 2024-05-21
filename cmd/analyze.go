package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(analyzeCmd)
}

var analyzeCmd = &cobra.Command{
	Use: "analyze",
	Short: "Runs all linters.",
	Long: "Runs all tools for all runtimes.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Hello from 'analyze'")
	},
}
