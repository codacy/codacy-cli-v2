package tools

import (
	"codacy/cli-v2/config"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// isArm64Architecture checks if the current system is running on ARM64
func isArm64Architecture() bool {
	cmd := exec.Command("uname", "-m")
	output, err := cmd.Output()
	if err != nil {
		// If command fails, check GOARCH environment
		if os.Getenv("GOARCH") == "arm64" || runtime.GOARCH == "arm64" {
			return true
		}
		return false
	}

	arch := strings.TrimSpace(string(output))
	return arch == "arm64" || arch == "aarch64"
}

// isRunningInWSL checks if we're running in Windows Subsystem for Linux
func isRunningInWSL() bool {
	// Check for existence of /proc/version which should contain Microsoft or WSL
	data, err := os.ReadFile("/proc/version")
	if err != nil {
		return false
	}

	version := strings.ToLower(string(data))
	return strings.Contains(version, "microsoft") || strings.Contains(version, "wsl")
}

// RunTrivy executes Trivy vulnerability scanner with the specified options
func RunTrivy(repositoryToAnalyseDirectory string, trivyBinary string, pathsToCheck []string, outputFile string, outputFormat string) error {
	fmt.Printf("[DEBUG] Running Trivy from binary: %s\n", trivyBinary)

	// Check the environment
	isArm64 := isArm64Architecture()
	isWSL := isRunningInWSL()

	if isArm64 {
		fmt.Printf("[DEBUG] Running on ARM64 architecture\n")
	}

	if isWSL {
		fmt.Printf("[DEBUG] Running in Windows Subsystem for Linux\n")
	}

	// Verify the trivy binary exists
	_, err := os.Stat(trivyBinary)
	if err != nil {
		fmt.Printf("[ERROR] Trivy binary not found at %s: %v\n", trivyBinary, err)

		// Check if we need to create a fallback output file
		if outputFile != "" && outputFormat == "sarif" {
			return createFallbackSarifOutput(outputFile, fmt.Sprintf("Trivy binary not found: %v", err))
		}

		return fmt.Errorf("trivy binary not found at %s: %w", trivyBinary, err)
	}

	cmd := exec.Command(trivyBinary, "fs")

	// Add config file from tools-configs directory if it exists
	if configFile, exists := ConfigFileExists(config.Config, "trivy.yaml"); exists {
		fmt.Printf("[DEBUG] Using Trivy config file: %s\n", configFile)
		cmd.Args = append(cmd.Args, "--config", configFile)
	}

	// Add format options
	if outputFile != "" {
		cmd.Args = append(cmd.Args, "--output", outputFile)
	}

	if outputFormat == "sarif" {
		cmd.Args = append(cmd.Args, "--format", "sarif")
	}

	// Add specific targets or use current directory
	if len(pathsToCheck) > 0 {
		cmd.Args = append(cmd.Args, pathsToCheck...)
	} else {
		cmd.Args = append(cmd.Args, ".")
	}

	cmd.Dir = repositoryToAnalyseDirectory
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	fmt.Printf("[DEBUG] Running Trivy command: %s %s\n", trivyBinary, strings.Join(cmd.Args, " "))
	fmt.Printf("[DEBUG] Working directory: %s\n", cmd.Dir)

	err = cmd.Run()
	if err != nil {
		fmt.Printf("[ERROR] Failed to run Trivy: %v\n", err)

		// If there's an error and we're on ARM64, provide more specific error handling
		if isArm64 && (strings.Contains(err.Error(), "illegal instruction") || strings.Contains(err.Error(), "SIGILL")) {
			fmt.Printf("[WARN] ARM64 error detected. The Trivy binary appears to be incompatible with this processor.\n")

			if isWSL {
				fmt.Printf("[INFO] Since you're running in WSL, you might need to use a different Trivy build specifically for your ARM64 WSL environment.\n")
			} else {
				fmt.Printf("[INFO] Please reinstall Trivy to generate a compatible binary: 'codacy-cli install'\n")
			}

			// Always create a fallback output file if an output file was requested
			if outputFile != "" && outputFormat == "sarif" {
				return createFallbackSarifOutput(outputFile, "Trivy execution failed (illegal instruction)")
			}

			return fmt.Errorf("trivy execution failed (illegal instruction) - incompatible with ARM64 processor")
		}

		// If any other error occurs and output file is needed, create a fallback
		if outputFile != "" && outputFormat == "sarif" {
			return createFallbackSarifOutput(outputFile, fmt.Sprintf("Trivy execution error: %v", err))
		}

		return fmt.Errorf("failed to run Trivy: %w", err)
	}

	fmt.Printf("[DEBUG] Trivy execution completed successfully\n")
	return nil
}

// createFallbackSarifOutput creates an empty SARIF file as a fallback when Trivy fails
func createFallbackSarifOutput(outputFile string, errorMessage string) error {
	fmt.Printf("[DEBUG] Creating fallback SARIF output file: %s\n", outputFile)
	// Create parent directories if they don't exist
	if err := os.MkdirAll(filepath.Dir(outputFile), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create a minimal valid SARIF file with the error in a tool message
	emptySarif := fmt.Sprintf(`{
  "version": "2.1.0",
  "$schema": "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
  "runs": [
    {
      "tool": {
        "driver": {
          "name": "Trivy",
          "informationUri": "https://github.com/aquasecurity/trivy",
          "rules": []
        }
      },
      "invocations": [
        {
          "executionSuccessful": false,
          "toolExecutionNotifications": [
            {
              "descriptor": {
                "id": "TRV001"
              },
              "level": "error",
              "message": {
                "text": %q
              }
            }
          ]
        }
      ],
      "results": []
    }
  ]
}`, errorMessage)

	err := os.WriteFile(outputFile, []byte(emptySarif), 0644)
	if err != nil {
		return fmt.Errorf("failed to create fallback output file: %w", err)
	}

	fmt.Printf("[INFO] Created fallback SARIF output file due to Trivy execution failure\n")
	return fmt.Errorf("trivy execution failed, created fallback output file with error details")
}
