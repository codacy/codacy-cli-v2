package plugins

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
)

const toolPluginYamlPathTemplate = "tools/%s/plugin.yaml"

//go:embed tools/*/plugin.yaml
var toolsFS embed.FS

// ToolBinary represents a binary executable provided by the tool
type ToolBinary struct {
	Name string `yaml:"name"`
	Path string `yaml:"path"`
}

// Formatter represents a supported output format
type Formatter struct {
	Name string `yaml:"name"`
	Flag string `yaml:"flag"`
}

// InstallationConfig holds the installation configuration from the plugin.yaml
type InstallationConfig struct {
	Command          string `yaml:"command"`
	RegistryTemplate string `yaml:"registry_template"`
}

// OutputOptions holds configuration for output handling
type OutputOptions struct {
	FileFlag string `yaml:"file_flag"`
}

// AnalysisOptions holds configuration for analysis options
type AnalysisOptions struct {
	AutofixFlag string `yaml:"autofix_flag"`
	DefaultPath string `yaml:"default_path"`
}

// RuntimeBinaries holds the mapping of runtime binary names
type RuntimeBinaries struct {
	PackageManager string `yaml:"package_manager"`
	Execution      string `yaml:"execution"`
}

// ToolPluginConfig holds the structure of the tool plugin.yaml file
type ToolPluginConfig struct {
	Name            string             `yaml:"name"`
	Description     string             `yaml:"description"`
	DefaultVersion  string             `yaml:"default_version"`
	Runtime         string             `yaml:"runtime"`
	RuntimeBinaries RuntimeBinaries    `yaml:"runtime_binaries"`
	Installation    InstallationConfig `yaml:"installation"`
	Download        DownloadConfig     `yaml:"download"`
	Environment     map[string]string  `yaml:"environment"`
	Binaries        []ToolBinary       `yaml:"binaries"`
	Formatters      []Formatter        `yaml:"formatters"`
	OutputOptions   OutputOptions      `yaml:"output_options"`
	AnalysisOptions AnalysisOptions    `yaml:"analysis_options"`
	ConfigFileName  string             `yaml:"config_file_name,omitempty"` // Optional field
}

// ToolConfig represents configuration for a tool
type ToolConfig struct {
	Name     string
	Version  string
	Registry string
}

// ToolInfo contains all processed information about a tool
type ToolInfo struct {
	Name        string
	Version     string
	Runtime     string
	InstallDir  string
	Binaries    map[string]string // Map of binary name to full path
	Formatters  map[string]string // Map of formatter name to flag
	OutputFlag  string
	AutofixFlag string
	DefaultPath string
	// Runtime binaries
	PackageManager  string
	ExecutionBinary string
	// Installation info
	InstallCommand  string
	RegistryCommand string
	// Download info for binary tools
	DownloadURL string
	FileName    string
	Extension   string
	// Environment variables
	Environment map[string]string
	// Config file
	ConfigFileName string
}

// ProcessTools processes a list of tool configurations and returns a map of tool information
func ProcessTools(configs []ToolConfig, toolDir string, runtimes map[string]*RuntimeInfo) (map[string]*ToolInfo, error) {
	result := make(map[string]*ToolInfo)

	for _, config := range configs {
		// Load the tool plugin - always use forward slashes for embedded filesystem paths (for windows support)
		pluginPath := fmt.Sprintf(toolPluginYamlPathTemplate, config.Name)

		// Read from embedded filesystem
		data, err := toolsFS.ReadFile(pluginPath)
		if err != nil {
			return nil, fmt.Errorf("error reading plugin.yaml for %s: %w", config.Name, err)
		}
		var pluginConfig ToolPluginConfig
		err = yaml.Unmarshal(data, &pluginConfig)
		if err != nil {
			return nil, fmt.Errorf("error parsing plugin.yaml for %s: %w", config.Name, err)
		}
		// Create the install directory path
		installDir := path.Join(toolDir, fmt.Sprintf("%s@%s", config.Name, config.Version))

		// Handle special case for dartanalyzer since it can be used with either dart or flutter
		toolRuntime := pluginConfig.Runtime
		if config.Name == "dartanalyzer" {
			if runtimes["flutter"] != nil {
				installDir = runtimes["flutter"].InstallDir
				toolRuntime = "flutter"
			} else {
				installDir = runtimes["dart"].InstallDir
				toolRuntime = "dart"
			}
		}
		// Create ToolInfo with basic information
		info := &ToolInfo{
			Name:        config.Name,
			Version:     config.Version,
			Runtime:     toolRuntime,
			InstallDir:  installDir,
			Binaries:    make(map[string]string),
			Formatters:  make(map[string]string),
			OutputFlag:  pluginConfig.OutputOptions.FileFlag,
			AutofixFlag: pluginConfig.AnalysisOptions.AutofixFlag,
			DefaultPath: pluginConfig.AnalysisOptions.DefaultPath,
			// Store runtime binary information
			PackageManager:  pluginConfig.RuntimeBinaries.PackageManager,
			ExecutionBinary: pluginConfig.RuntimeBinaries.Execution,
			// Store raw command templates (processing will happen later)
			InstallCommand:  pluginConfig.Installation.Command,
			RegistryCommand: pluginConfig.Installation.RegistryTemplate,
			// Store environment variables
			Environment: make(map[string]string),
			// Store config file name
			ConfigFileName: pluginConfig.ConfigFileName,
		}

		// Handle download configuration for directly downloaded tools
		if pluginConfig.Download.URLTemplate != "" {
			// Get the mapped architecture
			mappedArch := GetMappedArch(pluginConfig.Download.ArchMapping, runtime.GOARCH)

			// Get the mapped OS
			mappedOS := GetMappedOS(pluginConfig.Download.OSMapping, runtime.GOOS)

			// Get the appropriate extension
			extension := GetExtension(pluginConfig.Download.Extension, runtime.GOOS)
			info.Extension = extension

			// Get the filename using the template
			fileName := GetFileName(pluginConfig.Download.FileNameTemplate, config.Version, mappedArch, runtime.GOOS)
			info.FileName = fileName

			// Get the download URL using the template
			downloadURL := GetDownloadURL(pluginConfig.Download.URLTemplate, fileName, config.Version, mappedArch, mappedOS, extension, pluginConfig.Download.ReleaseVersion)
			info.DownloadURL = downloadURL
		}

		// Process binary paths
		for _, binary := range pluginConfig.Binaries {
			var binaryPath string

			// Process template variables in binary path
			tmpl, err := template.New("binary_path").Parse(binary.Path)
			if err != nil {
				return nil, fmt.Errorf("error parsing binary path template for %s: %w", config.Name, err)
			}

			var buf bytes.Buffer
			err = tmpl.Execute(&buf, struct {
				Version    string
				InstallDir string
			}{
				Version:    config.Version,
				InstallDir: installDir,
			})
			if err != nil {
				return nil, fmt.Errorf("error executing binary path template for %s: %w", config.Name, err)
			}

			binaryPath = buf.String()

			// If the binary path is relative, join it with the install directory
			if !path.IsAbs(binaryPath) {
				binaryPath = path.Join(installDir, binaryPath)
			}

			// Add file extension for Windows executables
			if runtime.GOOS == "windows" && !strings.HasSuffix(binaryPath, ".exe") {
				binaryPath += ".exe"
			}

			info.Binaries[binary.Name] = binaryPath
		}

		// Process formatters
		for _, formatter := range pluginConfig.Formatters {
			info.Formatters[formatter.Name] = formatter.Flag
		}

		// Process environment variables
		for key, value := range pluginConfig.Environment {
			// Process template variables in environment value
			tmpl, err := template.New("env_value").Parse(value)
			if err != nil {
				return nil, fmt.Errorf("error parsing environment value template for %s: %w", config.Name, err)
			}

			var buf bytes.Buffer
			err = tmpl.Execute(&buf, struct {
				Version           string
				InstallDir        string
				RuntimeInstallDir string
				Path              string
			}{
				Version:           config.Version,
				InstallDir:        installDir,
				RuntimeInstallDir: runtimes[toolRuntime].InstallDir,
				Path:              os.Getenv("PATH"),
			})
			if err != nil {
				return nil, fmt.Errorf("error executing environment value template for %s: %w", config.Name, err)
			}

			info.Environment[key] = buf.String()
		}

		result[config.Name] = info
	}

	return result, nil
}

// GetSupportedTools returns a map of supported tool names
func GetSupportedTools() (map[string]struct{}, error) {
	entries, err := toolsFS.ReadDir("tools")
	if err != nil {
		return nil, fmt.Errorf("error reading tools directory: %w", err)
	}

	tools := make(map[string]struct{})
	for _, entry := range entries {
		if entry.IsDir() {
			tools[entry.Name()] = struct{}{}
		}
	}

	return tools, nil
}

// GetToolVersions returns a map of tool names to their default versions
func GetToolVersions() map[string]string {
	return GetPluginManager().GetToolVersions()
}

// GetToolRuntimeDependencies returns a map of tool names to their runtime dependencies
func GetToolRuntimeDependencies() map[string]string {
	return GetPluginManager().GetToolRuntimeDependencies()
}
