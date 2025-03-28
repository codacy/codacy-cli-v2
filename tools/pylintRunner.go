package tools

import (
	"bytes"
	"codacy/cli-v2/utils"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

func RunPylint(repositoryToAnalyseDirectory string, pylintInstallationDirectory string, pythonBinary string, pathsToCheck []string, autoFix bool, outputFile string) {

	// Prepare the command to run pylint as a module with JSON output
	var args []string
	args = append(args, "-m", "pylint", "--output-format=json")

	// Create a temporary file for JSON output if an output file is specified
	tempFile := ""
	if outputFile != "" {
		tempFile = filepath.Join(os.TempDir(), "pylint_output.json")
		args = append(args, "--output", tempFile)
	}

	// Add files/directories to check
	if len(pathsToCheck) > 0 {
		args = append(args, pathsToCheck...)
	} else {
		args = append(args, repositoryToAnalyseDirectory)
	}

	cmd := exec.Command(pythonBinary, args...)
	cmd.Dir = repositoryToAnalyseDirectory

	// Set stderr and stdout to be displayed
	cmd.Stderr = os.Stderr

	// For terminal output capture mode
	var stdout bytes.Buffer
	if outputFile == "" {
		// Terminal output mode - capture JSON and print SARIF directly
		cmd.Stdout = &stdout
	} else {
		// File output mode - show output in terminal
		cmd.Stdout = os.Stdout
	}

	// Run pylint
	log.Printf("Running pylint command: %v", cmd.Args)
	// Pylint returns non-zero exit codes when it finds issues, so we're not checking the error
	cmd.Run()

	if outputFile != "" {
		// Read the JSON output from the temporary file
		outputData, err := os.ReadFile(tempFile)
		if err != nil {
			log.Printf("Failed to read pylint output from %s: %v", tempFile, err)
			return
		}

		// Delete temporary file
		defer os.Remove(tempFile)

		// Convert JSON to SARIF using the utility function
		sarifData := utils.ConvertPylintToSarif(outputData)

		// Write SARIF to the output file
		err = os.WriteFile(outputFile, sarifData, 0644)
		if err != nil {
			log.Printf("Failed to write SARIF output to %s: %v", outputFile, err)
		}

		log.Printf("SARIF output saved to: %s\n", outputFile)
	} else {
		// Get the JSON output from the buffer
		jsonOutput := stdout.Bytes()

		// Convert JSON to SARIF
		sarifOutput := utils.ConvertPylintToSarif(jsonOutput)

		// Print the SARIF output to stdout
		os.Stdout.Write(sarifOutput)
	}
}
