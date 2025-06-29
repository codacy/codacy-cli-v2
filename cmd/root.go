package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"codacy/cli-v2/config"
	"codacy/cli-v2/utils/logger"
	"codacy/cli-v2/version"

	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
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

		// Create a masked version of the full command for logging
		maskedArgs := maskSensitiveArgs(os.Args)

		// Log the command being executed with its arguments and flags
		logger.Info("Executing CLI command", logrus.Fields{
			"command":      cmd.Name(),
			"full_command": maskedArgs,
			"args":         args,
		})

		// Validate codacy.yaml for all commands except init, help, and version
		if !shouldSkipValidation(cmd.Name()) {
			if err := validateCodacyYAML(); err != nil {
				logger.Error("Global validation failed", logrus.Fields{
					"command": cmd.Name(),
					"error":   err.Error(),
				})
				fmt.Println(err)
				os.Exit(1)
			}
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		// Check if codacy.yaml exists
		if _, err := os.Stat(config.Config.ProjectConfigFile()); os.IsNotExist(err) {
			// Show welcome message if codacy.yaml doesn't exist
			showWelcomeMessage()
			return
		}

		// If codacy.yaml exists, show regular help
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

	fmt.Println()
	bold.Println("👋 Welcome to Codacy CLI!")
	fmt.Println()
	fmt.Println("This tool helps you analyze and maintain code quality in your projects.")
	fmt.Println()
	cyan.Println("Get started initializing with your Codacy account:")
	fmt.Println("  codacy-cli init --api-token <token> --provider <gh|gl|bb> --organization <org> --repository <repo>")
	fmt.Println()
	fmt.Println("ℹ️  This will synchronzize tools and paterns from Codacy to your local machine.")
	fmt.Println("   Find your API token at: https://app.codacy.com/account/access-management")
	fmt.Println()
	fmt.Println("Or initialize with default Codacy configuration:")
	fmt.Println("  codacy-cli init")
	fmt.Println()
	fmt.Println("For more information, run:")
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
    - eslint@8.57.0

` + color.New(color.FgCyan).Sprint("For more information and examples, visit:") + `
https://github.com/codacy/codacy-cli-v2
`)
}

// maskSensitiveArgs creates a copy of the arguments with sensitive values masked
func maskSensitiveArgs(args []string) []string {
	maskedArgs := make([]string, len(args))
	copy(maskedArgs, args)

	sensitiveFlags := map[string]bool{
		"--api-token":        true,
		"--repository-token": true,
		"--project-token":    true,
		"--codacy-api-token": true,
	}

	for i, arg := range maskedArgs {
		// Skip the first argument (program name)
		if i == 0 {
			continue
		}

		// Handle --flag=value format
		for flag := range sensitiveFlags {
			if strings.HasPrefix(arg, flag+"=") {
				maskedArgs[i] = flag + "=***"
				break
			}
		}

		// Handle --flag value format
		if sensitiveFlags[arg] && i < len(maskedArgs)-1 {
			maskedArgs[i+1] = "***"
		}
	}
	return maskedArgs
}
