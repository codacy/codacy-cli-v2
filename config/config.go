package config

import (
	"log"
	"os"
	"path/filepath"

	"codacy/cli-v2/plugins"
)

type ConfigType struct {
	homePath             string
	codacyDirectory      string
	runtimesDirectory    string
	toolsDirectory       string
	localCodacyDirectory string
	projectConfigFile    string

	runtimes map[string]*plugins.RuntimeInfo
	tools    map[string]*Runtime
}

func (c *ConfigType) HomePath() string {
	return c.homePath
}

func (c *ConfigType) CodacyDirectory() string {
	return c.codacyDirectory
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

// TODO do inheritance with tool
func (c *ConfigType) Tools() map[string]*Runtime {
	return c.tools
}

func (c *ConfigType) AddTool(t *Runtime) {
	c.tools[t.Name()] = t
}

func (c *ConfigType) initCodacyDirs() {
	c.codacyDirectory = filepath.Join(c.homePath, ".cache", "codacy")
	err := os.MkdirAll(c.codacyDirectory, 0777)
	if err != nil {
		log.Fatal(err)
	}

	c.runtimesDirectory = filepath.Join(c.codacyDirectory, "runtimes")
	err = os.MkdirAll(c.runtimesDirectory, 0777)
	if err != nil {
		log.Fatal(err)
	}

	c.toolsDirectory = filepath.Join(c.codacyDirectory, "tools")
	err = os.MkdirAll(c.toolsDirectory, 0777)
	if err != nil {
		log.Fatal(err)
	}

	c.localCodacyDirectory = ".codacy"
	err = os.MkdirAll(c.localCodacyDirectory, 0777)
	if err != nil {
		log.Fatal(err)
	}

	yamlPath := filepath.Join(c.localCodacyDirectory, "codacy.yaml")
	ymlPath := filepath.Join(c.localCodacyDirectory, "codacy.yml")

	if _, err := os.Stat(ymlPath); err == nil {
		c.projectConfigFile = ymlPath
	} else {
		c.projectConfigFile = yamlPath
	}
}

func Init() {
	homePath, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	Config.homePath = homePath

	Config.initCodacyDirs()

	Config.runtimes = make(map[string]*plugins.RuntimeInfo)
	Config.tools = make(map[string]*Runtime)
}

// Global singleton config-file
var Config = ConfigType{}
