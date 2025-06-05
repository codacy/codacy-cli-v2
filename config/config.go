package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"codacy/cli-v2/plugins"
	"codacy/cli-v2/utils"

	"gopkg.in/yaml.v3" // Added import for YAML parsing
)

// CliConfigYaml defines the structure for parsing .codacy/cli-config.yaml
type CliConfigYaml struct {
	Mode string `yaml:"mode"`
}

type ConfigType struct {
	repositoryDirectory string

	globalCacheDirectory string
	runtimesDirectory    string
	toolsDirectory       string
	localCodacyDirectory string
	toolsConfigDirectory string
	projectConfigFile    string
	cliConfigFile        string

	runtimes map[string]*plugins.RuntimeInfo
	tools    map[string]*plugins.ToolInfo
}

func (c *ConfigType) RepositoryDirectory() string {
	return c.repositoryDirectory
}

func (c *ConfigType) CodacyDirectory() string {
	return c.globalCacheDirectory
}

func (c *ConfigType) RuntimesDirectory() string {
	return c.runtimesDirectory
}

func (c *ConfigType) ToolsDirectory() string {
	return c.toolsDirectory
}

func (c *ConfigType) LocalCodacyDirectory() string {
	return c.localCodacyDirectory
}

func (c *ConfigType) ToolsConfigsDirectory() string {
	return c.toolsConfigDirectory
}

func (c *ConfigType) ProjectConfigFile() string {
	return c.projectConfigFile
}

func (c *ConfigType) CliConfigFile() string {
	return c.cliConfigFile
}

func (c *ConfigType) Runtimes() map[string]*plugins.RuntimeInfo {
	return c.runtimes
}

// loadConfigOrInitializeEmpty reads the existing YAML config or returns an empty configuration if the file doesn't exist
func (c *ConfigType) loadConfigOrInitializeEmpty(codacyPath string) (map[string]interface{}, error) {
	content, err := os.ReadFile(codacyPath)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("error reading codacy.yaml: %w", err)
	}

	var config map[string]interface{}
	if len(content) > 0 {
		if err := yaml.Unmarshal(content, &config); err != nil {
			return nil, fmt.Errorf("error parsing codacy.yaml: %w", err)
		}
	} else {
		config = make(map[string]interface{})
	}

	return config, nil
}

// updateRuntimesList updates or adds a runtime entry in the runtimes list
func updateRuntimesList(runtimes []interface{}, name, version string) []interface{} {
	runtimeEntry := fmt.Sprintf("%s@%s", name, version)

	// Check if runtime already exists and update it
	for i, r := range runtimes {
		if runtime, ok := r.(string); ok {
			if strings.Split(runtime, "@")[0] == name {
				runtimes[i] = runtimeEntry
				return runtimes
			}
		}
	}

	// Add new runtime if not found
	return append(runtimes, runtimeEntry)
}

// updateToolsList updates the tools list in the configuration, avoiding duplicates
func updateToolsList(tools []interface{}, name, version string) []interface{} {
	toolEntry := fmt.Sprintf("%s@%s", name, version)

	// Check if tool already exists
	for i, tool := range tools {
		if toolStr, ok := tool.(string); ok && strings.HasPrefix(toolStr, name+"@") {
			// Replace existing tool
			tools[i] = toolEntry
			return tools
		}
	}

	// Add new tool
	return append(tools, toolEntry)
}

// writeConfig writes the config back to the YAML file
func (c *ConfigType) writeConfig(codacyPath string, config map[string]interface{}) error {
	yamlData, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("error marshaling codacy.yaml: %w", err)
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(codacyPath), utils.DefaultDirPerms); err != nil {
		return fmt.Errorf("error creating .codacy directory: %w", err)
	}

	if err := os.WriteFile(codacyPath, yamlData, utils.DefaultFilePerms); err != nil {
		return fmt.Errorf("error writing codacy.yaml: %w", err)
	}

	return nil
}

// addRuntimeToCodacyYaml adds or updates a runtime entry in codacy.yaml as a YAML list
func (c *ConfigType) addRuntimeToCodacyYaml(name string, version string) error {
	codacyPath := filepath.Join(c.localCodacyDirectory, "codacy.yaml")

	config, err := c.loadConfigOrInitializeEmpty(codacyPath)
	if err != nil {
		return err
	}

	// Get or create runtimes list
	runtimes, ok := config["runtimes"].([]interface{})
	if !ok {
		runtimes = make([]interface{}, 0)
	}

	// Update runtimes list
	runtimes = updateRuntimesList(runtimes, name, version)
	config["runtimes"] = runtimes

	return c.writeConfig(codacyPath, config)
}

// addToolToCodacyYaml adds a tool to the codacy.yaml file
func (c *ConfigType) addToolToCodacyYaml(name string, version string) error {
	codacyPath := filepath.Join(c.localCodacyDirectory, "codacy.yaml")

	config, err := c.loadConfigOrInitializeEmpty(codacyPath)
	if err != nil {
		return err
	}

	// Get or create tools list
	tools, ok := config["tools"].([]interface{})
	if !ok {
		tools = make([]interface{}, 0)
	}

	// Update tools list
	tools = updateToolsList(tools, name, version)
	config["tools"] = tools

	return c.writeConfig(codacyPath, config)
}

func (c *ConfigType) AddRuntimes(configs []plugins.RuntimeConfig) error {
	// Process the runtime configurations using the plugins.ProcessRuntimes function
	runtimeInfoMap, err := plugins.ProcessRuntimes(configs, c.runtimesDirectory)
	if err != nil {
		return err
	}

	// Store the runtime information in the config
	for name, info := range runtimeInfoMap {
		c.runtimes[name] = info

		// Update codacy.yaml with the new runtime
		if err := c.addRuntimeToCodacyYaml(name, info.Version); err != nil {
			return fmt.Errorf("failed to update codacy.yaml with runtime %s: %w", name, err)
		}
	}

	return nil
}

func (c *ConfigType) Tools() map[string]*plugins.ToolInfo {
	return c.tools
}

func (c *ConfigType) AddTools(configs []plugins.ToolConfig) error {
	// Get the plugin manager to access tool configurations
	pluginManager := plugins.GetPluginManager()

	// Ensure all required runtimes are present before processing tools
	for _, toolConfig := range configs {
		// Get the tool's plugin configuration to access runtime info
		pluginConfig, err := pluginManager.GetToolConfig(toolConfig.Name)
		if err != nil {
			return fmt.Errorf("failed to get plugin config for tool %s: %w", toolConfig.Name, err)
		}

		if pluginConfig.Runtime != "" {
			// Special handling for dartanalyzer - check for existing dart or flutter runtimes
			if toolConfig.Name == "dartanalyzer" {
				// Check if either dart or flutter runtime is already available
				if runtimeInfo := c.runtimes["flutter"]; runtimeInfo != nil {
					// Flutter runtime exists, use it
					continue
				}
				if runtimeInfo := c.runtimes["dart"]; runtimeInfo != nil {
					// Dart runtime exists, use it
					continue
				}
				// Neither runtime exists, proceed with installing dart runtime
				pluginConfig.Runtime = "dart"
			}

			runtimeInfo := c.runtimes[pluginConfig.Runtime]
			if runtimeInfo == nil {
				// Get the default version for the runtime
				defaultVersions := plugins.GetRuntimeVersions()
				version, ok := defaultVersions[pluginConfig.Runtime]
				if !ok {
					return fmt.Errorf("no default version found for runtime %s", pluginConfig.Runtime)
				}

				// Add the runtime to the config
				fmt.Println("Adding missing runtime to codacy.yaml", pluginConfig.Runtime, version)
				if err := c.AddRuntimes([]plugins.RuntimeConfig{{Name: pluginConfig.Runtime, Version: version}}); err != nil {
					return fmt.Errorf("failed to add runtime %s: %w", pluginConfig.Runtime, err)
				}

				// Fetch the runtimeInfo again
				runtimeInfo = c.runtimes[pluginConfig.Runtime]
				if runtimeInfo == nil {
					return fmt.Errorf("runtime %s still missing after adding to config", pluginConfig.Runtime)
				}
			}
		}
	}

	// Now safe to process tools
	toolInfoMap, err := plugins.ProcessTools(configs, c.toolsDirectory, c.runtimes)
	if err != nil {
		return err
	}

	// Store the tool information in the config and update codacy.yaml
	for name, info := range toolInfoMap {
		c.tools[name] = info

		// Update codacy.yaml with the new tool
		if err := c.addToolToCodacyYaml(name, info.Version); err != nil {
			return fmt.Errorf("failed to update codacy.yaml with tool %s: %w", name, err)
		}
	}

	return nil
}

// AddToolWithDefaultVersion adds a tool with its default version to the configuration
func (c *ConfigType) AddToolWithDefaultVersion(toolName string) error {
	// Get the default version for the tool from plugins
	defaultVersions := plugins.GetToolVersions()
	version, ok := defaultVersions[toolName]
	if !ok {
		return fmt.Errorf("no default version found for tool %s", toolName)
	}

	fmt.Printf("Adding tool %s with default version %s to codacy.yaml\n", toolName, version)

	// Create tool config
	toolConfig := plugins.ToolConfig{
		Name:    toolName,
		Version: version,
	}

	// Add the tool using the existing AddTools function
	return c.AddTools([]plugins.ToolConfig{toolConfig})
}

func (c *ConfigType) ToolsConfigDirectory() string {
	return c.toolsConfigDirectory
}

func NewConfigType(repositoryDirectory string, repositoryCache string, globalCache string) *ConfigType {
	c := &ConfigType{}

	c.repositoryDirectory = repositoryDirectory
	c.localCodacyDirectory = repositoryCache
	c.globalCacheDirectory = globalCache

	c.runtimesDirectory = filepath.Join(c.globalCacheDirectory, "runtimes")
	c.toolsDirectory = filepath.Join(c.globalCacheDirectory, "tools")
	c.toolsConfigDirectory = filepath.Join(c.localCodacyDirectory, "tools-configs")

	// If codacy.yml exists, we should rely on it
	ymlPath := filepath.Join(c.localCodacyDirectory, "codacy.yml")
	if _, err := os.Stat(ymlPath); err == nil {
		c.projectConfigFile = ymlPath
	} else {
		// Otherwise default to codacy.yaml
		c.projectConfigFile = filepath.Join(c.localCodacyDirectory, "codacy.yaml")
	}

	c.cliConfigFile = filepath.Join(c.localCodacyDirectory, "cli-config.yaml")

	c.runtimes = make(map[string]*plugins.RuntimeInfo)
	c.tools = make(map[string]*plugins.ToolInfo)
	return c
}

// TODO: Consider not having a global config and instead pass the config object around
func setupGlobalConfig(repositoryDirectory string, repositoryCache string, globalCache string) {
	newConfig := NewConfigType(repositoryDirectory, repositoryCache, globalCache)

	Config.repositoryDirectory = newConfig.repositoryDirectory
	Config.localCodacyDirectory = newConfig.localCodacyDirectory
	Config.globalCacheDirectory = newConfig.globalCacheDirectory

	Config.runtimesDirectory = newConfig.runtimesDirectory
	Config.toolsDirectory = newConfig.toolsDirectory
	Config.toolsConfigDirectory = newConfig.toolsConfigDirectory

	Config.projectConfigFile = newConfig.projectConfigFile
	Config.cliConfigFile = newConfig.cliConfigFile

	Config.runtimes = newConfig.runtimes
	Config.tools = newConfig.tools
}

func (c *ConfigType) CreateCodacyDirs() error {
	if err := os.MkdirAll(c.globalCacheDirectory, utils.DefaultDirPerms); err != nil {
		return fmt.Errorf("failed to create codacy directory: %w", err)
	}

	if err := os.MkdirAll(c.runtimesDirectory, utils.DefaultDirPerms); err != nil {
		return fmt.Errorf("failed to create runtimes directory: %w", err)
	}

	if err := os.MkdirAll(c.toolsDirectory, utils.DefaultDirPerms); err != nil {
		return fmt.Errorf("failed to create tools directory: %w", err)
	}
	return nil
}

func (c *ConfigType) CreateLocalCodacyDir() error {
	if err := os.MkdirAll(c.localCodacyDirectory, utils.DefaultDirPerms); err != nil {
		return fmt.Errorf("failed to create local codacy directory: %w", err)
	}
	return nil
}

func Init() {
	homePath, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	// Repository directory is the current working directory
	repositoryDirectory := ""
	repositoryCache := ".codacy"
	globalCache := filepath.Join(homePath, ".cache", "codacy")

	setupGlobalConfig(repositoryDirectory, repositoryCache, globalCache)
}

// IsRuntimeInstalled checks if a runtime is already installed
func (c *ConfigType) IsRuntimeInstalled(name string, runtime *plugins.RuntimeInfo) bool {
	// If there are no binaries, check the install directory
	if len(runtime.Binaries) == 0 {
		_, err := os.Stat(runtime.InstallDir)
		return err == nil
	}

	// Check if at least one binary exists
	for _, binaryPath := range runtime.Binaries {
		_, err := os.Stat(binaryPath)
		if err == nil {
			return true
		}
	}

	return false
}

// IsToolInstalled checks if a tool is already installed
func (c *ConfigType) IsToolInstalled(name string, tool *plugins.ToolInfo) bool {
	// If there are no binaries, check the install directory
	if len(tool.Binaries) == 0 {
		_, err := os.Stat(tool.InstallDir)
		return err == nil
	}

	// Check if at least one binary exists
	for _, binaryPath := range tool.Binaries {
		// If the path is relative, join it with the installation directory
		if !filepath.IsAbs(binaryPath) {
			binaryPath = filepath.Join(tool.InstallDir, binaryPath)
		}

		// Try both with and without .exe
		_, err := os.Stat(binaryPath)
		if err == nil {
			return true
		}

		_, err = os.Stat(binaryPath + ".exe")
		if err == nil {
			return true
		}
	}

	return false
}

// Global singleton config-file
var Config = ConfigType{}

// GetCliMode reads and parses the .codacy/cli-config.yaml file to determine the CLI's operational mode.
// It returns "local" by default and an error if the file doesn't exist or an error occurs during parsing.
func (c *ConfigType) GetCliMode() (string, error) {
	cliConfigFilePath := c.CliConfigFile()
	currentCliMode := "local" // Default to local

	content, readErr := os.ReadFile(cliConfigFilePath)
	if readErr != nil {
		if os.IsNotExist(readErr) {
			// File does not exist. Return default mode and the error so the caller can warn.
			return currentCliMode, readErr
		}
		// Some other error occurred during reading the file.
		return currentCliMode, fmt.Errorf("failed to read %s: %w", cliConfigFilePath, readErr)
	}

	// If ReadFile was successful, the file exists. Now parse it.
	var parsedCliConfig CliConfigYaml
	if yamlErr := yaml.Unmarshal(content, &parsedCliConfig); yamlErr != nil {
		return currentCliMode, fmt.Errorf("failed to parse %s: %w", cliConfigFilePath, yamlErr)
	}

	if parsedCliConfig.Mode == "remote" || parsedCliConfig.Mode == "local" {
		currentCliMode = parsedCliConfig.Mode
	} else {
		// Invalid mode value in the config file.
		return "local", fmt.Errorf("invalid mode value \"%s\" in %s", parsedCliConfig.Mode, cliConfigFilePath)
	}

	return currentCliMode, nil
}
