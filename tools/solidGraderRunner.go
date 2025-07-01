package tools

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func RunSolidGrader(workDirectory string, installationDirectory string, binary string, files []string, outputFile string, outputFormat string) error {
	args := []string{}

	if outputFormat == "sarif" {
		args = append(args, "-f", "sarif")
	}

	if len(files) > 0 {
		args = append(args, files...)
	} else {
		args = append(args, ".")
	}

	/*if configExists != "" {
		log.Println("Config file found, using it")
		args = append(args, "--config", configExists)
	} else {
		log.Println("No config file found, using tool defaults")
	}*/

	cmd := exec.Command(binary, args...)
	cmd.Dir = workDirectory
	cmd.Stderr = os.Stderr
	if outputFile != "" {
		outputWriter, err := os.Create(filepath.Clean(outputFile))
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer outputWriter.Close()
		cmd.Stdout = outputWriter
	} else {
		cmd.Stdout = os.Stdout
	}
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run solid_grader: %w", err)
	}
	return nil
}
