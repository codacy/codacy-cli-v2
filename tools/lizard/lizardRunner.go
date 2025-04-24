package lizard

import (
	"codacy/cli-v2/plugins"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// RunLizard executes Lizard code complexity analyzer with the specified options
func RunLizard(workDirectory string, toolInfo *plugins.ToolInfo, files []string, outputFile string, outputFormat string) error {
	// Get Python binary from venv
	pythonPath := filepath.Join(toolInfo.InstallDir, "venv", "bin", "python3")

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
	cmd := exec.Command(pythonPath, args...)
	cmd.Dir = workDirectory
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the command
	err := cmd.Run()
	if err != nil {
		// Lizard returns non-zero exit code when it finds issues
		// We should not treat this as an error
		if _, ok := err.(*exec.ExitError); !ok {
			return fmt.Errorf("failed to run Lizard: %w", err)
		}
	}

	return nil
}
