package lizard

import (
	"fmt"
	"os"
	"os/exec"
)

// RunLizard executes Lizard code complexity analyzer with the specified options
func RunLizard(workDirectory string, binary string, files []string, outputFile string, outputFormat string) error {

	// Construct base command with -m lizard to run lizard module
	args := []string{"-m", "lizard"}

	// Add output file if specified
	if outputFile != "" {
		args = append(args, "-o", outputFile)
	}

	// Add files to analyze - if no files specified, analyze current directory
	if len(files) > 0 {
		args = append(args, files...)
	} else {
		args = append(args, ".")
	}

	// Create and run command
	cmd := exec.Command(binary, args...)
	cmd.Dir = workDirectory
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the command
	err := cmd.Run()
	if err != nil {
		// Lizard returns 1 when it finds issues, which is not a failure
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return nil
		}
		return fmt.Errorf("failed to run Lizard: %w", err)
	}
	return nil
}
