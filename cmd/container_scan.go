// Package cmd implements the CLI commands for the Codacy CLI tool.
package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"codacy/cli-v2/utils/logger"

	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// validImageNamePattern validates Docker image references
// Allows: registry/namespace/image:tag or image@sha256:digest
// Based on Docker image reference specification
var validImageNamePattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._\-/:@]*$`)

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
	Use:   "container-scan [FLAGS] <IMAGE_NAME> [IMAGE_NAME...]",
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

	// Validate against allowed pattern
	if !validImageNamePattern.MatchString(imageName) {
		return fmt.Errorf("invalid image name format: contains disallowed characters")
	}

	// Additional check for dangerous shell metacharacters
	dangerousChars := []string{";", "&", "|", "$", "`", "(", ")", "{", "}", "<", ">", "!", "\\", "\n", "\r", "'", "\""}
	for _, char := range dangerousChars {
		if strings.Contains(imageName, char) {
			return fmt.Errorf("invalid image name: contains disallowed character '%s'", char)
		}
	}

	return nil
}

// getTrivyPath returns the path to the Trivy binary or exits if not found
func getTrivyPath() string {
	trivyPath, err := exec.LookPath("trivy")
	if err != nil {
		logger.Error("Trivy not found", logrus.Fields{"error": err.Error()})
		color.Red("❌ Error: Trivy is not installed or not found in PATH")
		fmt.Println("Please install Trivy to use container scanning.")
		fmt.Println("Visit: https://trivy.dev/latest/getting-started/installation/")
		os.Exit(1)
	}
	logger.Info("Found Trivy", logrus.Fields{"path": trivyPath})
	return trivyPath
}

func runContainerScan(_ *cobra.Command, args []string) {
	imageNames := args

	// Validate all image names first
	for _, imageName := range imageNames {
		if err := validateImageName(imageName); err != nil {
			logger.Error("Invalid image name", logrus.Fields{"image": imageName, "error": err.Error()})
			color.Red("❌ Error: %v", err)
			os.Exit(1)
		}
	}

	logger.Info("Starting container scan", logrus.Fields{"images": imageNames, "count": len(imageNames)})

	trivyPath := getTrivyPath()
	hasVulnerabilities := false

	for i, imageName := range imageNames {
		if len(imageNames) > 1 {
			fmt.Printf("\n📦 [%d/%d] Scanning image: %s\n", i+1, len(imageNames), imageName)
			fmt.Println(strings.Repeat("-", 50))
		} else {
			fmt.Printf("🔍 Scanning container image: %s\n\n", imageName)
		}

		trivyCmd := exec.Command(trivyPath, buildTrivyArgs(imageName)...)
		trivyCmd.Stdout = os.Stdout
		trivyCmd.Stderr = os.Stderr

		logger.Info("Running Trivy container scan", logrus.Fields{"command": trivyCmd.String()})

		if err := trivyCmd.Run(); err != nil {
			if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == 1 {
				logger.Warn("Vulnerabilities found in image", logrus.Fields{"image": imageName})
				hasVulnerabilities = true
			} else {
				logger.Error("Failed to run Trivy", logrus.Fields{"error": err.Error(), "image": imageName})
				color.Red("❌ Error: Failed to run Trivy for %s: %v", imageName, err)
				os.Exit(1)
			}
		} else {
			logger.Info("No vulnerabilities found in image", logrus.Fields{"image": imageName})
		}
	}

	// Print summary for multiple images
	fmt.Println()
	if hasVulnerabilities {
		logger.Warn("Container scan completed with vulnerabilities", logrus.Fields{"images": imageNames})
		color.Red("❌ Scanning failed: vulnerabilities found in one or more container images")
		os.Exit(1)
	}

	logger.Info("Container scan completed successfully", logrus.Fields{"images": imageNames})
	color.Green("✅ Success: No vulnerabilities found matching the specified criteria")
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
