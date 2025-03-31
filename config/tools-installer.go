package config

import (
	"bytes"
	"codacy/cli-v2/plugins"
	"fmt"
	"log"
	"os"
	"os/exec"
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

	// Get the runtime for this tool
	runtimeInfo, ok := Config.Runtimes()[toolInfo.Runtime]
	if !ok {
		return fmt.Errorf("required runtime %s not found for tool %s", toolInfo.Runtime, name)
	}

	// Make sure the installation directory exists
	err := os.MkdirAll(toolInfo.InstallDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create installation directory: %w", err)
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
