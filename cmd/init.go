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
	trivyVersion := "0.59.1" // Latest stable version

	for _, tool := range tools {
		if tool.Uuid == "f8b29663-2cb2-498d-b923-a10c6a8c05cd" {
			eslintVersion = tool.Version
		}
		if tool.Uuid == "2fd7fbe0-33f9-4ab3-ab73-e9b62404e2cb" {
			trivyVersion = tool.Version
		}
	}

	return fmt.Sprintf(`runtimes:
    - node@22.2.0
    - python@3.11.11
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

	// Create Trivy configuration after processing ESLint
	trivyApiConfiguration := extractTrivyConfiguration(apiToolConfigurations)
	if trivyApiConfiguration != nil {
		// Create trivy.yaml file based on API configuration
		err = createTrivyConfigFile(*trivyApiConfiguration)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Trivy configuration created based on Codacy settings")
	} else {
		// Create default trivy.yaml if no configuration from API
		err = createDefaultTrivyConfigFile()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Default Trivy configuration created")
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

// extractTrivyConfiguration extracts Trivy configuration from the Codacy API response
func extractTrivyConfiguration(toolConfigurations []CodacyToolConfiguration) *CodacyToolConfiguration {
	// Trivy internal codacy uuid
	const TrivyUUID = "2fd7fbe0-33f9-4ab3-ab73-e9b62404e2cb"

	for _, toolConfiguration := range toolConfigurations {
		if toolConfiguration.Uuid == TrivyUUID {
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

// createTrivyConfigFile creates a trivy.yaml configuration file based on the API configuration
func createTrivyConfigFile(config CodacyToolConfiguration) error {
	// Convert CodacyToolConfiguration to tools.ToolConfiguration
	trivyDomainConfiguration := convertAPIToolConfigurationForTrivy(config)

	// Use the shared CreateTrivyConfig function to generate the config content
	trivyConfigurationString := tools.CreateTrivyConfig(trivyDomainConfiguration)

	// Write to file
	return os.WriteFile("trivy.yaml", []byte(trivyConfigurationString), 0644)
}

// convertAPIToolConfigurationForTrivy converts API tool configuration to domain model for Trivy
func convertAPIToolConfigurationForTrivy(config CodacyToolConfiguration) tools.ToolConfiguration {
	var patterns []tools.PatternConfiguration

	// Only process if tool is enabled
	if config.IsEnabled {
		for _, pattern := range config.Patterns {
			var parameters []tools.PatternParameterConfiguration

			// By default patterns are enabled
			patternEnabled := true

			// Check if there's an explicit enabled parameter
			for _, param := range pattern.Parameters {
				if param.name == "enabled" && param.value == "false" {
					patternEnabled = false
				}
			}

			// Add enabled parameter
			parameters = append(parameters, tools.PatternParameterConfiguration{
				Name:  "enabled",
				Value: fmt.Sprintf("%t", patternEnabled),
			})

			patterns = append(
				patterns,
				tools.PatternConfiguration{
					PatternId:                pattern.InternalId,
					ParamenterConfigurations: parameters,
				},
			)
		}
	}

	return tools.ToolConfiguration{
		PatternsConfiguration: patterns,
	}
}

// createDefaultTrivyConfigFile creates a default trivy.yaml configuration file
func createDefaultTrivyConfigFile() error {
	// Use empty tool configuration to get default settings
	emptyConfig := tools.ToolConfiguration{}
	content := tools.CreateTrivyConfig(emptyConfig)

	// Write to file
	return os.WriteFile("trivy.yaml", []byte(content), 0644)
}
