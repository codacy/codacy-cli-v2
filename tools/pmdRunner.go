package tools

import (
	"codacy/cli-v2/config"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
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
	if runtime.GOOS == "windows" {
		cmd = exec.Command(pmdBinary) // On Windows, don't add "pmd" subcommand
	} else {
		cmd = exec.Command(pmdBinary, "pmd") // On Unix, use "pmd" subcommand
	}

	// Check for ruleset file in tools-configs directory
	configFile, exists := ConfigFileExists(config, "ruleset.xml")
	if !exists {
		// If no ruleset exists, copy the default ruleset
		defaultRuleset := filepath.Join("tools", "pmd", "default-ruleset.xml")
		targetRuleset := filepath.Join(config.ToolsConfigDirectory(), "ruleset.xml")

		// Ensure tools config directory exists
		if err := os.MkdirAll(config.ToolsConfigDirectory(), 0755); err != nil {
			return fmt.Errorf("failed to create tools config directory: %w", err)
		}

		// Copy default ruleset
		src, err := os.Open(defaultRuleset)
		if err != nil {
			return fmt.Errorf("failed to open default ruleset: %w", err)
		}
		defer src.Close()

		dst, err := os.Create(targetRuleset)
		if err != nil {
			return fmt.Errorf("failed to create target ruleset: %w", err)
		}
		defer dst.Close()

		if _, err := io.Copy(dst, src); err != nil {
			return fmt.Errorf("failed to copy default ruleset: %w", err)
		}

		configFile = targetRuleset
	}

	// Add ruleset file to command arguments
	cmd.Args = append(cmd.Args, "-R", configFile)

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
	}

	cmd.Dir = repositoryToAnalyseDirectory
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	// Get tool info to access environment variables
	toolInfo := config.Tools()["pmd"]
	if toolInfo != nil {
		// Get Java runtime info
		javaRuntime := config.Runtimes()["java"]
		if javaRuntime != nil {
			// Get current environment
			env := os.Environ()

			// Set JAVA_HOME to Java runtime install directory
			env = append(env, fmt.Sprintf("JAVA_HOME=%s", javaRuntime.InstallDir))

			// Add Java bin directory to PATH
			javaBinDir := filepath.Join(javaRuntime.InstallDir, "bin")
			pathEnv := ""
			for _, e := range env {
				if strings.HasPrefix(e, "PATH=") {
					pathEnv = e[5:] // Remove "PATH=" prefix
					break
				}
			}
			if pathEnv == "" {
				pathEnv = os.Getenv("PATH")
			}

			// On Windows, use semicolon as path separator
			pathSeparator := ":"
			if strings.Contains(pathEnv, ";") {
				pathSeparator = ";"
			}

			// Add Java bin directory to the beginning of PATH
			env = append(env, fmt.Sprintf("PATH=%s%s%s", javaBinDir, pathSeparator, pathEnv))

			// Set the environment for the command
			cmd.Env = env
		}
	}

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
