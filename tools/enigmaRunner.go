package tools

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

func RunEnigma(workDirectory string, installationDirectory string, binary string, files []string, outputFile string, outputFormat string) error {

	configFiles := []string{"enigma.yaml", "enigma.yml"}

	if outputFormat == "" {
		outputFormat = "text"
	}

	args := []string{"analyze", "--format", outputFormat}

	if len(files) > 0 {
		args = append(args, append([]string{"--paths"}, files...)...)
	} else {
		args = append(args, "--paths", ".")
	}

	configExists := ""
	for _, configFile := range configFiles {
		if _, err := os.Stat(filepath.Join(workDirectory, configFile)); err == nil {
			configExists = filepath.Join(workDirectory, configFile)
			break
		}
	}

	if configExists != "" {
		log.Println("Config file found, using it")
		args = append(args, "--configuration-file", configExists)
	} else {
		log.Println("No config file found, using tool defaults")

	}

	cmd := exec.Command(binary, args...)
	cmd.Dir = workDirectory
	cmd.Stderr = os.Stderr
	if outputFile != "" {
		// If output file is specified, create it and redirect output
		var outputWriter *os.File
		var err error
		outputWriter, err = os.Create(filepath.Clean(outputFile))
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer outputWriter.Close()
		cmd.Stdout = outputWriter
	} else {
		cmd.Stdout = os.Stdout
	}
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to run Enigma: %w", err)
	}
	return nil
}
