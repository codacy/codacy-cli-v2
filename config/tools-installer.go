package config

import (
	"bytes"
	"codacy/cli-v2/plugins"
	"codacy/cli-v2/utils"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

// InstallTools installs all tools defined in the configuration
func InstallTools() error {
	for name, toolInfo := range Config.Tools() {
		err := InstallTool(name, toolInfo)
		if err != nil {
			return fmt.Errorf("failed to install tool %s: %w", name, err)
		}
	}
	return nil
}

// InstallTool installs a specific tool
func InstallTool(name string, toolInfo *plugins.ToolInfo) error {
	// Check if the tool is already installed
	if isToolInstalled(toolInfo) {
		fmt.Printf("Tool %s v%s is already installed\n", name, toolInfo.Version)
		return nil
	}

	// Make sure the installation directory exists
	err := os.MkdirAll(toolInfo.InstallDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create installation directory: %w", err)
	}

	// Check if this is a download-based tool (like trivy) or a runtime-based tool (like eslint)
	if toolInfo.DownloadURL != "" {
		// This is a download-based tool
		return installDownloadBasedTool(toolInfo)
	}

	// This is a runtime-based tool, proceed with regular installation

	// Get the runtime for this tool
	runtimeInfo, ok := Config.Runtimes()[toolInfo.Runtime]
	if !ok {
		return fmt.Errorf("required runtime %s not found for tool %s", toolInfo.Runtime, name)
	}

	// Prepare template data
	templateData := map[string]string{
		"InstallDir":  toolInfo.InstallDir,
		"PackageName": toolInfo.Name,
		"Version":     toolInfo.Version,
		"Registry":    "", // TODO: Get registry from config
	}

	// Get package manager binary based on the tool configuration
	packageManagerName := toolInfo.PackageManager
	packageManagerBinary, ok := runtimeInfo.Binaries[packageManagerName]
	if !ok {
		return fmt.Errorf("package manager binary %s not found in runtime %s", packageManagerName, toolInfo.Runtime)
	}

	// Set registry if provided
	if toolInfo.RegistryCommand != "" {
		regCmd, err := executeToolTemplate(toolInfo.RegistryCommand, templateData)
		if err != nil {
			return fmt.Errorf("failed to prepare registry command: %w", err)
		}

		if regCmd != "" {
			registryCmd := exec.Command(packageManagerBinary, strings.Split(regCmd, " ")...)
			if output, err := registryCmd.CombinedOutput(); err != nil {
				return fmt.Errorf("failed to set registry: %s: %w", string(output), err)
			}
		}
	}

	// Execute installation command
	installCmd, err := executeToolTemplate(toolInfo.InstallCommand, templateData)
	if err != nil {
		return fmt.Errorf("failed to prepare install command: %w", err)
	}

	// Execute the installation command using the package manager
	cmd := exec.Command(packageManagerBinary, strings.Split(installCmd, " ")...)

	log.Printf("Installing %s v%s...\n", toolInfo.Name, toolInfo.Version)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to install tool: %s: %w", string(output), err)
	}

	log.Printf("Successfully installed %s v%s\n", toolInfo.Name, toolInfo.Version)
	return nil
}

// installDownloadBasedTool installs a tool by downloading and extracting it
func installDownloadBasedTool(toolInfo *plugins.ToolInfo) error {
	// Create a file name for the downloaded archive
	fileName := filepath.Base(toolInfo.DownloadURL)
	downloadPath := filepath.Join(Config.ToolsDirectory(), fileName)

	// Check if the file already exists
	_, err := os.Stat(downloadPath)
	if os.IsNotExist(err) {
		// Download the file
		log.Printf("Downloading %s v%s...\n", toolInfo.Name, toolInfo.Version)
		downloadPath, err = utils.DownloadFile(toolInfo.DownloadURL, Config.ToolsDirectory())
		if err != nil {
			return fmt.Errorf("failed to download tool: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("error checking for existing download: %w", err)
	} else {
		log.Printf("Using existing download for %s v%s\n", toolInfo.Name, toolInfo.Version)
	}

	// Open the downloaded file
	file, err := os.Open(downloadPath)
	if err != nil {
		return fmt.Errorf("failed to open downloaded file: %w", err)
	}
	defer file.Close()

	// Create a temporary extraction directory
	tempExtractDir := filepath.Join(Config.ToolsDirectory(), fmt.Sprintf("%s-%s-temp", toolInfo.Name, toolInfo.Version))

	// Clean up any previous extraction attempt
	os.RemoveAll(tempExtractDir)

	// Create the temporary extraction directory
	err = os.MkdirAll(tempExtractDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create temporary extraction directory: %w", err)
	}

	// Extract to the temporary directory first
	log.Printf("Extracting %s v%s...\n", toolInfo.Name, toolInfo.Version)
	if strings.HasSuffix(fileName, ".zip") {
		err = utils.ExtractZip(file.Name(), tempExtractDir)
	} else {
		err = utils.ExtractTarGz(file, tempExtractDir)
	}

	if err != nil {
		return fmt.Errorf("failed to extract tool: %w", err)
	}

	// Create the final installation directory
	err = os.MkdirAll(toolInfo.InstallDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create installation directory: %w", err)
	}

	// Find and copy the tool binaries
	for binName, binPath := range toolInfo.Binaries {
		// Get the base name of the binary (without the path)
		binBaseName := filepath.Base(binPath)

		// Try to find the binary in the extracted files
		foundPath := ""

		// First check if it's at the expected location directly
		expectedPath := filepath.Join(tempExtractDir, binBaseName)
		if _, err := os.Stat(expectedPath); err == nil {
			foundPath = expectedPath
		} else {
			// Look for the binary anywhere in the extracted directory
			err := filepath.Walk(tempExtractDir, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if !info.IsDir() && filepath.Base(path) == binBaseName {
					foundPath = path
					return io.EOF // Stop the walk
				}
				return nil
			})

			// io.EOF is expected when we find the file and stop the walk
			if err != nil && err != io.EOF {
				return fmt.Errorf("error searching for %s binary: %w", binName, err)
			}
		}

		if foundPath == "" {
			return fmt.Errorf("could not find %s binary in extracted files", binName)
		}

		// Make sure the destination directory exists
		err = os.MkdirAll(filepath.Dir(binPath), 0755)
		if err != nil {
			return fmt.Errorf("failed to create directory for binary: %w", err)
		}

		// Copy the binary to the installation directory
		input, err := os.Open(foundPath)
		if err != nil {
			return fmt.Errorf("failed to open %s binary: %w", binName, err)
		}
		defer input.Close()

		output, err := os.OpenFile(binPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
		if err != nil {
			return fmt.Errorf("failed to create destination file for %s: %w", binName, err)
		}
		defer output.Close()

		_, err = io.Copy(output, input)
		if err != nil {
			return fmt.Errorf("failed to copy %s binary: %w", binName, err)
		}
	}

	// Clean up the temporary directory
	os.RemoveAll(tempExtractDir)

	log.Printf("Successfully installed %s v%s\n", toolInfo.Name, toolInfo.Version)
	return nil
}

// isToolInstalled checks if a tool is already installed by checking for the binary
func isToolInstalled(toolInfo *plugins.ToolInfo) bool {
	// If there are no binaries, check the install directory
	if len(toolInfo.Binaries) == 0 {
		_, err := os.Stat(toolInfo.InstallDir)
		return err == nil
	}

	// Check if at least one binary exists
	for _, binaryPath := range toolInfo.Binaries {
		_, err := os.Stat(binaryPath)
		if err == nil {
			return true
		}
	}

	return false
}

// executeToolTemplate executes a template with the given data
func executeToolTemplate(tmplStr string, data map[string]string) (string, error) {
	tmpl, err := template.New("command").Parse(tmplStr)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
