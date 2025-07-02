package tools

import (
	"codacy/cli-v2/config"
	parenttools "codacy/cli-v2/tools"
	"codacy/cli-v2/utils/logger"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/sirupsen/logrus"
)

// RunRevive executes revive analysis on the specified files or directory
func RunRevive(workDirectory string, binary string, files []string, outputFile string, outputFormat string) error {
	cmdArgs := []string{}

	// Check if a config file exists in the expected location and use it if present
	if configFile, exists := parenttools.ConfigFileExists(config.Config, "revive.toml"); exists {
		log.Printf("[REVIVE] Using config file: %s", configFile)
		cmdArgs = append(cmdArgs, "-config", configFile)
	}

	// Add output format if specified
	if outputFormat != "" {
		cmdArgs = append(cmdArgs, "-formatter", outputFormat)
	}

	// Add files to analyze - if no files specified, analyze current directory
	if len(files) > 0 {
		cmdArgs = append(cmdArgs, files...)
	} else {
		cmdArgs = append(cmdArgs, "./...")
	}

	cmd := exec.Command(binary, cmdArgs...)
	cmd.Dir = workDirectory
	cmd.Stderr = os.Stderr

	// Handle output file redirection
	if outputFile != "" {
		outputWriter, err := os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer outputWriter.Close()
		cmd.Stdout = outputWriter
	} else {
		cmd.Stdout = os.Stdout
	}

	logger.Debug("Running Revive command", logrus.Fields{
		"command": cmd.String(),
	})

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
