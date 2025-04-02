package tools

import (
	"os"
	"os/exec"
)

// RunPmd executes PMD static code analyzer with the specified options
func RunPmd(repositoryToAnalyseDirectory string, pmdBinary string, pathsToCheck []string, outputFile string, outputFormat string, rulesetFile string) error {
	cmd := exec.Command(pmdBinary, "check")

	// Add ruleset file if provided
	if rulesetFile != "" {
		cmd.Args = append(cmd.Args, "--rulesets", rulesetFile)
	}

	// Add format options
	if outputFormat == "sarif" {
		cmd.Args = append(cmd.Args, "--format", "sarif")
	}

	if outputFile != "" {
		cmd.Args = append(cmd.Args, "--report-file", outputFile)
	}

	// Add directory to scan
	cmd.Args = append(cmd.Args, "--dir", repositoryToAnalyseDirectory)

	cmd.Dir = repositoryToAnalyseDirectory
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 4 {
			// Exit status 4 means violations were found, which is not an error
			return nil
		}
		return err
	}
	return nil
}
