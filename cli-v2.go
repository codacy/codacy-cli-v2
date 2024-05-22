package main

import (
	"codacy/cli-v2/cmd"
	"codacy/cli-v2/config"
	cfg "codacy/cli-v2/config-file"
	"log"
)

func main() {
	config.Init()

	configErr := cfg.ReadConfigFile(config.Config.ProjectConfigFile())
	if configErr != nil {
		log.Fatal(configErr)
	}

	cmd.Execute()
}