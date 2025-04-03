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

	// Show help if any argument contains help
	for _, arg := range os.Args {
		if arg == "--help" || arg == "-h" || arg == "help" {
			cmd.Execute()
			return
		}
	}

	// Check if command is init
	if len(os.Args) > 1 && os.Args[1] == "init" {
		cmd.Execute()
		return
	}

	// All other commands require a configuration file
	if configErr != nil && len(os.Args) > 1 {
		fmt.Println("No configuration file was found, execute init command first.")
		return
	}

	cmd.Execute()
}
