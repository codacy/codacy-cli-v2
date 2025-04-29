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
	"strings"
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
logs/
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
	// Maps to track which tools are enabled
	toolsMap := make(map[string]bool)
	toolVersions := make(map[string]string)

	// Track needed runtimes
	needsNode := false
	needsPython := false
	needsDart := false

	// Default versions
	defaultVersions := map[string]string{
		ESLint:       "9.3.0",
		Trivy:        "0.59.1",
		PyLint:       "3.3.6",
		PMD:          "6.55.0",
		DartAnalyzer: "3.7.2",
		Semgrep:      "1.78.0",
		Lizard:       "1.17.19",
	}

	// Build map of enabled tools with their versions
	for _, tool := range tools {
		toolsMap[tool.Uuid] = true
		if tool.Version != "" {
			toolVersions[tool.Uuid] = tool.Version
		} else {
			toolVersions[tool.Uuid] = defaultVersions[tool.Uuid]
		}

		// Check if tool needs a runtime
		if tool.Uuid == ESLint {
			needsNode = true
		} else if tool.Uuid == PyLint || tool.Uuid == Lizard {
			needsPython = true
		} else if tool.Uuid == DartAnalyzer {
			needsDart = true
		}
	}

	// Start building the YAML content
	var sb strings.Builder
	sb.WriteString("runtimes:\n")

	// Only include runtimes needed by the enabled tools
	if len(tools) > 0 {
		if needsNode {
			sb.WriteString("    - node@22.2.0\n")
		}
		if needsPython {
			sb.WriteString("    - python@3.11.11\n")
		}
		if needsDart {
			sb.WriteString("    - dart@3.7.2\n")
		}
	} else {
		// In local mode with no tools specified, include all runtimes
		sb.WriteString("    - node@22.2.0\n")
		sb.WriteString("    - python@3.11.11\n")
		sb.WriteString("    - dart@3.7.2\n")
	}

	sb.WriteString("tools:\n")

	// If we have tools from the API (enabled tools), use only those
	if len(tools) > 0 {
		// Add only the tools that are in the API response (enabled tools)
		uuidToName := map[string]string{
			ESLint:       "eslint",
			Trivy:        "trivy",
			PyLint:       "pylint",
			PMD:          "pmd",
			DartAnalyzer: "dartanalyzer",
			Semgrep:      "semgrep",
			Lizard:       "lizard",
		}

		for uuid, name := range uuidToName {
			if toolsMap[uuid] {
				sb.WriteString(fmt.Sprintf("    - %s@%s\n", name, toolVersions[uuid]))
			}
		}
	} else {
		// If no tools were specified (local mode), include all defaults
		sb.WriteString(fmt.Sprintf("    - eslint@%s\n", defaultVersions[ESLint]))
		sb.WriteString(fmt.Sprintf("    - trivy@%s\n", defaultVersions[Trivy]))
		sb.WriteString(fmt.Sprintf("    - pylint@%s\n", defaultVersions[PyLint]))
		sb.WriteString(fmt.Sprintf("    - pmd@%s\n", defaultVersions[PMD]))
		sb.WriteString(fmt.Sprintf("    - dartanalyzer@%s\n", defaultVersions[DartAnalyzer]))
		sb.WriteString(fmt.Sprintf("    - semgrep@%s\n", defaultVersions[Semgrep]))
		sb.WriteString(fmt.Sprintf("    - lizard@%s\n", defaultVersions[Lizard]))
	}

	return sb.String()
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

	// Clear any previous configuration files
	if err := cleanConfigDirectory(toolsConfigDir); err != nil {
		return fmt.Errorf("failed to clean configuration directory: %w", err)
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	apiTools, err := tools.GetRepositoryTools(CodacyApiBase, token, initFlags.provider, initFlags.organization, initFlags.repository)
	if err != nil {
		return err
	}

	// Map UUID to tool shortname for lookup
	uuidToName := map[string]string{
		ESLint:       "eslint",
		Trivy:        "trivy",
		PyLint:       "pylint",
		PMD:          "pmd",
		DartAnalyzer: "dartanalyzer",
	}

	// Generate languages configuration based on API tools response
	if err := tools.CreateLanguagesConfigFile(apiTools, toolsConfigDir, uuidToName, token, initFlags.provider, initFlags.organization, initFlags.repository); err != nil {
		return fmt.Errorf("failed to create languages configuration file: %w", err)
	}

	// Filter out any tools that use configuration file
	configuredToolsWithUI := tools.FilterToolsByConfigUsage(apiTools)

	// Create main config files with all enabled API tools
	err = createConfigurationFiles(apiTools, false)
	if err != nil {
		log.Fatal(err)
	}

	// Only generate config files for tools not using their own config file
	for _, tool := range configuredToolsWithUI {

		url := fmt.Sprintf("%s/api/v3/analysis/organizations/%s/%s/repositories/%s/tools/%s/patterns?enabled=true&limit=1000",
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
			fmt.Println("ESLint configuration created based on Codacy settings. Ignoring plugin rules. ESLint plugins are not supported yet.")
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
	case PyLint:
		if len(patternConfiguration) > 0 {
			err := createPylintConfigFile(patternConfiguration, toolsConfigDir)
			if err != nil {
				return fmt.Errorf("failed to create Pylint config: %v", err)
			}
		} else {
			err := createDefaultPylintConfigFile(toolsConfigDir)
			if err != nil {
				return fmt.Errorf("failed to create default Pylint config: %v", err)
			}
		}
		fmt.Println("Pylint configuration created based on Codacy settings")
	case DartAnalyzer:
		if len(patternConfiguration) > 0 {
			err := createDartAnalyzerConfigFile(patternConfiguration, toolsConfigDir)
			if err != nil {
				return fmt.Errorf("failed to create Dart Analyzer config: %v", err)
			}
		}
	case Semgrep:
		if len(patternConfiguration) > 0 {
			err := createSemgrepConfigFile(patternConfiguration, toolsConfigDir)
			if err != nil {
				return fmt.Errorf("failed to create Semgrep config: %v", err)
			}
		}
	}
	return nil
}

func createPMDConfigFile(config []domain.PatternConfiguration, toolsConfigDir string) error {
	pmdConfigurationString := tools.CreatePmdConfig(config)
	return os.WriteFile(filepath.Join(toolsConfigDir, "ruleset.xml"), []byte(pmdConfigurationString), utils.DefaultFilePerms)
}

func createDefaultPMDConfigFile(toolsConfigDir string) error {
	content := tools.CreatePmdConfig([]domain.PatternConfiguration{})
	return os.WriteFile(filepath.Join(toolsConfigDir, "ruleset.xml"), []byte(content), utils.DefaultFilePerms)

}

func createPylintConfigFile(config []domain.PatternConfiguration, toolsConfigDir string) error {
	pylintConfigurationString := pylint.GeneratePylintRC(config)
	return os.WriteFile(filepath.Join(toolsConfigDir, "pylint.rc"), []byte(pylintConfigurationString), utils.DefaultFilePerms)
}

func createDefaultPylintConfigFile(toolsConfigDir string) error {
	pylintConfigurationString := pylint.GeneratePylintRCDefault()
	return os.WriteFile(filepath.Join(toolsConfigDir, "pylint.rc"), []byte(pylintConfigurationString), utils.DefaultFilePerms)
}

// createTrivyConfigFile creates a trivy.yaml configuration file based on the API configuration
func createTrivyConfigFile(config []domain.PatternConfiguration, toolsConfigDir string) error {

	trivyConfigurationString := tools.CreateTrivyConfig(config)

	// Write to file
	return os.WriteFile(filepath.Join(toolsConfigDir, "trivy.yaml"), []byte(trivyConfigurationString), utils.DefaultFilePerms)
}

func createDartAnalyzerConfigFile(config []domain.PatternConfiguration, toolsConfigDir string) error {

	dartAnalyzerConfigurationString := tools.CreateDartAnalyzerConfig(config)
	return os.WriteFile(filepath.Join(toolsConfigDir, "analysis_options.yaml"), []byte(dartAnalyzerConfigurationString), utils.DefaultFilePerms)
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

// SemgrepRulesFile represents the structure of the rules.yaml file
type SemgrepRulesFile struct {
	Rules []map[string]interface{} `yaml:"rules"`
}

// createSemgrepConfigFile creates a semgrep.yaml configuration file based on the API configuration
func createSemgrepConfigFile(config []domain.PatternConfiguration, toolsConfigDir string) error {
	// Use the refactored function from tools package
	configData, err := tools.GetSemgrepConfig(config)

	if err != nil {
		return fmt.Errorf("failed to create Semgrep config: %v", err)
	}

	// Write to file
	return os.WriteFile(filepath.Join(toolsConfigDir, "semgrep.yaml"), configData, utils.DefaultFilePerms)
}

// cleanConfigDirectory removes all previous configuration files in the tools-configs directory
func cleanConfigDirectory(toolsConfigDir string) error {
	// Check if directory exists
	if _, err := os.Stat(toolsConfigDir); os.IsNotExist(err) {
		return nil // Directory doesn't exist, nothing to clean
	}

	// Read directory contents
	entries, err := os.ReadDir(toolsConfigDir)
	if err != nil {
		return fmt.Errorf("failed to read config directory: %w", err)
	}

	// Remove all files
	for _, entry := range entries {
		if !entry.IsDir() { // Only remove files, not subdirectories
			filePath := filepath.Join(toolsConfigDir, entry.Name())
			if err := os.Remove(filePath); err != nil {
				return fmt.Errorf("failed to remove file %s: %w", filePath, err)
			}
		}
	}

	fmt.Println("Cleaned previous configuration files")
	return nil
}

const (
	ESLint       string = "f8b29663-2cb2-498d-b923-a10c6a8c05cd"
	Trivy        string = "2fd7fbe0-33f9-4ab3-ab73-e9b62404e2cb"
	PMD          string = "9ed24812-b6ee-4a58-9004-0ed183c45b8f"
	PyLint       string = "31677b6d-4ae0-4f56-8041-606a8d7a8e61"
	DartAnalyzer string = "d203d615-6cf1-41f9-be5f-e2f660f7850f"
	Semgrep      string = "6792c561-236d-41b7-ba5e-9d6bee0d548b"
	Lizard       string = "76348462-84b3-409a-90d3-955e90abfb87"
)
