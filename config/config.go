package config

import (
	"log"
	"os"
	"path/filepath"
)

type configType struct {
	homePath string
	codacyDirectory string
	runtimesDirectory string
	toolsDirectory string
}

var Config = configType{}

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

func (c *configType) initCodacyDirs() {
	c.codacyDirectory = filepath.Join(c.homePath, ".cache", "codacy")
	c.runtimesDirectory = filepath.Join(c.codacyDirectory, "runtimes")
	c.toolsDirectory = filepath.Join(c.codacyDirectory, "tools")

	err := os.MkdirAll(c.codacyDirectory, 0777)
	if err != nil {
		log.Fatal(err)
	}
	err = os.MkdirAll(c.runtimesDirectory, 0777)
	if err != nil {
		log.Fatal(err)
	}
	err = os.MkdirAll(c.toolsDirectory, 0777)
	if err != nil {
		log.Fatal(err)
	}
}

func Init() {
	homePath, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	Config.homePath = homePath

	Config.initCodacyDirs()
}