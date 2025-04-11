package plugins

import (
	"bytes"
	"embed"
	"fmt"
	"path"
	"path/filepath"
	"runtime"
	"text/template"

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

// ExtensionConfig defines the file extension based on OS
type ExtensionConfig struct {
	Windows string `yaml:"windows"`
	Default string `yaml:"default"`
}

// DownloadConfig holds the download configuration for directly downloading tools
type DownloadConfig struct {
	URLTemplate      string            `yaml:"url_template"`
	FileNameTemplate string            `yaml:"file_name_template"`
	Extension        ExtensionConfig   `yaml:"extension"`
	ArchMapping      map[string]string `yaml:"arch_mapping"`
	OSMapping        map[string]string `yaml:"os_mapping"`
}

// ToolPluginConfig holds the structure of the tool plugin.yaml file
type ToolPluginConfig struct {
	Name            string             `yaml:"name"`
	Description     string             `yaml:"description"`
	Runtime         string             `yaml:"runtime"`
	RuntimeBinaries RuntimeBinaries    `yaml:"runtime_binaries"`
	Installation    InstallationConfig `yaml:"installation"`
	Download        DownloadConfig     `yaml:"download"`
	Binaries        []ToolBinary       `yaml:"binaries"`
	Formatters      []Formatter        `yaml:"formatters"`
	OutputOptions   OutputOptions      `yaml:"output_options"`
	AnalysisOptions AnalysisOptions    `yaml:"analysis_options"`
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
}

// ProcessTools processes a list of tool configurations and returns a map of tool information
func ProcessTools(configs []ToolConfig, toolDir string, runtimes map[string]*RuntimeInfo) (map[string]*ToolInfo, error) {
	result := make(map[string]*ToolInfo)

	for _, config := range configs {
		// Load the tool plugin
		pluginPath := filepath.Join("tools", config.Name, "plugin.yaml")

		// Read from embedded filesystem
		data, err := toolsFS.ReadFile(pluginPath)
		if err != nil {
			return nil, fmt.Errorf("error reading plugin.yaml for %s: %w", config.Name, err)
		}
		fmt.Println("Plugin path", pluginPath)
		var pluginConfig ToolPluginConfig
		err = yaml.Unmarshal(data, &pluginConfig)
		if err != nil {
			return nil, fmt.Errorf("error parsing plugin.yaml for %s: %w", config.Name, err)
		}
		fmt.Println("EOD")
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
		}

		// Handle download configuration for directly downloaded tools
		if pluginConfig.Download.URLTemplate != "" {
			// Get the mapped architecture
			mappedArch := getMappedArch(pluginConfig.Download.ArchMapping, runtime.GOARCH)

			// Get the mapped OS
			mappedOS := getMappedOS(pluginConfig.Download.OSMapping, runtime.GOOS)

			// Get the appropriate extension
			extension := getExtension(pluginConfig.Download.Extension, runtime.GOOS)
			info.Extension = extension

			// Get the filename using the template
			fileName := getFileName(pluginConfig.Download.FileNameTemplate, config.Version, mappedArch, runtime.GOOS)
			info.FileName = fileName

			// Get the download URL using the template
			downloadURL := getDownloadURL(pluginConfig.Download.URLTemplate, fileName, config.Version, mappedArch, mappedOS, extension)
			info.DownloadURL = downloadURL
		}

		// Process binary paths
		for _, binary := range pluginConfig.Binaries {
			// Process template variables in binary path
			tmpl, err := template.New("binary_path").Parse(binary.Path)
			if err != nil {
				return nil, fmt.Errorf("error parsing binary path template for %s: %w", config.Name, err)
			}

			var buf bytes.Buffer
			err = tmpl.Execute(&buf, struct {
				Version string
			}{
				Version: config.Version,
			})
			if err != nil {
				return nil, fmt.Errorf("error executing binary path template for %s: %w", config.Name, err)
			}

			binaryPath := filepath.Join(installDir, buf.String())
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

// Helper functions for processing download configuration

// getMappedArch returns the architecture mapping for the current system
func getMappedArch(archMapping map[string]string, goarch string) string {
	// Check if there's a mapping for this architecture
	if mappedArch, ok := archMapping[goarch]; ok {
		return mappedArch
	}
	// Return the original architecture if no mapping exists
	return goarch
}

// getMappedOS returns the OS mapping for the current system
func getMappedOS(osMapping map[string]string, goos string) string {
	// Check if there's a mapping for this OS
	if mappedOS, ok := osMapping[goos]; ok {
		return mappedOS
	}
	// Return the original OS if no mapping exists
	return goos
}

// getExtension returns the appropriate file extension based on the OS
func getExtension(extensionConfig ExtensionConfig, goos string) string {
	if goos == "windows" {
		return extensionConfig.Windows
	}
	return extensionConfig.Default
}

// getFileName generates the filename based on the template
func getFileName(fileNameTemplate string, version string, mappedArch string, goos string) string {
	// Prepare template data
	data := struct {
		Version string
		OS      string
		Arch    string
	}{
		Version: version,
		OS:      goos,
		Arch:    mappedArch,
	}

	// Execute template substitution for filename
	tmpl, err := template.New("filename").Parse(fileNameTemplate)
	if err != nil {
		return ""
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return ""
	}

	return buf.String()
}

// getDownloadURL generates the download URL based on the template
func getDownloadURL(urlTemplate string, fileName string, version string, mappedArch string, mappedOS string, extension string) string {
	// Prepare template data
	data := struct {
		Version   string
		FileName  string
		OS        string
		Arch      string
		Extension string
	}{
		Version:   version,
		FileName:  fileName,
		OS:        mappedOS,
		Arch:      mappedArch,
		Extension: extension,
	}

	// Execute template substitution for URL
	tmpl, err := template.New("url").Parse(urlTemplate)
	if err != nil {
		return ""
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return ""
	}

	return buf.String()
}

// GetSupportedTools returns a map of supported tool names based on the tools folder
func GetSupportedTools() (map[string]struct{}, error) {
	supportedTools := make(map[string]struct{})

	// Read all directories in the tools folder
	entries, err := toolsFS.ReadDir("tools")
	if err != nil {
		return nil, fmt.Errorf("failed to read tools directory: %w", err)
	}

	// For each directory, check if it has a plugin.yaml file
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		toolName := entry.Name()
		pluginPath := filepath.Join("tools", toolName, "plugin.yaml")

		// Check if plugin.yaml exists
		_, err := toolsFS.ReadFile(pluginPath)
		if err != nil {
			continue // Skip if no plugin.yaml
		}

		supportedTools[toolName] = struct{}{}
	}

	return supportedTools, nil
}
