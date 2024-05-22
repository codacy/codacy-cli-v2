package config

import (
	"log"
	"os"
	"path/filepath"
)


var Config = configType{}

type configType struct {
	homePath string
	codacyDirectory string
	runtimesDirectory string
	toolsDirectory string
	localCodacyDirectory string
	projectConfigFile string

	runtimes map[string]*Runtime
}

func (c *configType) HomePath() string {
	return c.homePath
}

func (c *configType) CodacyDirectory() string {
	return c.codacyDirectory
}

func (c *configType) RuntimesDirectory() string {
	return c.runtimesDirectory
}

func (c *configType) ToolsDirectory() string {
	return c.toolsDirectory
}

func (c *configType) LocalCodacyDirectory() string {
	return c.localCodacyDirectory
}

func (c *configType) ProjectConfigFile() string {
	return c.projectConfigFile
}

func (c *configType) Runtimes() map[string]*Runtime {
	return c.runtimes
}

func (c *configType) SetRuntimes(runtimes map[string]*Runtime) {
	c.runtimes = runtimes
}

func (c configType) initCodacyDirs() {
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
	c.projectConfigFile = filepath.Join(c.localCodacyDirectory, "codacy.yaml")
}

func Init() {
	homePath, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	Config.homePath = homePath

	Config.initCodacyDirs()
}