package main

import (
	"codacy/cli-v2/cmd"
	"codacy/cli-v2/config"
	config_file "codacy/cli-v2/config-file"
	"fmt"
	"os"
)

func main() {
	// Initialize config global object
	config.Init()

	// This also setup the config global !
	configErr := config_file.ReadConfigFile(config.Config.ProjectConfigFile())

	// whenever there is no configuration file, the only command allowed to run is the 'init'
	if configErr != nil && len(os.Args) > 1 && os.Args[1] != "init" {
		fmt.Println("No configuration file was found, execute init command first.")
		return
	}

	cmd.Execute()
}
