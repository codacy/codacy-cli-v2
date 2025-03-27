package cmd

import (
	"codacy/cli-v2/config"
	"codacy/cli-v2/tools"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"
)

const CodacyApiBase = "https://app.codacy.com"

var codacyRepositoryToken string

func init() {
	initCmd.Flags().StringVar(&codacyRepositoryToken, "repository-token", "", "optional codacy repository token, if defined configurations will be fetched from codacy")
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Bootstraps project configuration",
	Long:  "Bootstraps project configuration, creates codacy configuration file",
	Run: func(cmd *cobra.Command, args []string) {
		if len(codacyRepositoryToken) == 0 {
			fmt.Println("No project token was specified, skipping fetch configurations ")
			noTools := []tools.Tool{}
			err := createConfigurationFile(noTools)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			apiTools, err := tools.GetTools()
			if err != nil {
				log.Fatal(err)
			}
			err = createConfigurationFile(apiTools)
			if err != nil {
				log.Fatal(err)
			}
			err = buildRepositoryConfigurationFiles(codacyRepositoryToken)
			if err != nil {
				log.Fatal(err)
			}
		}
		fmt.Println("Run install command to install dependencies.")
	},
}

func createConfigurationFile(tools []tools.Tool) error {

	configFile, err := os.Create(config.Config.ProjectConfigFile())
	defer configFile.Close()
	if err != nil {
		log.Fatal(err)
	}

	_, err = configFile.WriteString(configFileTemplate(tools))
	if err != nil {
		log.Fatal(err)
	}

	return nil
}

func configFileTemplate(tools []tools.Tool) string {

	// Default versions
	eslintVersion := "9.3.0"
	trivyVersion := "0.50.0" // Use the latest stable version

	for _, tool := range tools {
		if tool.Uuid == "f8b29663-2cb2-498d-b923-a10c6a8c05cd" {
			eslintVersion = tool.Version
		}
		// If Codacy API provides UUID for Trivy, you would check it here
	}

	return fmt.Sprintf(`runtimes:
    - node@22.2.0
tools:
    - eslint@%s
    - trivy@%s
`, eslintVersion, trivyVersion)
}

func buildRepositoryConfigurationFiles(token string) error {

	fmt.Println("Building project configuration files ...")
	fmt.Println("Fetching project configuration from codacy ...")

	// API call to fetch settings
	url := CodacyApiBase + "/2.0/project/analysis/configuration"

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Create a new GET request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}

	// Set the headers
	req.Header.Set("project-token", token)

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return errors.New("failed to get repository's configuration from Codacy API")
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}

	var objmap map[string]json.RawMessage
	_ = json.Unmarshal(body, &objmap)

	var apiToolConfigurations []CodacyToolConfiguration
	err = json.Unmarshal(objmap["toolConfiguration"], &apiToolConfigurations)

	eslintApiConfiguration := extractESLintConfiguration(apiToolConfigurations)

	eslintDomainConfiguration := convertAPIToolConfigurationToDomain(*eslintApiConfiguration)

	eslintConfigurationString := tools.CreateEslintConfig(eslintDomainConfiguration)

	eslintConfigFile, err := os.Create("eslint.config.mjs")
	defer eslintConfigFile.Close()
	if err != nil {
		log.Fatal(err)
	}

	_, err = eslintConfigFile.WriteString(eslintConfigurationString)
	if err != nil {
		log.Fatal(err)

	}

	return nil
}

func convertAPIToolConfigurationToDomain(config CodacyToolConfiguration) tools.ToolConfiguration {
	var patterns []tools.PatternConfiguration

	for _, pattern := range config.Patterns {
		var parameters []tools.PatternParameterConfiguration

		for _, parameter := range pattern.Parameters {
			parameters = append(parameters, tools.PatternParameterConfiguration{
				Name:  parameter.name,
				Value: parameter.value,
			})
		}

		patterns = append(
			patterns,
			tools.PatternConfiguration{
				PatternId:                pattern.InternalId,
				ParamenterConfigurations: parameters,
			},
		)
	}

	return tools.ToolConfiguration{
		PatternsConfiguration: patterns,
	}
}

func extractESLintConfiguration(toolConfigurations []CodacyToolConfiguration) *CodacyToolConfiguration {

	//ESLInt internal codacy uuid, to filter ot not ESLint tools
	//"f8b29663-2cb2-498d-b923-a10c6a8c05cd"

	for _, toolConfiguration := range toolConfigurations {
		if toolConfiguration.Uuid == "f8b29663-2cb2-498d-b923-a10c6a8c05cd" {
			return &toolConfiguration
		}
	}

	return nil
}

type CodacyToolConfiguration struct {
	Uuid      string                 `json:"uuid"`
	IsEnabled bool                   `json:"isEnabled"`
	Patterns  []PatternConfiguration `json:"patterns"`
}

type PatternConfiguration struct {
	InternalId string                   `json:"internalId"`
	Parameters []ParameterConfiguration `json:"parameters"`
}

type ParameterConfiguration struct {
	name  string `json:"name"`
	value string `json:"value"`
}
