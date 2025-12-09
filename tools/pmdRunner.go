package tools

import (
	"codacy/cli-v2/config"
	"codacy/cli-v2/utils/logger"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
)

// RunPmd executes PMD static code analyzer with the specified options
//
// Parameters:
//   - repositoryToAnalyseDirectory: The root directory of the repository to analyze
//   - pmdBinary: Path to the PMD executable
//   - pathsToCheck: List of specific paths to analyze, if empty analyzes whole repository
//   - outputFile: Path where analysis results should be written
//   - outputFormat: Format for the output (e.g. "sarif")
//   - config: Configuration object containing tool info
//
// Returns:
//   - error: nil if analysis succeeds or violations found, error otherwise
func RunPmd(repositoryToAnalyseDirectory string, pmdBinary string, pathsToCheck []string, outputFile string, outputFormat string, config config.ConfigType) error {
	var cmd *exec.Cmd

	// Debug: Log the binary path being used
	logger.Debug("PMD binary path", logrus.Fields{
		"pmdBinary": pmdBinary,
	})

	// Check if the binary exists
	if _, err := os.Stat(pmdBinary); err != nil {
		logger.Error("PMD binary not found", logrus.Fields{
			"pmdBinary": pmdBinary,
			"error":     err,
		})
		return fmt.Errorf("PMD binary not found at %s: %w", pmdBinary, err)
	}

	// Get tool info to check version
	toolInfo := config.Tools()["pmd"]
	if toolInfo == nil {
		logger.Warn("PMD tool info not found in configuration")
		return fmt.Errorf("pmd tool info not found in configuration")
	}

	// Check if we're using a newer version (7.0.0+)
	isNewVersion := toolInfo.Version >= "7.0.0"

	if isNewVersion {
		// For newer versions (7.0.0+), use the binary with 'check' command
		cmd = exec.Command(pmdBinary, "check", "--no-fail-on-violation")
	} else {
		// For older versions, use "pmd" subcommand
		if runtime.GOOS == "windows" {
			cmd = exec.Command(pmdBinary) // On Windows, don't add "pmd" subcommand
		} else {
			cmd = exec.Command(pmdBinary, "pmd") // On Unix, use "pmd" subcommand
		}
	}

	// Add config file from tools-configs directory if it exists
	if configFile, exists := ConfigFileExists(config, "ruleset.xml"); exists {
		cmd.Args = append(cmd.Args, "-R", configFile)
	}

	// Add source directories (comma-separated list for PMD)
	if len(pathsToCheck) > 0 {
		dirArg := strings.Join(pathsToCheck, ",")
		cmd.Args = append(cmd.Args, "-d", dirArg)
	} else {
		// Fall back to whole repo if no specific paths given
		cmd.Args = append(cmd.Args, "-d", repositoryToAnalyseDirectory)
	}

	// Format
	if outputFormat != "" {
		cmd.Args = append(cmd.Args, "-f", outputFormat)
	}

	// Output file
	if outputFile != "" {
		cmd.Args = append(cmd.Args, "-r", outputFile)
		// When storing results in a file, all the logs output should go to stderr
		// Note that for formats like SARIF, tools output their results to a temporary file
		cmd.Stdout = os.Stderr
	} else {
		cmd.Stdout = os.Stdout
	}

	cmd.Stderr = os.Stderr
	cmd.Dir = repositoryToAnalyseDirectory

	// Get Java runtime info
	javaRuntime := config.Runtimes()["java"]
	if javaRuntime != nil {
		logger.Debug("Setting up Java environment", logrus.Fields{
			"javaHome": javaRuntime.InstallDir,
		})

		// Get current environment
		env := os.Environ()

		// Set JAVA_HOME to Java runtime install directory
		javaHome := fmt.Sprintf("JAVA_HOME=%s", javaRuntime.InstallDir)
		env = append(env, javaHome)

		// Get Java binary path from runtime configuration
		javaBinary := javaRuntime.Binaries["java"]
		javaBinDir := filepath.Dir(javaBinary)

		// Check if Java binary exists
		if _, err := os.Stat(javaBinary); err != nil {
			logger.Error("Java binary not found", logrus.Fields{
				"expectedPath": javaBinary,
				"error":        err,
			})

			// Not throwing - Fallback to the default Java runtime
			// This fallback going to be removed in the future https://codacy.atlassian.net/browse/PLUTO-1421
			fmt.Printf("⚠️ Warning: Java binary not found at %s: %v\n", javaBinary, err)
			fmt.Println("⚠️ Trying to continue with the default Java runtime")
			logger.Warn("Java binary not found. Continuing with the default Java runtime", logrus.Fields{
				"expectedPath": javaBinary,
				"error":        err,
			})
		} else {
			// When java binary is found, we need to add it to the PATH

			// Get current PATH
			pathEnv := os.Getenv("PATH")

			// On Windows, use semicolon as path separator
			pathSeparator := ":"
			if runtime.GOOS == "windows" {
				pathSeparator = ";"
			}

			// Add Java bin directory to the beginning of PATH
			newPath := fmt.Sprintf("PATH=%s%s%s", javaBinDir, pathSeparator, pathEnv)
			env = append(env, newPath)

			logger.Debug("Updated environment variables", logrus.Fields{
				"javaHome":   javaHome,
				"path":       newPath,
				"binDir":     javaBinDir,
				"javaBinary": javaBinary,
			})

			// Set the environment for the command
			cmd.Env = env
		}

	} else {
		logger.Warn("Java runtime not found in configuration")
		return fmt.Errorf("java runtime not found in configuration")
	}

	logger.Debug("Running PMD command", logrus.Fields{
		"command": cmd.String(),
	})

	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 4 {
			// Exit code 4 means violations found – treat as success
			return nil
		}
		return fmt.Errorf("failed to run PMD: %w", err)
	}
	return nil
}
