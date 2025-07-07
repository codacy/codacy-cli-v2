package main

import (
	"codacy/cli-v2/cmd"
	"codacy/cli-v2/config"
	config_file "codacy/cli-v2/config-file"
	"codacy/cli-v2/utils/logger"
	"codacy/cli-v2/version"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

func main() {

	fmt.Printf("TESTING CLI BUILD")
	// Initialize config global object
	config.Init()

	// Initialize logger with the logs directory
	logsDir := filepath.Join(config.Config.LocalCodacyDirectory(), "logs")
	if err := logger.Initialize(logsDir); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
	}

	// Log startup message and version
	logger.Info("Starting Codacy CLI", logrus.Fields{
		"version": version.GetVersion(),
	})

	// This also setup the config global !
	configErr := config_file.ReadConfigFile(config.Config.ProjectConfigFile())

	// Show help if any argument contains help
	for _, arg := range os.Args {
		if arg == "--help" || arg == "-h" || arg == "help" {
			cmd.Execute()
			return
		}
	}

	// Check if command is init/update/version/help - these don't require configuration
	if len(os.Args) > 1 {
		cmdName := os.Args[1]
		if cmdName == "init" || cmdName == "update" || cmdName == "version" || cmdName == "help" {
			cmd.Execute()
			return
		}
	}

	// All other commands require a configuration file
	if configErr != nil && len(os.Args) > 1 {
		if os.IsNotExist(configErr) {
			message := "No configuration file was found, execute init command first."
			logger.Info(message)
			fmt.Println(message)
		} else {
			logger.Error("Configuration error", logrus.Fields{
				"error": configErr.Error(),
			})
			fmt.Printf("Failed to parse configuration file: %v\n", configErr)
			fmt.Println("Please check the file format and try again, or run init command to create a new configuration.")
		}
		return
	}

	cmd.Execute()
}
