package tools

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// RunLicenseSim runs the license-sim tool (duncan.py) in the Python venv with the required environment variable.
func RunLicenseSim(workDirectory string, binary string, files []string, outputFile string, outputFormat string) error {
	// Determine the file to check and extension
	var fileToCheck string
	var ext string
	if len(files) > 0 {
		fileToCheck = files[0]
		ext = filepath.Ext(fileToCheck)
		if ext != "" {
			ext = ext[1:] // remove dot
		} else {
			ext = "py" // default
		}
	} else {
		return fmt.Errorf("No file specified for license-sim")
	}

	// Prepare command: ../../license-sim/venv/bin/python ../../license-sim/duncan.py search -f <file> -e <ext>
	parts := strings.Split(binary, " ")
	cmdArgs := append(parts[1:], "search", "-f", fileToCheck, "-e", ext)
	cmd := exec.Command(parts[0], cmdArgs...)
	cmd.Dir = workDirectory
	cmd.Env = append(os.Environ(), "KMP_DUPLICATE_LIB_OK=TRUE")

	// Output handling
	if outputFile != "" {
		outputWriter, err := os.Create(filepath.Clean(outputFile))
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer outputWriter.Close()
		cmd.Stdout = outputWriter
		cmd.Stderr = outputWriter
	} else {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run license-sim: %w", err)
	}
	return nil
}
