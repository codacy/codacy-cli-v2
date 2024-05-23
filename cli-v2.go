package main

import (
	"codacy/cli-v2/cmd"
	"codacy/cli-v2/config"
	cfg "codacy/cli-v2/config-file"
	"fmt"
)

func main() {

	config.Init()
	configErr := cfg.ReadConfigFile(config.Config.ProjectConfigFile())
	if configErr != nil {
		fmt.Println("No configuration file was found, execute init command first.")
	}

	cmd.Execute()
}
