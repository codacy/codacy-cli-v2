package plugins

import (
	"embed"
	"fmt"
	"path"
	"runtime"
	"strings"

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
	Name           string         `yaml:"name"`
	Description    string         `yaml:"description"`
	Download       DownloadConfig `yaml:"download"`
	Binaries       []binary       `yaml:"binaries"`
	DefaultVersion string         `yaml:"default_version"`
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
	pluginConfig, err := GetPluginManager().GetRuntimeConfig(config.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to load plugin for runtime %s: %w", config.Name, err)
	}

	// Map Go architecture and OS to runtime-specific values
	mappedArch := GetMappedArch(pluginConfig.Download.ArchMapping, runtime.GOARCH)
	mappedOS := GetMappedOS(pluginConfig.Download.OSMapping, runtime.GOOS)

	// Get the appropriate extension
	extension := GetExtension(pluginConfig.Download.Extension, runtime.GOOS)

	// Get the filename using the template
	fileName := GetFileName(pluginConfig.Download.FileNameTemplate, config.Version, mappedArch, runtime.GOOS)

	// Get the download URL using the template
	downloadURL := GetDownloadURL(pluginConfig.Download.URLTemplate, fileName, config.Version, mappedArch, mappedOS, extension, pluginConfig.Download.ReleaseVersion)

	// For Python, we want to use a simpler directory structure
	var installDir string
	if config.Name == "python" {
		installDir = path.Join(runtimesDir, "python")
	} else {
		installDir = path.Join(runtimesDir, fileName)
	}

	// Create RuntimeInfo with essential information
	info := &RuntimeInfo{
		Name:        config.Name,
		Version:     config.Version,
		InstallDir:  installDir,
		DownloadURL: downloadURL,
		FileName:    fileName,
		Extension:   extension,
		Binaries:    make(map[string]string),
	}

	// Process binary paths
	for _, binary := range pluginConfig.Binaries {
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

// GetRuntimeVersions returns a map of runtime names to their default versions
func GetRuntimeVersions() map[string]string {
	return GetPluginManager().GetRuntimeVersions()
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

	var config PluginConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("error parsing plugin.yaml: %w", err)
	}

	return &runtimePlugin{
		Config:     config,
		ConfigPath: pluginPath,
	}, nil
}
