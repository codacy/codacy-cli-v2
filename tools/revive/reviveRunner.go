package tools

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

// RunRevive executes revive analysis on the specified files or directory
func RunRevive(workDirectory string, binary string, files []string, outputFile string, outputFormat string) error {
	cmdArgs := []string{}

	// Add output format if specified
	if outputFormat != "" {
		cmdArgs = append(cmdArgs, "-formatter", outputFormat)
	}

	// Add output file if specified
	if outputFile != "" {
		cmdArgs = append(cmdArgs, "-o", outputFile)
	}

	// Add files to analyze - if no files specified, analyze current directory
	if len(files) > 0 {
		cmdArgs = append(cmdArgs, files...)
	} else {
		cmdArgs = append(cmdArgs, ".")
	}

	cmd := exec.Command(binary, cmdArgs...)
	cmd.Dir = workDirectory
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			log.Printf("[REVIVE] Error running revive: %v", err)
			return fmt.Errorf("failed to run revive: %w", err)
		}
		log.Printf("[REVIVE] revive exited with non-zero status (findings may be present)")
	} else {
		log.Printf("[REVIVE] revive completed successfully")
	}

	return nil
}
