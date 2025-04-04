package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "codacy-cli",
	Short: "Codacy CLI - A command line interface for Codacy",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		// Check if .codacy directory exists
		if _, err := os.Stat(".codacy"); os.IsNotExist(err) {
			// Show welcome message if .codacy doesn't exist
			showWelcomeMessage()
			return
		}

		// If .codacy exists, show regular help
		cmd.Help()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func showWelcomeMessage() {
	bold := color.New(color.Bold)
	cyan := color.New(color.FgCyan)
	yellow := color.New(color.FgYellow)

	fmt.Println()
	bold.Println("👋 Welcome to Codacy CLI!")
	fmt.Println()
	fmt.Println("This tool helps you analyze and maintain code quality in your projects.")
	fmt.Println()
	yellow.Println("To get started, you'll need a Codacy API token.")
	fmt.Println("You can find your Project API token in Codacy under:")
	fmt.Println("Project > Settings > Integrations > Repository API tokens")
	fmt.Println()
	cyan.Println("Initialize your project with:")
	fmt.Println("  codacy-cli init --repository-token YOUR_TOKEN")
	fmt.Println()
	fmt.Println("Or run without a token to use local configuration:")
	fmt.Println("  codacy-cli init")
}
