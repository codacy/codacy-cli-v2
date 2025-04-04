package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"codacy/cli-v2/plugins"
)

type ConfigType struct {
	homePath             string
	globalCacheDirectory string
	runtimesDirectory    string
	toolsDirectory       string
	localCodacyDirectory string
	projectConfigFile    string
	cliConfigFile        string

	runtimes map[string]*plugins.RuntimeInfo
	tools    map[string]*plugins.ToolInfo
}

func (c *ConfigType) HomePath() string {
	return c.homePath
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
	}

	return nil
}

func (c *ConfigType) Tools() map[string]*plugins.ToolInfo {
	return c.tools
}

func (c *ConfigType) AddTools(configs []plugins.ToolConfig) error {
	// Process the tool configurations using the plugins.ProcessTools function
	toolInfoMap, err := plugins.ProcessTools(configs, c.toolsDirectory)
	if err != nil {
		return err
	}

	// Store the tool information in the config
	for name, info := range toolInfoMap {
		c.tools[name] = info
	}

	return nil
}

func (c *ConfigType) setupCodacyPaths() {
	c.globalCacheDirectory = filepath.Join(c.homePath, ".cache", "codacy")
	c.runtimesDirectory = filepath.Join(c.globalCacheDirectory, "runtimes")
	c.toolsDirectory = filepath.Join(c.globalCacheDirectory, "tools")
	c.localCodacyDirectory = ".codacy"

	c.projectConfigFile = filepath.Join(c.localCodacyDirectory, "codacy.yaml")
	c.cliConfigFile = filepath.Join(c.localCodacyDirectory, "cli-config.yaml")
}

func (c *ConfigType) CreateCodacyDirs() error {
	if err := os.MkdirAll(c.globalCacheDirectory, 0777); err != nil {
		return fmt.Errorf("failed to create codacy directory: %w", err)
	}

	if err := os.MkdirAll(c.runtimesDirectory, 0777); err != nil {
		return fmt.Errorf("failed to create runtimes directory: %w", err)
	}

	if err := os.MkdirAll(c.toolsDirectory, 0777); err != nil {
		return fmt.Errorf("failed to create tools directory: %w", err)
	}
	return nil
}

func (c *ConfigType) CreateLocalCodacyDir() error {
	if err := os.MkdirAll(c.localCodacyDirectory, 0777); err != nil {
		return fmt.Errorf("failed to create local codacy directory: %w", err)
	}
	return nil
}

func Init() {
	homePath, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	Config.homePath = homePath

	Config.setupCodacyPaths()
	Config.runtimes = make(map[string]*plugins.RuntimeInfo)
	Config.tools = make(map[string]*plugins.ToolInfo)
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
		_, err := os.Stat(binaryPath)
		if err == nil {
			return true
		}
	}

	return false
}

// Global singleton config-file
var Config = ConfigType{}
