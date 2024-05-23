package cmd

import (
	"codacy/cli-v2/config"
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"os"
)

func init() {
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Bootstraps project configuration",
	Long:  "Bootstraps project configuration, creates codacy configuration file",
	Run: func(cmd *cobra.Command, args []string) {
		err := configurationFileSetup()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Run install command to install dependencies.")
	},
}

func configurationFileSetup() error {
	configFile, err := os.Open(config.Config.ProjectConfigFile())
	defer configFile.Close()
	if err != nil {
		fmt.Println("Codacy cli configuration file was not found in", config.Config.LocalCodacyDirectory(), "- Creating file now.")
		err := createConfigurationFile()
		if err != nil {
			return err
		}
		return nil
	} else {
		fmt.Println("Codacy cli configuration file was found in ", config.Config.LocalCodacyDirectory())
	}

	return nil
}

func createConfigurationFile() error {

	configFile, err := os.Create(config.Config.ProjectConfigFile())
	defer configFile.Close()
	if err != nil {
		log.Fatal(err)
	}

	_, err = configFile.WriteString(configFileTemplate())
	if err != nil {
		log.Fatal(err)
	}

	return nil
}

func configFileTemplate() string {
	return `runtimes:
    - node@22.2.0
tools:
    - eslint@9.3.0

`
}
