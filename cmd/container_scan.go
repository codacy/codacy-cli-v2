// Package cmd implements the CLI commands for the Codacy CLI tool.
package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"codacy/cli-v2/config"
	config_file "codacy/cli-v2/config-file"
	"codacy/cli-v2/utils/logger"

	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// validImageNamePattern validates Docker image references
// Allows: registry/namespace/image:tag or image@sha256:digest
// Based on Docker image reference specification
var validImageNamePattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._\-/:@]*$`)

// exitFunc is a variable to allow mocking os.Exit in tests
var exitFunc = os.Exit

// CommandRunner interface for running external commands (allows mocking in tests)
type CommandRunner interface {
	Run(name string, args []string) error
}

// ExecCommandRunner runs commands using exec.Command
type ExecCommandRunner struct{}

// Run executes a command and returns its exit error
func (r *ExecCommandRunner) Run(name string, args []string) error {
	// #nosec G204 -- name comes from config (codacy-installed Trivy path),
	// and args are validated by validateImageName() which checks for shell metacharacters.
	// exec.Command passes arguments directly without shell interpretation.
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// commandRunner is the default command runner, can be replaced in tests
var commandRunner CommandRunner = &ExecCommandRunner{}

// ExitCoder interface for errors that have an exit code
type ExitCoder interface {
	ExitCode() int
}

// getExitCode returns the exit code from an error if it implements ExitCoder
func getExitCode(err error) int {
	if exitErr, ok := err.(ExitCoder); ok {
		return exitErr.ExitCode()
	}
	return -1
}

// Flag variables for container-scan command
var (
	severityFlag      string
	pkgTypesFlag      string
	ignoreUnfixedFlag bool
)

func init() {
	containerScanCmd.Flags().StringVar(&severityFlag, "severity", "", "Comma-separated list of severities to scan for (default: HIGH,CRITICAL)")
	containerScanCmd.Flags().StringVar(&pkgTypesFlag, "pkg-types", "", "Comma-separated list of package types to scan (default: os)")
	containerScanCmd.Flags().BoolVar(&ignoreUnfixedFlag, "ignore-unfixed", true, "Ignore unfixed vulnerabilities")
	rootCmd.AddCommand(containerScanCmd)
}

var containerScanCmd = &cobra.Command{
	Use:   "container-scan <IMAGE_NAME> [IMAGE_NAME...]",
	Short: "Scan container images for vulnerabilities using Trivy",
	Long: `Scan one or more container images for vulnerabilities using Trivy.

By default, scans for HIGH and CRITICAL vulnerabilities in OS packages,
ignoring unfixed issues. Use flags to override these defaults.

The --exit-code 1 flag is always applied (not user-configurable) to ensure
the command fails when vulnerabilities are found in any image.`,
	Example: `  # Scan a single image
  codacy-cli container-scan myapp:latest

  # Scan multiple images
  codacy-cli container-scan myapp:latest nginx:alpine redis:7

  # Scan only for CRITICAL vulnerabilities across multiple images
  codacy-cli container-scan --severity CRITICAL myapp:latest nginx:alpine

  # Scan all severities and package types
  codacy-cli container-scan --severity LOW,MEDIUM,HIGH,CRITICAL --pkg-types os,library myapp:latest

  # Include unfixed vulnerabilities
  codacy-cli container-scan --ignore-unfixed=false myapp:latest`,
	Args: cobra.MinimumNArgs(1),
	Run:  runContainerScan,
}

// validateImageName checks if the image name is a valid Docker image reference
// and doesn't contain shell metacharacters that could be used for command injection
func validateImageName(imageName string) error {
	if imageName == "" {
		return fmt.Errorf("image name cannot be empty")
	}

	// Check for maximum length (Docker has a practical limit)
	if len(imageName) > 256 {
		return fmt.Errorf("image name is too long (max 256 characters)")
	}

	// Check for dangerous shell metacharacters first for specific error messages
	dangerousChars := []string{";", "&", "|", "$", "`", "(", ")", "{", "}", "<", ">", "!", "\\", "\n", "\r", "'", "\""}
	for _, char := range dangerousChars {
		if strings.Contains(imageName, char) {
			return fmt.Errorf("invalid image name: contains disallowed character '%s'", char)
		}
	}

	// Validate against allowed pattern for any other invalid characters
	if !validImageNamePattern.MatchString(imageName) {
		return fmt.Errorf("invalid image name format: contains disallowed characters")
	}

	return nil
}

// getTrivyPathResolver is set by tests to mock Trivy path resolution; when nil, real config/install logic is used
var getTrivyPathResolver func() (string, error)

// getTrivyPath returns the path to the Trivy binary (codacy-installed, installed on demand if needed) and an error if not found
func getTrivyPath() (string, error) {
	if getTrivyPathResolver != nil {
		return getTrivyPathResolver()
	}
	if err := config.Config.CreateCodacyDirs(); err != nil {
		return "", fmt.Errorf("failed to create codacy directories: %w", err)
	}
	_ = config_file.ReadConfigFile(config.Config.ProjectConfigFile())
	tool := config.Config.Tools()["trivy"]
	if tool == nil || !config.Config.IsToolInstalled("trivy", tool) {
		if err := config.InstallTool("trivy", tool, ""); err != nil {
			return "", fmt.Errorf("failed to install Trivy: %w", err)
		}
		tool = config.Config.Tools()["trivy"]
	}
	if tool == nil {
		return "", fmt.Errorf("trivy not in config after install")
	}
	trivyPath, ok := tool.Binaries["trivy"]
	if !ok || trivyPath == "" {
		return "", fmt.Errorf("trivy binary path not found")
	}
	logger.Info("Found Trivy", logrus.Fields{"path": trivyPath})
	return trivyPath, nil
}

// handleTrivyNotFound prints error message and exits with code 2
func handleTrivyNotFound(err error) {
	logger.Error("Trivy not found", logrus.Fields{"error": err.Error()})
	color.Red("❌ Error: Trivy could not be installed or found")
	fmt.Println("Run 'codacy-cli init' if you have no project yet, then try container-scan again so Trivy can be installed automatically.")
	fmt.Println("exit-code 2")
	exitFunc(2)
}

func runContainerScan(_ *cobra.Command, args []string) {
	exitCode := executeContainerScan(args)
	exitFunc(exitCode)
}

// executeContainerScan performs the container scan and returns an exit code
// Exit codes: 0 = success, 1 = vulnerabilities found, 2 = error
func executeContainerScan(imageNames []string) int {
	if code := validateAllImages(imageNames); code != 0 {
		return code
	}
	logger.Info("Starting container scan", logrus.Fields{"images": imageNames, "count": len(imageNames)})

	trivyPath, err := getTrivyPath()
	if err != nil {
		handleTrivyNotFound(err)
		return 2
	}

	hasVulnerabilities := scanAllImages(imageNames, trivyPath)
	if hasVulnerabilities == -1 {
		return 2
	}
	return printScanSummary(hasVulnerabilities == 1, imageNames)
}

func validateAllImages(imageNames []string) int {
	for _, imageName := range imageNames {
		if err := validateImageName(imageName); err != nil {
			logger.Error("Invalid image name", logrus.Fields{"image": imageName, "error": err.Error()})
			color.Red("❌ Error: %v", err)
			fmt.Println("exit-code 2")
			return 2
		}
	}
	return 0
}

// scanAllImages scans all images and returns: 0=no vulns, 1=vulns found, -1=error
func scanAllImages(imageNames []string, trivyPath string) int {
	hasVulnerabilities := false
	for i, imageName := range imageNames {
		printScanHeader(imageNames, imageName, i)
		args := buildTrivyArgs(imageName)
		logger.Info("Running Trivy container scan", logrus.Fields{"command": fmt.Sprintf("%s %v", trivyPath, args)})

		if err := commandRunner.Run(trivyPath, args); err != nil {
			if getExitCode(err) == 1 {
				logger.Warn("Vulnerabilities found in image", logrus.Fields{"image": imageName})
				hasVulnerabilities = true
			} else {
				logger.Error("Failed to run Trivy", logrus.Fields{"error": err.Error(), "image": imageName})
				color.Red("❌ Error: Failed to run Trivy for %s: %v", imageName, err)
				fmt.Println("exit-code 2")
				return -1
			}
		} else {
			logger.Info("No vulnerabilities found in image", logrus.Fields{"image": imageName})
		}
	}
	if hasVulnerabilities {
		return 1
	}
	return 0
}

func printScanHeader(imageNames []string, imageName string, index int) {
	if len(imageNames) > 1 {
		fmt.Printf("\n📦 [%d/%d] Scanning image: %s\n", index+1, len(imageNames), imageName)
		fmt.Println(strings.Repeat("-", 50))
	} else {
		fmt.Printf("🔍 Scanning container image: %s\n\n", imageName)
	}
}

func printScanSummary(hasVulnerabilities bool, imageNames []string) int {
	fmt.Println()
	if hasVulnerabilities {
		logger.Warn("Container scan completed with vulnerabilities", logrus.Fields{"images": imageNames})
		color.Red("❌ Scanning failed: vulnerabilities found in one or more container images")
		fmt.Println("exit-code 1")
		return 1
	}
	logger.Info("Container scan completed successfully", logrus.Fields{"images": imageNames})
	color.Green("✅ Success: No vulnerabilities found matching the specified criteria")
	return 0
}

// buildTrivyArgs constructs the Trivy command arguments based on flags
func buildTrivyArgs(imageName string) []string {
	args := []string{
		"image",
		"--scanners", "vuln",
	}

	// Apply --ignore-unfixed if enabled (default: true)
	if ignoreUnfixedFlag {
		args = append(args, "--ignore-unfixed")
	}

	// Apply --severity (use default if not specified)
	severity := severityFlag
	if severity == "" {
		severity = "HIGH,CRITICAL"
	}
	args = append(args, "--severity", severity)

	// Apply --pkg-types (use default if not specified)
	pkgTypes := pkgTypesFlag
	if pkgTypes == "" {
		pkgTypes = "os"
	}
	args = append(args, "--pkg-types", pkgTypes)

	// Always apply --exit-code 1 (not user-configurable)
	args = append(args, "--exit-code", "1")

	// Add the image name as the last argument
	args = append(args, imageName)

	return args
}
