package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"codacy/cli-v2/utils/logger"

	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// validImageNamePattern validates Docker image references
// Allows: registry/namespace/image:tag or image@sha256:digest
// Based on Docker image reference specification
var validImageNamePattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._\-/:@]*$`)

// dockerfileFromPattern matches FROM instructions in Dockerfiles
var dockerfileFromPattern = regexp.MustCompile(`(?i)^\s*FROM\s+([^\s]+)`)

// Flag variables for container-scan command
var (
	severityFlag      string
	pkgTypesFlag      string
	ignoreUnfixedFlag bool
	dockerfileFlag    string
	composeFileFlag   string
)

func init() {
	containerScanCmd.Flags().StringVar(&severityFlag, "severity", "", "Comma-separated list of severities to scan for (default: HIGH,CRITICAL)")
	containerScanCmd.Flags().StringVar(&pkgTypesFlag, "pkg-types", "", "Comma-separated list of package types to scan (default: os)")
	containerScanCmd.Flags().BoolVar(&ignoreUnfixedFlag, "ignore-unfixed", true, "Ignore unfixed vulnerabilities")
	containerScanCmd.Flags().StringVar(&dockerfileFlag, "dockerfile", "", "Path to Dockerfile for image auto-detection (useful in CI)")
	containerScanCmd.Flags().StringVar(&composeFileFlag, "compose-file", "", "Path to docker-compose.yml for image auto-detection (useful in CI)")
	rootCmd.AddCommand(containerScanCmd)
}

var containerScanCmd = &cobra.Command{
	Use:   "container-scan [FLAGS] [IMAGE_NAME]",
	Short: "Scan container images for vulnerabilities using Trivy",
	Long: `Scan container images for vulnerabilities using Trivy.

By default, scans for HIGH and CRITICAL vulnerabilities in OS packages,
ignoring unfixed issues. Use flags to override these defaults.

If no image is specified, the command will auto-detect images from:
1. Dockerfile (FROM instruction) - scans the base image
2. docker-compose.yml (image fields) - scans all referenced images

Use --dockerfile or --compose-file flags to specify explicit paths (useful in CI/CD).

The --exit-code 1 flag is always applied (not user-configurable) to ensure
the command fails when vulnerabilities are found.`,
	Example: `  # Auto-detect from Dockerfile or docker-compose.yml in current directory
  codacy-cli container-scan

  # Specify Dockerfile path (useful in CI/CD)
  codacy-cli container-scan --dockerfile ./docker/Dockerfile.prod

  # Specify docker-compose file path
  codacy-cli container-scan --compose-file ./deploy/docker-compose.yml

  # Scan a specific image
  codacy-cli container-scan myapp:latest

  # Scan only for CRITICAL vulnerabilities
  codacy-cli container-scan --severity CRITICAL myapp:latest

  # CI/CD example: scan all images before deploy
  codacy-cli container-scan --dockerfile ./Dockerfile --severity HIGH,CRITICAL`,
	Args: cobra.MaximumNArgs(1),
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

// handleTrivyResult processes the Trivy command result and exits appropriately
func handleTrivyResult(err error, imageName string) {
	if err == nil {
		logger.Info("Container scan completed successfully", logrus.Fields{"image": imageName})
		fmt.Println()
		color.Green("✅ Success: No vulnerabilities found matching the specified criteria")
		return
	}

	if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == 1 {
		logger.Warn("Container scan completed with vulnerabilities", logrus.Fields{
			"image": imageName, "exit_code": 1,
		})
		fmt.Println()
		color.Red("❌ Scanning failed: vulnerabilities found in the container image")
		os.Exit(1)
	}

	logger.Error("Failed to run Trivy", logrus.Fields{"error": err.Error()})
	color.Red("❌ Error: Failed to run Trivy: %v", err)
	os.Exit(1)
}

func runContainerScan(_ *cobra.Command, args []string) {
	var images []string

	if len(args) > 0 {
		images = []string{args[0]}
	} else {
		images = detectImages()
		if len(images) == 0 {
			color.Red("❌ Error: No image specified and none found in Dockerfile or docker-compose.yml")
			fmt.Println("Usage: codacy-cli container-scan <IMAGE_NAME>")
			os.Exit(1)
		}
	}

	scanImages(images)
}

// scanImages validates and scans multiple images
func scanImages(images []string) {
	trivyPath := getTrivyPath()
	hasFailures := false

	for _, imageName := range images {
		if err := validateImageName(imageName); err != nil {
			logger.Error("Invalid image name", logrus.Fields{"image": imageName, "error": err.Error()})
			color.Red("❌ Error: %v", err)
			hasFailures = true
			continue
		}

		logger.Info("Starting container scan", logrus.Fields{"image": imageName})
		fmt.Printf("🔍 Scanning container image: %s\n\n", imageName)

		trivyCmd := exec.Command(trivyPath, buildTrivyArgs(imageName)...)
		trivyCmd.Stdout = os.Stdout
		trivyCmd.Stderr = os.Stderr

		logger.Info("Running Trivy container scan", logrus.Fields{"command": trivyCmd.String()})

		if err := trivyCmd.Run(); err != nil {
			hasFailures = true
			handleScanError(err, imageName)
		} else {
			logger.Info("Container scan completed successfully", logrus.Fields{"image": imageName})
			fmt.Println()
			color.Green("✅ Success: No vulnerabilities found in %s", imageName)
		}

		if len(images) > 1 {
			fmt.Println("\n" + strings.Repeat("-", 60) + "\n")
		}
	}

	if hasFailures {
		os.Exit(1)
	}
}

// handleScanError processes scan errors without exiting (for multi-image scans)
func handleScanError(err error, imageName string) {
	if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == 1 {
		logger.Warn("Container scan completed with vulnerabilities", logrus.Fields{
			"image": imageName, "exit_code": 1,
		})
		fmt.Println()
		color.Red("❌ Vulnerabilities found in %s", imageName)
		return
	}
	logger.Error("Failed to run Trivy", logrus.Fields{"error": err.Error()})
	color.Red("❌ Error scanning %s: %v", imageName, err)
}

// printFoundImages displays the found images to the user
func printFoundImages(source string, images []string) {
	color.Cyan("📄 Found images in %s:", source)
	for _, img := range images {
		fmt.Printf("   • %s\n", img)
	}
	fmt.Println()
}

// detectImages auto-detects images from Dockerfile or docker-compose.yml
func detectImages() []string {
	// Priority 0: Check explicit --dockerfile flag
	if dockerfileFlag != "" {
		return detectFromDockerfile(dockerfileFlag, true)
	}

	// Priority 0: Check explicit --compose-file flag
	if composeFileFlag != "" {
		return detectFromCompose(composeFileFlag, true)
	}

	// Priority 1: Auto-detect Dockerfile in current directory
	if images := detectFromDockerfile("Dockerfile", false); images != nil {
		return images
	}

	// Priority 2: Auto-detect docker-compose files
	composeFiles := []string{"docker-compose.yml", "docker-compose.yaml", "compose.yml", "compose.yaml"}
	for _, composeFile := range composeFiles {
		if images := detectFromCompose(composeFile, false); images != nil {
			return images
		}
	}

	return nil
}

// detectFromDockerfile tries to detect images from a Dockerfile
func detectFromDockerfile(path string, showWarning bool) []string {
	images := parseDockerfile(path)
	if len(images) > 0 {
		printFoundImages(path, images)
		return images
	}
	if showWarning {
		color.Yellow("⚠️  No FROM instructions found in %s", path)
	}
	return nil
}

// detectFromCompose tries to detect images from a docker-compose file
func detectFromCompose(path string, showWarning bool) []string {
	images := parseDockerCompose(path)
	if len(images) > 0 {
		printFoundImages(path, images)
		return images
	}
	if showWarning {
		color.Yellow("⚠️  No images found in %s", path)
	}
	return nil
}

// parseDockerfile extracts FROM images from a Dockerfile
func parseDockerfile(path string) []string {
	file, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer file.Close()

	var images []string
	seen := make(map[string]bool)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		matches := dockerfileFromPattern.FindStringSubmatch(line)
		if len(matches) > 1 {
			image := matches[1]
			// Skip build stage aliases (e.g., FROM golang:1.21 AS builder)
			// and scratch images
			if image != "scratch" && !seen[image] {
				seen[image] = true
				images = append(images, image)
			}
		}
	}

	return images
}

// dockerComposeConfig represents the structure of docker-compose.yml
type dockerComposeConfig struct {
	Services map[string]struct {
		Image string `yaml:"image"`
		Build *struct {
			Context    string `yaml:"context"`
			Dockerfile string `yaml:"dockerfile"`
		} `yaml:"build"`
	} `yaml:"services"`
}

// parseDockerCompose extracts images from docker-compose.yml
func parseDockerCompose(path string) []string {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var config dockerComposeConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		logger.Warn("Failed to parse docker-compose file", logrus.Fields{"path": path, "error": err.Error()})
		return nil
	}

	var images []string
	seen := make(map[string]bool)

	for serviceName, service := range config.Services {
		images = processServiceImage(service.Image, images, seen)
		images = processServiceBuild(serviceName, service.Build, images, seen)
	}

	return images
}

// processServiceImage adds a service's image to the list if not already seen
func processServiceImage(image string, images []string, seen map[string]bool) []string {
	if image != "" && !seen[image] {
		seen[image] = true
		images = append(images, image)
	}
	return images
}

// processServiceBuild extracts images from a service's build context Dockerfile
func processServiceBuild(serviceName string, build *struct {
	Context    string `yaml:"context"`
	Dockerfile string `yaml:"dockerfile"`
}, images []string, seen map[string]bool) []string {
	if build == nil {
		return images
	}

	dockerfilePath := resolveDockerfilePath(build.Context, build.Dockerfile)
	dockerfileImages := parseDockerfile(dockerfilePath)

	for _, img := range dockerfileImages {
		if !seen[img] {
			seen[img] = true
			images = append(images, img)
			logger.Info("Found base image from Dockerfile", logrus.Fields{
				"service":    serviceName,
				"dockerfile": dockerfilePath,
				"image":      img,
			})
		}
	}
	return images
}

// resolveDockerfilePath constructs the full path to a Dockerfile
func resolveDockerfilePath(context, dockerfile string) string {
	path := "Dockerfile"
	if dockerfile != "" {
		path = dockerfile
	}
	if context != "" {
		path = filepath.Join(context, path)
	}
	return path
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
