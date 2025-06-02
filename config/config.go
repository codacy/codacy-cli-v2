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
		if err := addRuntimeToCodacyYaml(name, info.Version); err != nil {
			return fmt.Errorf("failed to update codacy.yaml with runtime %s: %w", name, err)
		}
	}

	return nil
}

// addRuntimeToCodacyYaml adds or updates a runtime entry in .codacy/codacy.yaml as a YAML list
func addRuntimeToCodacyYaml(name string, version string) error {
	codacyPath := ".codacy/codacy.yaml"

	type CodacyConfig struct {
		Runtimes []string `yaml:"runtimes"`
		Tools    []string `yaml:"tools"`
	}

	// Read existing config
	var config CodacyConfig
	if data, err := os.ReadFile(codacyPath); err == nil {
		if err := yaml.Unmarshal(data, &config); err != nil {
			return fmt.Errorf("failed to parse %s: %w", codacyPath, err)
		}
	}

	// Prepare the new runtime string
	runtimeEntry := name + "@" + version
	found := false
	for i, entry := range config.Runtimes {
		if strings.HasPrefix(entry, name+"@") {
			config.Runtimes[i] = runtimeEntry
			found = true
			break
		}
	}
	if !found {
		config.Runtimes = append(config.Runtimes, runtimeEntry)
	}

	// Create .codacy directory if it doesn't exist
	if err := os.MkdirAll(".codacy", 0755); err != nil {
		return fmt.Errorf("failed to create .codacy directory: %w", err)
	}

	// Write back to .codacy/codacy.yaml
	yamlData, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config to YAML: %w", err)
	}
	if err := os.WriteFile(codacyPath, yamlData, 0644); err != nil {
		return fmt.Errorf("failed to write config to %s: %w", codacyPath, err)
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
			runtimeInfo := c.runtimes[pluginConfig.Runtime]
			if runtimeInfo == nil {
				// Get the default version for the runtime
				defaultVersions := plugins.GetRuntimeVersions()
				version, ok := defaultVersions[pluginConfig.Runtime]
				if !ok {
					return fmt.Errorf("no default version found for runtime %s", pluginConfig.Runtime)
				}

				// Add the runtime to the config
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

	// Store the tool information in the config
	for name, info := range toolInfoMap {
		c.tools[name] = info
	}

	return nil
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
