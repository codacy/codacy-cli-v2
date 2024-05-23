package main

import (
	"codacy/cli-v2/cmd"
	"codacy/cli-v2/config"
	cfg "codacy/cli-v2/config-file"
	"fmt"
	"os"
)

func main() {
	config.Init()

	configErr := cfg.ReadConfigFile(config.Config.ProjectConfigFile())
	// whenever there is no configuration file, the only command allowed to run is the 'init'
	if configErr != nil && os.Args[1] != "init" {
		fmt.Println("No configuration file was found, execute init command first.")
		return
	}

	cmd.Execute()
}
