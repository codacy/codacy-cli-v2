package main

import (
	"codacy/cli-v2/cmd"
	cfg "codacy/cli-v2/config"
	"log"
)

func main() {
	cfg.Init()

	runtimes, configErr := cfg.ReadConfigFile(cfg.Config.ProjectConfigFile())
	if configErr != nil {
		log.Fatal(configErr)
	}
	cfg.Config.SetRuntimes(runtimes)

	cmd.Execute()
}