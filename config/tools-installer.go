package config

import (
	"bytes"
	"codacy/cli-v2/constants"
	"codacy/cli-v2/plugins"
	"codacy/cli-v2/utils"
	"codacy/cli-v2/utils/logger"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"

	"github.com/sirupsen/logrus"
)

// InstallTools installs all tools defined in the configuration
func InstallTools(config *ConfigType, registry string) error {
	var failedTools []string

	for name, toolInfo := range config.Tools() {
		logger.Info("Starting tool installation", logrus.Fields{
			"tool":    name,
			"version": toolInfo.Version,
			"runtime": toolInfo.Runtime,
		})

		fmt.Printf("Installing tool: %s v%s...\n", name, toolInfo.Version)
		err := InstallTool(name, toolInfo, registry)
		if err != nil {
			logger.Error("Failed to install tool", logrus.Fields{
				"tool":    name,
				"version": toolInfo.Version,
				"runtime": toolInfo.Runtime,
				"error":   err.Error(),
			})
			failedTools = append(failedTools, name)
			continue
		}

		logger.Info("Successfully installed tool", logrus.Fields{
			"tool":    name,
			"version": toolInfo.Version,
			"runtime": toolInfo.Runtime,
		})
		fmt.Printf("Successfully installed %s v%s\n", name, toolInfo.Version)
	}

	if len(failedTools) > 0 {
		return fmt.Errorf("failed to install the following tools: %v", failedTools)
	}
	return nil
}

// InstallTool installs a specific tool
func InstallTool(name string, toolInfo *plugins.ToolInfo, registry string) error {

	// If toolInfo is nil, it means the tool is not in the configuration
	// Add it with the default version
	if toolInfo == nil {
		logger.Info("Tool not found in configuration, adding with default version", logrus.Fields{
			"tool": name,
		})

		// Add the tool to the configuration with its default version
		err := Config.AddToolWithDefaultVersion(name)
		if err != nil {
			return fmt.Errorf("failed to add tool %s to configuration: %w", name, err)
		}

		// Get the tool info from the updated configuration
		toolInfo = Config.Tools()[name]
		if toolInfo == nil {
			return fmt.Errorf("tool %s still not found in configuration after adding", name)
		}
	}

	// Check if the tool is already installed AND its runtime is available
	isToolInstalled := Config.IsToolInstalled(name, toolInfo)

	// Also check if the required runtime is installed
	var isRuntimeInstalled bool
	if toolInfo.Runtime != "" {
		runtimeInfo, exists := Config.Runtimes()[toolInfo.Runtime]
		isRuntimeInstalled = exists && Config.IsRuntimeInstalled(toolInfo.Runtime, runtimeInfo)
	} else {
		// No runtime dependency
		isRuntimeInstalled = true
	}

	if isToolInstalled && isRuntimeInstalled {
		logger.Info("Tool and runtime already installed", logrus.Fields{
			"tool":    name,
			"version": toolInfo.Version,
			"runtime": toolInfo.Runtime,
		})
		fmt.Printf("Tool %s v%s and its runtime are already installed\n", name, toolInfo.Version)
		return nil
	}

	// Make sure the installation directory exists
	err := os.MkdirAll(toolInfo.InstallDir, constants.DefaultDirPerms)
	if err != nil {
		return fmt.Errorf("failed to create installation directory: %w", err)
	}

	// Check if this is a download-based tool (like trivy) or a runtime-based tool (like eslint)
	if toolInfo.DownloadURL != "" {
		// This is a download-based tool
		logger.Debug("Installing download-based tool", logrus.Fields{
			"tool":        name,
			"version":     toolInfo.Version,
			"downloadURL": toolInfo.DownloadURL,
		})
		fmt.Printf("Downloading %s...\n", name)
		return installDownloadBasedTool(toolInfo)
	}

	// Handle Python tools differently
	if toolInfo.Runtime == "python" {
		logger.Debug("Installing Python tool", logrus.Fields{
			"tool":    name,
			"version": toolInfo.Version,
		})
		fmt.Printf("Installing Python tool %s...\n", name)
		return installPythonTool(name, toolInfo)
	}

	// For runtime-based tools
	logger.Debug("Installing runtime-based tool", logrus.Fields{
		"tool":    name,
		"version": toolInfo.Version,
		"runtime": toolInfo.Runtime,
	})
	fmt.Printf("Installing %s using %s runtime...\n", name, toolInfo.Runtime)
	return installRuntimeTool(name, toolInfo, registry)
}

func installRuntimeTool(name string, toolInfo *plugins.ToolInfo, registry string) error {
	// Get the runtime configuration for this tool
	runtimeInfo, ok := Config.Runtimes()[toolInfo.Runtime]
	if !ok {
		return fmt.Errorf("required runtime %s not found for tool %s", toolInfo.Runtime, name)
	}

	// Check if the runtime is actually installed
	if !Config.IsRuntimeInstalled(toolInfo.Runtime, runtimeInfo) {
		fmt.Printf("Runtime %s v%s is not installed, installing...\n", toolInfo.Runtime, runtimeInfo.Version)
		err := InstallRuntime(toolInfo.Runtime, runtimeInfo)
		if err != nil {
			return fmt.Errorf("failed to install runtime %s: %w", toolInfo.Runtime, err)
		}
		fmt.Printf("Runtime %s v%s installed successfully, proceeding with tool installation...\n", toolInfo.Runtime, runtimeInfo.Version)
	}

	// Prepare template data
	templateData := map[string]string{
		"InstallDir":  toolInfo.InstallDir,
		"PackageName": toolInfo.Name,
		"Version":     toolInfo.Version,
		"Registry":    registry,
	}

	// Get package manager binary based on the tool configuration
	packageManagerName := toolInfo.PackageManager
	packageManagerBinary, ok := runtimeInfo.Binaries[packageManagerName]
	if !ok {
		return fmt.Errorf("package manager binary %s not found in runtime %s", packageManagerName, toolInfo.Runtime)
	}

	// Set registry if provided
	if registry != "" {
		regCmd, err := executeToolTemplate(toolInfo.RegistryCommand, templateData)
		if err != nil {
			return fmt.Errorf("failed to prepare registry command: %w", err)
		}

		logger.Debug("Setting registry", logrus.Fields{
			"tool":              name,
			"packageManager":    packageManagerName,
			"packageManagerBin": packageManagerBinary,
			"command":           regCmd,
		})

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

	logger.Debug("Installing tool", logrus.Fields{
		"tool":              name,
		"version":           toolInfo.Version,
		"packageManager":    packageManagerName,
		"packageManagerBin": packageManagerBinary,
		"command":           installCmd,
	})

	// Execute the installation command using the package manager
	cmd := exec.Command(packageManagerBinary, strings.Split(installCmd, " ")...)

	// Special handling for Go tools: set GOBIN so the binary is installed in the tool's install directory
	if toolInfo.Runtime == "go" {
		env := os.Environ()
		env = append(env, "GOBIN="+toolInfo.InstallDir)
		cmd.Env = env
	}

	// Special handling for ESLint installation in Linux (WSL) environment
	if toolInfo.Name == "eslint" && runtime.GOOS == "linux" {
		// Get node binary directory to add to PATH
		nodeBinary, exist := runtimeInfo.Binaries["node"]
		if exist {
			nodeDir := filepath.Dir(nodeBinary)
			// Get current PATH
			currentPath := os.Getenv("PATH")
			// For Linux (WSL), always use Linux path separator
			pathSeparator := ":"
			newPath := nodeDir + pathSeparator + currentPath
			cmd.Env = append(os.Environ(), "PATH="+newPath)
			logger.Debug("Setting PATH environment for ESLint installation", logrus.Fields{
				"nodeDir":     nodeDir,
				"currentPath": currentPath,
				"newPath":     newPath,
			})
		}
	}

	log.Printf("Installing %s v%s...\n", toolInfo.Name, toolInfo.Version)
	logger.Debug("Running command", logrus.Fields{
		"command": fmt.Sprintf("%s %s", packageManagerBinary, installCmd),
	})
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to install tool: %s: %w", string(output), err)
	}

	logger.Debug("Tool installation completed", logrus.Fields{
		"tool":    name,
		"version": toolInfo.Version,
	})
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
		logger.Debug("Downloading tool", logrus.Fields{
			"tool":         toolInfo.Name,
			"version":      toolInfo.Version,
			"downloadURL":  toolInfo.DownloadURL,
			"downloadPath": downloadPath,
		})
		downloadPath, err = utils.DownloadFile(toolInfo.DownloadURL, Config.ToolsDirectory())
		if err != nil {
			return fmt.Errorf("failed to download tool: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("error checking for existing download: %w", err)
	} else {
		logger.Debug("Using existing tool download", logrus.Fields{
			"tool":         toolInfo.Name,
			"version":      toolInfo.Version,
			"downloadPath": downloadPath,
		})
	}

	// Open the downloaded file
	file, err := os.Open(downloadPath)
	if err != nil {
		return fmt.Errorf("failed to open downloaded file: %w", err)
	}
	defer file.Close()

	// Create the installation directory
	err = os.MkdirAll(toolInfo.InstallDir, constants.DefaultDirPerms)
	if err != nil {
		return fmt.Errorf("failed to create installation directory: %w", err)
	}

	// Extract based on file extension
	logger.Debug("Extracting tool", logrus.Fields{
		"tool":             toolInfo.Name,
		"version":          toolInfo.Version,
		"fileName":         fileName,
		"extractDirectory": toolInfo.InstallDir,
	})

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
		err = os.Chmod(filepath.Join(toolInfo.InstallDir, filepath.Base(binaryPath)), constants.DefaultDirPerms)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to make binary executable: %w", err)
		}
	}

	logger.Debug("Tool extraction completed", logrus.Fields{
		"tool":    toolInfo.Name,
		"version": toolInfo.Version,
	})
	return nil
}

func installPythonTool(name string, toolInfo *plugins.ToolInfo) error {
	logger.Debug("Starting Python tool installation", logrus.Fields{
		"tool":    toolInfo.Name,
		"version": toolInfo.Version,
	})

	// Get the runtime configuration for this tool
	runtimeInfo, ok := Config.Runtimes()[toolInfo.Runtime]
	if !ok {
		return fmt.Errorf("required runtime %s not found for tool %s", toolInfo.Runtime, name)
	}

	// Check if the Python runtime is actually installed
	if !Config.IsRuntimeInstalled(toolInfo.Runtime, runtimeInfo) {
		fmt.Printf("Python runtime %s v%s is not installed, installing...\n", toolInfo.Runtime, runtimeInfo.Version)
		err := InstallRuntime(toolInfo.Runtime, runtimeInfo)
		if err != nil {
			return fmt.Errorf("failed to install runtime %s: %w", toolInfo.Runtime, err)
		}
		fmt.Printf("Python runtime %s v%s installed successfully, proceeding with tool installation...\n", toolInfo.Runtime, runtimeInfo.Version)
	}

	pythonBinary, ok := runtimeInfo.Binaries["python3"]
	if !ok {
		return fmt.Errorf("python3 binary not found in runtime")
	}

	// Create venv
	logger.Debug("Creating Python virtual environment", logrus.Fields{
		"tool":    toolInfo.Name,
		"version": toolInfo.Version,
		"venvDir": filepath.Join(toolInfo.InstallDir, "venv"),
	})

	cmd := exec.Command(pythonBinary, "-m", "venv", filepath.Join(toolInfo.InstallDir, "venv"))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create venv: %s\nError: %w", string(output), err)
	}

	// Install the tool using pip from venv
	pipPath := filepath.Join(toolInfo.InstallDir, "venv", "bin", "pip")
	logger.Debug("Installing Python package", logrus.Fields{
		"tool":    toolInfo.Name,
		"version": toolInfo.Version,
		"pipPath": pipPath,
	})

	cmd = exec.Command(pipPath, "install", fmt.Sprintf("%s==%s", toolInfo.Name, toolInfo.Version))
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to install tool: %s\nError: %w", string(output), err)
	}

	logger.Debug("Python tool installation completed", logrus.Fields{
		"tool":    toolInfo.Name,
		"version": toolInfo.Version,
	})
	return nil
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
