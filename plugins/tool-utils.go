package plugins

import (
	"embed"
	"fmt"
	"path"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

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
	Name            string            `yaml:"name"`
	Description     string            `yaml:"description"`
	Runtime         string            `yaml:"runtime"`
	RuntimeBinaries RuntimeBinaries   `yaml:"runtime_binaries"`
	Installation    InstallationConfig `yaml:"installation"`
	Binaries        []ToolBinary      `yaml:"binaries"`
	Formatters      []Formatter       `yaml:"formatters"`
	OutputOptions   OutputOptions     `yaml:"output_options"`
	AnalysisOptions AnalysisOptions   `yaml:"analysis_options"`
}

// ToolConfig represents configuration for a tool
type ToolConfig struct {
	Name     string
	Version  string
	Registry string
}

// ToolInfo contains all processed information about a tool
type ToolInfo struct {
	Name           string
	Version        string
	Runtime        string
	InstallDir     string
	Binaries       map[string]string // Map of binary name to full path
	Formatters     map[string]string // Map of formatter name to flag
	OutputFlag     string
	AutofixFlag    string
	DefaultPath    string
	// Runtime binaries
	PackageManager string
	ExecutionBinary string
	// Installation info
	InstallCommand  string
	RegistryCommand string
}

// ProcessTools processes a list of tool configurations and returns a map of tool information
func ProcessTools(configs []ToolConfig, toolDir string) (map[string]*ToolInfo, error) {
	result := make(map[string]*ToolInfo)
	
	for _, config := range configs {
		// Load the tool plugin
		pluginPath := filepath.Join("tools", config.Name, "plugin.yaml")
		
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
		
		// Create ToolInfo with basic information
		info := &ToolInfo{
			Name:            config.Name,
			Version:         config.Version,
			Runtime:         pluginConfig.Runtime,
			InstallDir:      installDir,
			Binaries:        make(map[string]string),
			Formatters:      make(map[string]string),
			OutputFlag:      pluginConfig.OutputOptions.FileFlag,
			AutofixFlag:     pluginConfig.AnalysisOptions.AutofixFlag,
			DefaultPath:     pluginConfig.AnalysisOptions.DefaultPath,
			// Store runtime binary information
			PackageManager:  pluginConfig.RuntimeBinaries.PackageManager,
			ExecutionBinary: pluginConfig.RuntimeBinaries.Execution,
			// Store raw command templates (processing will happen later)
			InstallCommand:  pluginConfig.Installation.Command,
			RegistryCommand: pluginConfig.Installation.RegistryTemplate,
		}
		
		// Process binary paths
		for _, binary := range pluginConfig.Binaries {
			binaryPath := path.Join(installDir, binary.Path)
			info.Binaries[binary.Name] = binaryPath
		}
		
		// Process formatters
		for _, formatter := range pluginConfig.Formatters {
			info.Formatters[formatter.Name] = formatter.Flag
		}
		
		result[config.Name] = info
	}
	
	return result, nil
}
