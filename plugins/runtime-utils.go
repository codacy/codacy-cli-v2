package plugins

import (
	"bytes"
	"embed"
	"fmt"
	"path"
	"runtime"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
)

//go:embed runtimes/*/plugin.yaml
var pluginsFS embed.FS

// binary represents a binary executable provided by the runtime
type binary struct {
	Name string      `yaml:"name"`
	Path interface{} `yaml:"path"` // Can be either string or map[string]string
}

// binaryPath represents OS-specific paths for a binary
type binaryPath struct {
	Darwin string `yaml:"darwin"`
	Linux  string `yaml:"linux"`
}

// pluginConfig holds the structure of the plugin.yaml file
type pluginConfig struct {
	Name        string         `yaml:"name"`
	Description string         `yaml:"description"`
	Download    downloadConfig `yaml:"download"`
	Binaries    []binary       `yaml:"binaries"`
}

// downloadConfig holds the download configuration from the plugin.yaml
type downloadConfig struct {
	URLTemplate      string            `yaml:"url_template"`
	FileNameTemplate string            `yaml:"file_name_template"`
	Extension        extensionConfig   `yaml:"extension"`
	ArchMapping      map[string]string `yaml:"arch_mapping"`
	OSMapping        map[string]string `yaml:"os_mapping"`
	ReleaseVersion   string            `yaml:"release_version"`
}

// extensionConfig defines the file extension based on OS
type extensionConfig struct {
	Windows string `yaml:"windows"`
	Default string `yaml:"default"`
}

// templateData holds the data to be used in template substitution
type templateData struct {
	Version        string
	MajorVersion   string
	FileName       string
	OS             string
	Arch           string
	Extension      string
	ReleaseVersion string
}

// runtimePlugin represents a runtime plugin with methods to interact with it
type runtimePlugin struct {
	Config     pluginConfig
	ConfigPath string
}

// RuntimeConfig represents configuration for a runtime
type RuntimeConfig struct {
	Name    string
	Version string
}

// RuntimeInfo contains all processed information about a runtime
type RuntimeInfo struct {
	Name        string
	Version     string
	InstallDir  string
	DownloadURL string
	FileName    string
	Extension   string
	Binaries    map[string]string // Map of binary name to full path
}

// ProcessRuntimes processes a list of runtime configurations and returns a map of runtime information
func ProcessRuntimes(configs []RuntimeConfig, runtimesDir string) (map[string]*RuntimeInfo, error) {
	result := make(map[string]*RuntimeInfo)

	for _, config := range configs {
		runtimeInfo, err := processRuntime(config, runtimesDir)
		if err != nil {
			return nil, err
		}

		result[config.Name] = runtimeInfo
	}

	return result, nil
}

// ProcessRuntime processes a single runtime configuration and returns detailed runtime info
func processRuntime(config RuntimeConfig, runtimesDir string) (*RuntimeInfo, error) {
	plugin, err := loadPlugin(config.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to load plugin for runtime %s: %w", config.Name, err)
	}

	fileName := plugin.getFileName(config.Version)
	extension := plugin.getExtension(runtime.GOOS)
	installDir := plugin.getInstallationDirectoryPath(runtimesDir, config.Version)

	// Create RuntimeInfo with essential information
	info := &RuntimeInfo{
		Name:        config.Name,
		Version:     config.Version,
		InstallDir:  installDir,
		DownloadURL: plugin.getDownloadURL(config.Version),
		FileName:    fileName,
		Extension:   extension,
		Binaries:    make(map[string]string),
	}

	// Process binary paths
	for _, binary := range plugin.Config.Binaries {
		var binaryPath string

		switch path := binary.Path.(type) {
		case string:
			// If path is a simple string, use it directly
			binaryPath = path
		case map[string]interface{}:
			// If path is a map, get the OS-specific path
			if osPath, ok := path[runtime.GOOS]; ok {
				binaryPath = osPath.(string)
			} else {
				return nil, fmt.Errorf("no binary path specified for OS %s", runtime.GOOS)
			}
		default:
			return nil, fmt.Errorf("invalid path format for binary %s", binary.Name)
		}

		fullPath := path.Join(installDir, binaryPath)

		// Add file extension for Windows executables
		if runtime.GOOS == "windows" && !strings.HasSuffix(fullPath, ".exe") {
			fullPath += ".exe"
		}

		info.Binaries[binary.Name] = fullPath
	}

	return info, nil
}

// LoadPlugin loads a plugin configuration from the specified plugin directory
func loadPlugin(runtimeName string) (*runtimePlugin, error) {
	// Always use forward slashes for embedded filesystem paths (for windows support)
	pluginPath := fmt.Sprintf("runtimes/%s/plugin.yaml", runtimeName)

	// Read from embedded filesystem
	data, err := pluginsFS.ReadFile(pluginPath)
	if err != nil {
		return nil, fmt.Errorf("error reading plugin.yaml: %w", err)
	}

	var config pluginConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("error parsing plugin.yaml: %w", err)
	}

	return &runtimePlugin{
		Config:     config,
		ConfigPath: pluginPath,
	}, nil
}

// GetMappedArch returns the architecture mapping for the current system
func (p *runtimePlugin) getMappedArch(goarch string) string {
	// Check if there's a mapping for this architecture
	if mappedArch, ok := p.Config.Download.ArchMapping[goarch]; ok {
		return mappedArch
	}
	// Return the original architecture if no mapping exists
	return goarch
}

// GetExtension returns the appropriate file extension based on the OS
func (p *runtimePlugin) getExtension(goos string) string {
	if goos == "windows" {
		return p.Config.Download.Extension.Windows
	}
	return p.Config.Download.Extension.Default
}

// GetFileName generates the filename based on the template in plugin.yaml
func (p *runtimePlugin) getFileName(version string) string {
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	// Map Go architecture and OS to runtime-specific values
	mappedArch := p.getMappedArch(goarch)
	mappedOS := p.getMappedOS(goos)
	releaseVersion := p.getReleaseVersion()

	// Extract major version from version string (e.g. "17.0.10" -> "17")
	majorVersion := version
	if idx := strings.Index(version, "."); idx != -1 {
		majorVersion = version[:idx]
	}

	// Prepare template data
	data := templateData{
		Version:        version,
		MajorVersion:   majorVersion,
		OS:             mappedOS,
		Arch:           mappedArch,
		ReleaseVersion: releaseVersion,
	}

	// Execute template substitution for filename
	tmpl, err := template.New("filename").Parse(p.Config.Download.FileNameTemplate)
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

// GetReleaseVersion returns the release version from the plugin configuration
func (p *runtimePlugin) getReleaseVersion() string {
	return p.Config.Download.ReleaseVersion
}

// GetDownloadURL generates the download URL based on the template in plugin.yaml
func (p *runtimePlugin) getDownloadURL(version string) string {
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	// Map Go architecture and OS to runtime-specific values
	mappedArch := p.getMappedArch(goarch)
	mappedOS := p.getMappedOS(goos)

	// Get the appropriate extension
	extension := p.getExtension(goos)

	// Get the filename
	fileName := p.getFileName(version)

	releaseVersion := p.getReleaseVersion()

	// Extract major version from version string (e.g. "17.0.10" -> "17")
	majorVersion := version
	if idx := strings.Index(version, "."); idx != -1 {
		majorVersion = version[:idx]
	}

	// Prepare template data
	data := templateData{
		Version:        version,
		MajorVersion:   majorVersion,
		FileName:       fileName,
		OS:             mappedOS,
		Arch:           mappedArch,
		Extension:      extension,
		ReleaseVersion: releaseVersion,
	}

	// Execute template substitution for URL
	tmpl, err := template.New("url").Parse(p.Config.Download.URLTemplate)
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

// GetInstallationDirectoryPath returns the installation directory path for the runtime
func (p *runtimePlugin) getInstallationDirectoryPath(runtimesDir string, version string) string {
	// For Python, we want to use a simpler directory structure
	if p.Config.Name == "python" {
		return path.Join(runtimesDir, "python")
	}
	// For other runtimes, keep using the filename-based directory
	fileName := p.getFileName(version)
	return path.Join(runtimesDir, fileName)
}

// GetMappedOS returns the OS mapping for the current system
func (p *runtimePlugin) getMappedOS(goos string) string {
	// Check if there's a mapping for this OS
	if mappedOS, ok := p.Config.Download.OSMapping[goos]; ok {
		return mappedOS
	}
	// Return the original OS if no mapping exists
	return goos
}
