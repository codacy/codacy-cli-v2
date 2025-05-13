package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"codacy/cli-v2/config"
	"codacy/cli-v2/utils/logger"
	"codacy/cli-v2/version"

	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var rootCmd = &cobra.Command{
	Use:     "codacy-cli",
	Short:   fmt.Sprintf("Codacy CLI v%s - A command line interface for Codacy", version.GetVersion()),
	Long:    "",
	Example: getExampleText(),
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Initialize logger before any command runs
		logsDir := filepath.Join(config.Config.LocalCodacyDirectory(), "logs")
		if err := logger.Initialize(logsDir); err != nil {
			fmt.Printf("Warning: Failed to initialize file logger: %v\n", err)
		}

		// Create a map to store all flags and their values
		flags := make(map[string]string)
		cmd.Flags().VisitAll(func(flag *pflag.Flag) {
			if flag.Changed {
				// Mask sensitive values
				value := flag.Value.String()
				switch flag.Name {
				case "api-token", "repository-token", "project-token", "codacy-api-token":
					value = "***"
				}
				flags[flag.Name] = value
			}
		})

		// Create a masked version of the full command for logging
		maskedArgs := make([]string, len(os.Args))
		copy(maskedArgs, os.Args)
		for i, arg := range maskedArgs {
			if i > 0 && (arg == "--api-token" || arg == "--repository-token" ||
				arg == "--project-token" || arg == "--codacy-api-token") && i < len(maskedArgs)-1 {
				maskedArgs[i+1] = "***"
			}
		}

		// Log the command being executed with its arguments and flags
		logger.Info("Executing CLI command", logrus.Fields{
			"command":      cmd.Name(),
			"full_command": maskedArgs,
			"args":         args,
		})
	},
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
	fmt.Println("  codacy-cli init --codacy-api-token YOUR_TOKEN")
	fmt.Println()
	fmt.Println("Or run without a token to use local configuration:")
	fmt.Println("  codacy-cli init")
	fmt.Println()
	fmt.Println("For more information about available commands, run:")
	fmt.Println("  codacy-cli --help")
}

func getExampleText() string {
	return color.New(color.FgCyan).Sprint("Initialize a project:") + "\n" +
		color.New(color.FgGreen).Sprint("  codacy-cli init") + "\n\n" +
		color.New(color.FgCyan).Sprint("Install required tools:") + "\n" +
		color.New(color.FgGreen).Sprint("  codacy-cli install") + "\n\n" +
		color.New(color.FgCyan).Sprint("Run analysis with ESLint:") + "\n" +
		color.New(color.FgGreen).Sprint("  codacy-cli analyze --tool eslint") + "\n\n" +
		color.New(color.FgCyan).Sprint("Run analysis and output in SARIF format:") + "\n" +
		color.New(color.FgGreen).Sprint("  codacy-cli analyze --tool eslint --format sarif") + "\n\n" +
		color.New(color.FgCyan).Sprint("Upload results to Codacy:") + "\n" +
		color.New(color.FgGreen).Sprint("  codacy-cli upload -s results.sarif -c <commit-uuid> -t <project-token>")
}

func init() {
	// Add global flags here
	rootCmd.PersistentFlags().String("config", filepath.Join(".codacy", "codacy.yaml"), "config file")

	// Customize help template
	rootCmd.SetUsageTemplate(`
` + color.New(color.FgCyan).Sprint("Usage:") + `
  {{.UseLine}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

` + color.New(color.FgCyan).Sprint("Aliases:") + `
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

` + color.New(color.FgCyan).Sprint("Examples:") + `
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

` + color.New(color.FgCyan).Sprint("Available Commands:") + `{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  ` + "{{$cmd := .Name}}" + color.New(color.FgGreen).Sprintf("{{rpad .Name .NamePadding}}") + ` {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

` + color.New(color.FgCyan).Sprint("Flags:") + `
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

` + color.New(color.FgCyan).Sprint("Global Flags:") + `
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

` + color.New(color.FgCyan).Sprint("Additional help topics:") + `{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}

` + color.New(color.FgCyan).Sprint("Configuration Example") + ` (.codacy/codacy.yaml):
  runtimes:
    - node@22.2.0
  tools:
    - eslint@9.3.0

` + color.New(color.FgCyan).Sprint("For more information and examples, visit:") + `
https://github.com/codacy/codacy-cli-v2
`)
}
