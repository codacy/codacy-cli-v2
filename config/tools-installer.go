package config

import (
	"bytes"
	"codacy/cli-v2/plugins"
	"codacy/cli-v2/utils"
	"fmt"
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
		fmt.Printf("Installing tool: %s v%s...\n", name, toolInfo.Version)
		err := InstallTool(name, toolInfo)
		if err != nil {
			return fmt.Errorf("failed to install tool %s: %w", name, err)
		}
		fmt.Printf("Successfully installed %s v%s\n", name, toolInfo.Version)
	}
	return nil
}

// InstallTool installs a specific tool
func InstallTool(name string, toolInfo *plugins.ToolInfo) error {
	fmt.Println("Installing tool", name, "in", toolInfo.InstallDir)
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
		fmt.Printf("Downloading %s...\n", name)
		return installDownloadBasedTool(toolInfo)
	}

	// Handle Python tools differently
	if toolInfo.Runtime == "python" {
		fmt.Printf("Installing Python tool %s...\n", name)
		return installPythonTool(name, toolInfo)
	}

	// For runtime-based tools
	fmt.Printf("Installing %s using %s runtime...\n", name, toolInfo.Runtime)
	return installRuntimeTool(name, toolInfo)
}

func installRuntimeTool(name string, toolInfo *plugins.ToolInfo) error {
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

	// Create the installation directory
	err = os.MkdirAll(toolInfo.InstallDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create installation directory: %w", err)
	}

	// Extract directly to the installation directory
	log.Printf("Extracting %s v%s...\n", toolInfo.Name, toolInfo.Version)
	if strings.HasSuffix(fileName, ".zip") {
		err = utils.ExtractZip(file.Name(), toolInfo.InstallDir)
	} else {
		err = utils.ExtractTarGz(file, toolInfo.InstallDir)
	}

	if err != nil {
		return fmt.Errorf("failed to extract tool: %w", err)
	}

	// Make sure all binaries are executable
	for _, binaryPath := range toolInfo.Binaries {
		err = os.Chmod(filepath.Join(toolInfo.InstallDir, filepath.Base(binaryPath)), 0755)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to make binary executable: %w", err)
		}
	}

	log.Printf("Successfully installed %s v%s\n", toolInfo.Name, toolInfo.Version)
	return nil
}

func installPythonTool(name string, toolInfo *plugins.ToolInfo) error {
	log.Printf("Installing %s v%s...\n", toolInfo.Name, toolInfo.Version)

	runtimeInfo, ok := Config.Runtimes()[toolInfo.Runtime]
	if !ok {
		return fmt.Errorf("required runtime %s not found for tool %s", toolInfo.Runtime, name)
	}

	pythonBinary, ok := runtimeInfo.Binaries["python3"]
	if !ok {
		return fmt.Errorf("python3 binary not found in runtime")
	}

	// Create venv
	cmd := exec.Command(pythonBinary, "-m", "venv", filepath.Join(toolInfo.InstallDir, "venv"))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create venv: %s\nError: %w", string(output), err)
	}

	// Install the tool using pip from venv
	pipPath := filepath.Join(toolInfo.InstallDir, "venv", "bin", "pip")
	cmd = exec.Command(pipPath, "install", fmt.Sprintf("%s==%s", toolInfo.Name, toolInfo.Version))
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to install tool: %s\nError: %w", string(output), err)
	}

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
