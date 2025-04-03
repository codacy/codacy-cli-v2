package cmd

import (
	"codacy/cli-v2/config"
	"codacy/cli-v2/domain"
	"codacy/cli-v2/tools"
	"codacy/cli-v2/tools/pylint"
	"codacy/cli-v2/utils"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

const CodacyApiBase = "https://app.codacy.com"

type InitFlags struct {
	apiToken     string
	provider     string
	organization string
	repository   string
}

var initFlags InitFlags

func init() {
	initCmd.Flags().StringVar(&initFlags.apiToken, "api-token", "", "optional codacy api token, if defined configurations will be fetched from codacy")
	initCmd.Flags().StringVar(&initFlags.provider, "provider", "", "provider (gh/bb/gl) to fetch configurations from codacy, required when api-token is provided")
	initCmd.Flags().StringVar(&initFlags.organization, "organization", "", "remote organization name to fetch configurations from codacy, required when api-token is provided")
	initCmd.Flags().StringVar(&initFlags.repository, "repository", "", "remote repository name to fetch configurations from codacy, required when api-token is provided")
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Bootstraps project configuration",
	Long:  "Bootstraps project configuration, creates codacy configuration file",
	Run: func(cmd *cobra.Command, args []string) {
		// Create local codacy directory first
		if err := config.Config.CreateLocalCodacyDir(); err != nil {
			log.Fatalf("Failed to create local codacy directory: %v", err)
		}

		// Create tools-configs directory
		toolsConfigDir := config.Config.ToolsConfigDirectory()
		if err := os.MkdirAll(toolsConfigDir, 0777); err != nil {
			log.Fatalf("Failed to create tools-configs directory: %v", err)
		}

		cliLocalMode := len(initFlags.apiToken) == 0

		if cliLocalMode {
			fmt.Println()
			fmt.Println("ℹ️  No project token was specified, skipping fetch configurations")
			noTools := []tools.Tool{}
			err := createConfigurationFiles(noTools, cliLocalMode)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			err := buildRepositoryConfigurationFiles(initFlags.apiToken)
			if err != nil {
				log.Fatal(err)
			}
			createGitIgnoreFile()
		}
		fmt.Println()
		fmt.Println("✅ Successfully initialized Codacy configuration!")
		fmt.Println()
		fmt.Println("🔧 Next steps:")
		fmt.Println("  1. Run 'codacy-cli install' to install all dependencies")
		fmt.Println("  2. Run 'codacy-cli analyze' to start analyzing your code")
		fmt.Println()
	},
}

func createGitIgnoreFile() error {
	gitIgnorePath := filepath.Join(config.Config.LocalCodacyDirectory(), ".gitignore")
	gitIgnoreFile, err := os.Create(gitIgnorePath)
	if err != nil {
		return fmt.Errorf("failed to create .gitignore file: %w", err)
	}
	defer gitIgnoreFile.Close()

	content := `# Codacy CLI 
tools-configs/
.gitignore
cli-config.yaml
`
	if _, err := gitIgnoreFile.WriteString(content); err != nil {
		return fmt.Errorf("failed to write to .gitignore file: %w", err)
	}

	return nil
}

func createConfigurationFiles(tools []tools.Tool, cliLocalMode bool) error {
	configFile, err := os.Create(config.Config.ProjectConfigFile())
	if err != nil {
		return fmt.Errorf("failed to create project config file: %w", err)
	}
	defer configFile.Close()

	configContent := configFileTemplate(tools)
	_, err = configFile.WriteString(configContent)
	if err != nil {
		return fmt.Errorf("failed to write project config file: %w", err)
	}

	cliConfigFile, err := os.Create(config.Config.CliConfigFile())
	if err != nil {
		return fmt.Errorf("failed to create CLI config file: %w", err)
	}
	defer cliConfigFile.Close()

	cliConfigContent := cliConfigFileTemplate(cliLocalMode)
	_, err = cliConfigFile.WriteString(cliConfigContent)
	if err != nil {
		return fmt.Errorf("failed to write CLI config file: %w", err)
	}

	return nil
}

func configFileTemplate(tools []tools.Tool) string {

	// Default versions
	eslintVersion := "9.3.0"
	trivyVersion := "0.59.1" // Latest stable version
	pylintVersion := "3.3.6"
	pmdVersion := "6.55.0"

	for _, tool := range tools {
		switch tool.Uuid {
		case ESLint:
			eslintVersion = tool.Version
		case Trivy:
			trivyVersion = tool.Version
		case PyLint:
			pylintVersion = tool.Version
		case PMD:
			pmdVersion = tool.Version
		}
	}

	return fmt.Sprintf(`runtimes:
    - node@22.2.0
    - python@3.11.11
tools:
    - eslint@%s
    - trivy@%s
    - pylint@%s
    - pmd@%s
`, eslintVersion, trivyVersion, pylintVersion, pmdVersion)
}

func cliConfigFileTemplate(cliLocalMode bool) string {
	var cliModeString string

	if cliLocalMode {
		cliModeString = "local"
	} else {
		cliModeString = "remote"
	}

	return fmt.Sprintf(`mode: %s`, cliModeString)
}

func buildRepositoryConfigurationFiles(token string) error {
	fmt.Println("Fetching repository configuration from codacy ...")

	toolsConfigDir := config.Config.ToolsConfigDirectory()

	// Create tools-configs directory if it doesn't exist
	if err := os.MkdirAll(toolsConfigDir, utils.DefaultDirPerms); err != nil {
		return fmt.Errorf("failed to create tools-configs directory: %w", err)
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	apiTools, err := tools.GetRepositoryTools(CodacyApiBase, token, initFlags.provider, initFlags.organization, initFlags.repository)
	if err != nil {
		return err
	}

	err = createConfigurationFiles(apiTools, true)
	if err != nil {
		log.Fatal(err)
	}

	for _, tool := range apiTools {
		url := fmt.Sprintf("%s/api/v3/analysis/organizations/%s/%s/repositories/%s/tools/%s/patterns?enabled=true",
			CodacyApiBase,
			initFlags.provider,
			initFlags.organization,
			initFlags.repository,
			tool.Uuid)

		// Create a new GET request
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			fmt.Println("Error:", err)
			return err
		}

		// Set the headers
		req.Header.Set("api-token", token)

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
		err = json.Unmarshal(body, &objmap)

		if err != nil {
			fmt.Println("Error unmarshaling response:", err)
			return err
		}

		var apiToolConfigurations []domain.PatternConfiguration
		err = json.Unmarshal(objmap["data"], &apiToolConfigurations)

		if err != nil {
			fmt.Println("Error unmarshaling tool configurations:", err)
			return err
		}

		createToolFileConfigurations(tool, apiToolConfigurations)
	}

	return nil
}

// map tool uuid to tool name
func createToolFileConfigurations(tool tools.Tool, patternConfiguration []domain.PatternConfiguration) error {
	toolsConfigDir := config.Config.ToolsConfigDirectory()
	switch tool.Uuid {
	case ESLint:
		if len(patternConfiguration) > 0 {
			eslintConfigurationString := tools.CreateEslintConfig(patternConfiguration)

			eslintConfigFile, err := os.Create(filepath.Join(toolsConfigDir, "eslint.config.mjs"))
			if err != nil {
				return fmt.Errorf("failed to create eslint config file: %v", err)
			}
			defer eslintConfigFile.Close()

			_, err = eslintConfigFile.WriteString(eslintConfigurationString)
			if err != nil {
				return fmt.Errorf("failed to write eslint config: %v", err)
			}
			fmt.Println("ESLint configuration created based on Codacy settings")
		} else {
			err := createDefaultEslintConfigFile(toolsConfigDir)
			if err != nil {
				return fmt.Errorf("failed to create default ESLint config: %v", err)
			}
			fmt.Println("Default ESLint configuration created")
		}
	case Trivy:
		if len(patternConfiguration) > 0 {
			err := createTrivyConfigFile(patternConfiguration, toolsConfigDir)
			if err != nil {
				return fmt.Errorf("failed to create Trivy config: %v", err)
			}
		} else {
			err := createDefaultTrivyConfigFile(toolsConfigDir)
			if err != nil {
				return fmt.Errorf("failed to create default Trivy config: %v", err)
			}
		}
		fmt.Println("Trivy configuration created based on Codacy settings")
	case PMD:
		if len(patternConfiguration) > 0 {
			err := createPMDConfigFile(patternConfiguration, toolsConfigDir)
			if err != nil {
				return fmt.Errorf("failed to create PMD config: %v", err)
			}
		} else {
			err := createDefaultPMDConfigFile(toolsConfigDir)
			if err != nil {
				return fmt.Errorf("failed to create default PMD config: %v", err)
			}
		}
		fmt.Println("PMD configuration created based on Codacy settings")
	}
	return nil
}

func createPMDConfigFile(config []domain.PatternConfiguration, toolsConfigDir string) error {
	pmdConfigurationString := tools.CreatePmdConfig(config)
	return os.WriteFile(filepath.Join(toolsConfigDir, "pmd-ruleset.xml"), []byte(pmdConfigurationString), utils.DefaultFilePerms)
}

func createDefaultPMDConfigFile(toolsConfigDir string) error {
	content := tools.CreatePmdConfig([]domain.PatternConfiguration{})
	return os.WriteFile(filepath.Join(toolsConfigDir, "pmd-ruleset.xml"), []byte(content), utils.DefaultFilePerms)
}

func createDefaultPylintConfigFile() error {
	pylintConfigurationString := pylint.GeneratePylintRCDefault()
	return os.WriteFile(".pylintrc", []byte(pylintConfigurationString), 0644)
}

// createTrivyConfigFile creates a trivy.yaml configuration file based on the API configuration
func createTrivyConfigFile(config []domain.PatternConfiguration, toolsConfigDir string) error {

	trivyConfigurationString := tools.CreateTrivyConfig(config)

	// Write to file
	return os.WriteFile(filepath.Join(toolsConfigDir, "trivy.yaml"), []byte(trivyConfigurationString), utils.DefaultFilePerms)
}

// createDefaultTrivyConfigFile creates a default trivy.yaml configuration file
func createDefaultTrivyConfigFile(toolsConfigDir string) error {
	// Use empty tool configuration to get default settings
	emptyConfig := []domain.PatternConfiguration{}
	content := tools.CreateTrivyConfig(emptyConfig)

	// Write to file
	return os.WriteFile(filepath.Join(toolsConfigDir, "trivy.yaml"), []byte(content), utils.DefaultFilePerms)
}

// createDefaultEslintConfigFile creates a default eslint.config.mjs configuration file
func createDefaultEslintConfigFile(toolsConfigDir string) error {
	// Use empty tool configuration to get default settings
	emptyConfig := []domain.PatternConfiguration{}
	content := tools.CreateEslintConfig(emptyConfig)

	// Write to file
	return os.WriteFile(filepath.Join(toolsConfigDir, "eslint.config.mjs"), []byte(content), utils.DefaultFilePerms)
}

const (
	ESLint string = "f8b29663-2cb2-498d-b923-a10c6a8c05cd"
	Trivy  string = "2fd7fbe0-33f9-4ab3-ab73-e9b62404e2cb"
	PMD    string = "9ed24812-b6ee-4a58-9004-0ed183c45b8f"
	PyLint string = "31677b6d-4ae0-4f56-8041-606a8d7a8e61"
)
