package tools

import (
	"codacy/cli-v2/config"
	"os"
	"os/exec"
	"path/filepath"
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
//   - rulesetFile: Path to custom ruleset XML file, if empty uses default ruleset
//
// Returns:
//   - error: nil if analysis succeeds or violations found, error otherwise
func RunPmd(repositoryToAnalyseDirectory string, pmdBinary string, pathsToCheck []string, outputFile string, outputFormat string, rulesetFile string) error {
	cmd := exec.Command(pmdBinary, "pmd")

	// Add config file from tools-configs directory if not specified
	if rulesetFile == "" {
		configFile := filepath.Join(config.Config.ToolsConfigDirectory(), "pmd-ruleset.xml")
		cmd.Args = append(cmd.Args, "-R", configFile)
	} else {
		cmd.Args = append(cmd.Args, "-R", rulesetFile)
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
	}

	cmd.Dir = repositoryToAnalyseDirectory
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 4 {
			// Exit code 4 means violations found – treat as success
			return nil
		}
		return err
	}
	return nil
}
