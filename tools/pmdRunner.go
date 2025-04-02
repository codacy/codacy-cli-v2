package tools

import (
	"os"
	"os/exec"
	"strings"
)

// RunPmd executes PMD static code analyzer with the specified options
func RunPmd(repositoryToAnalyseDirectory string, pmdBinary string, pathsToCheck []string, outputFile string, outputFormat string, rulesetFile string) error {
	cmdArgs := []string{"pmd"}

	// Add source directories (comma-separated list for PMD)
	if len(pathsToCheck) > 0 {
		dirArg := strings.Join(pathsToCheck, ",")
		cmdArgs = append(cmdArgs, "-d", dirArg)
	} else {
		// Fall back to whole repo if no specific paths given
		cmdArgs = append(cmdArgs, "-d", repositoryToAnalyseDirectory)
	}

	// Add ruleset
	if rulesetFile != "" {
		cmdArgs = append(cmdArgs, "-R", rulesetFile)
	}

	// Format
	if outputFormat != "" {
		cmdArgs = append(cmdArgs, "-f", outputFormat)
	}

	// Output file
	if outputFile != "" {
		cmdArgs = append(cmdArgs, "-r", outputFile)
	}

	cmd := exec.Command(pmdBinary, cmdArgs...)

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
