package main

import (
	"codacy/cli-v2/cmd"
	"codacy/cli-v2/config"
	cfg "codacy/cli-v2/config-file"
	"fmt"
	"os"
)

func main() {
	fmt.Println("Running original CLI functionality...")
	// Original functionality
	config.Init()

	configErr := cfg.ReadConfigFile(config.Config.ProjectConfigFile())
	// whenever there is no configuration file, the only command allowed to run is the 'init'
	if configErr != nil && len(os.Args) > 1 && os.Args[1] != "init" {
		fmt.Println(configErr)
		fmt.Println("No configuration file was found, execute init command first.")
		return
	}

	cmd.Execute()
}
