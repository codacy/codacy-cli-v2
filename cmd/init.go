package cmd

import (
	codacyclient "codacy/cli-v2/codacy-client"
	"codacy/cli-v2/config"
	"codacy/cli-v2/domain"
	"codacy/cli-v2/tools"
	"codacy/cli-v2/tools/lizard"
	"codacy/cli-v2/tools/pylint"
	"codacy/cli-v2/utils"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var initFlags domain.InitFlags

func init() {
	initCmd.Flags().StringVar(&initFlags.ApiToken, "api-token", "", "optional codacy api token, if defined configurations will be fetched from codacy")
	initCmd.Flags().StringVar(&initFlags.Provider, "provider", "", "provider (gh/bb/gl) to fetch configurations from codacy, required when api-token is provided")
	initCmd.Flags().StringVar(&initFlags.Organization, "organization", "", "remote organization name to fetch configurations from codacy, required when api-token is provided")
	initCmd.Flags().StringVar(&initFlags.Repository, "repository", "", "remote repository name to fetch configurations from codacy, required when api-token is provided")
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

		cliLocalMode := len(initFlags.ApiToken) == 0

		if cliLocalMode {
			fmt.Println()
			fmt.Println("ℹ️  No project token was specified, detecting languages and configuring tools...")

			// Create language detector and scan the project
			detector := domain.NewLanguageDetector()
			languages, err := detector.DetectLanguages(".")
			if err != nil {
				log.Fatalf("Failed to detect languages: %v", err)
			}

			// Map detected languages to tools
			var enabledTools []domain.Tool
			toolVersions := map[string]string{
				ESLint:       "8.57.0",
				Trivy:        "0.59.1",
				PyLint:       "3.3.6",
				PMD:          "6.55.0",
				DartAnalyzer: "3.7.2",
				Semgrep:      "1.78.0",
				Lizard:       "1.17.19",
			}

			// Map languages to tools
			languageToTools := map[string][]string{
				"JavaScript": {ESLint},
				"TypeScript": {ESLint},
				"Python":     {PyLint},
				"Java":       {PMD},
				"Dart":       {DartAnalyzer},
				"Go":         {Semgrep},
				"Ruby":       {Semgrep},
				"PHP":        {Semgrep},
				"C":          {Semgrep},
				"C++":        {Semgrep},
				"C#":         {Semgrep},
				"Kotlin":     {Semgrep},
				"Swift":      {Semgrep},
				"Scala":      {PMD},
				"Rust":       {Semgrep},
			}

			// Always enable Trivy for security scanning
			enabledTools = append(enabledTools, domain.Tool{
				Uuid:    Trivy,
				Name:    "trivy",
				Version: toolVersions[Trivy],
			})

			// Enable tools based on detected languages
			toolsEnabled := make(map[string]bool)
			for langName := range languages {
				if tools, ok := languageToTools[langName]; ok {
					for _, toolID := range tools {
						if !toolsEnabled[toolID] {
							toolsEnabled[toolID] = true
							enabledTools = append(enabledTools, domain.Tool{
								Uuid:    toolID,
								Name:    toolNameFromUUID(toolID),
								Version: toolVersions[toolID],
							})
						}
					}
				}
			}

			// Always enable Lizard for complexity analysis if any supported language is detected
			if shouldEnableLizard(languages) {
				enabledTools = append(enabledTools, domain.Tool{
					Uuid:    Lizard,
					Name:    "lizard",
					Version: toolVersions[Lizard],
				})
			}

			// Create configuration files
			err = createConfigurationFiles(enabledTools, cliLocalMode)
			if err != nil {
				log.Fatal(err)
			}

			// Create default configuration files for enabled tools
			if err := buildDefaultConfigurationFiles(toolsConfigDir, enabledTools); err != nil {
				log.Fatal(err)
			}

			// Print summary of detected languages and enabled tools
			fmt.Println("\nDetected languages:")
			for langName, info := range languages {
				fmt.Printf("  - %s (%d files)\n", langName, len(info.Files))
			}

			fmt.Println("\nEnabled tools:")
			for _, tool := range enabledTools {
				fmt.Printf("  - %s@%s\n", tool.Name, tool.Version)
			}
		} else {
			err := buildRepositoryConfigurationFiles(initFlags.ApiToken)
			if err != nil {
				log.Fatal(err)
			}
		}

		createGitIgnoreFile()
		fmt.Println()
		fmt.Println("✅ Successfully initialized Codacy configuration!")
		fmt.Println()
		fmt.Println("🔧 Next steps:")
		fmt.Println("  1. Run 'codacy-cli install' to install all dependencies")
		fmt.Println("  2. Run 'codacy-cli analyze' to start analyzing your code")
		fmt.Println()
	},
}

// shouldEnableLizard checks if Lizard should be enabled based on detected languages
func shouldEnableLizard(languages map[string]*domain.LanguageInfo) bool {
	lizardSupportedLangs := map[string]bool{
		"C": true, "C++": true, "Java": true, "C#": true,
		"JavaScript": true, "TypeScript": true, "Python": true,
		"Ruby": true, "PHP": true, "Scala": true, "Go": true,
		"Rust": true, "Kotlin": true, "Swift": true,
	}

	for lang := range languages {
		if lizardSupportedLangs[lang] {
			return true
		}
	}
	return false
}

// toolNameFromUUID returns the short name for a tool UUID
func toolNameFromUUID(uuid string) string {
	switch uuid {
	case ESLint:
		return "eslint"
	case Trivy:
		return "trivy"
	case PyLint:
		return "pylint"
	case PMD:
		return "pmd"
	case DartAnalyzer:
		return "dartanalyzer"
	case Semgrep:
		return "semgrep"
	case Lizard:
		return "lizard"
	default:
		return "unknown"
	}
}

func createGitIgnoreFile() error {
	gitIgnorePath := filepath.Join(config.Config.LocalCodacyDirectory(), ".gitignore")
	gitIgnoreFile, err := os.Create(gitIgnorePath)
	if err != nil {
		return fmt.Errorf("failed to create .gitignore file: %w", err)
	}
	defer gitIgnoreFile.Close()

	content := "# Codacy CLI\ntools-configs/\n.gitignore\ncli-config.yaml\nlogs/\n"
	if _, err := gitIgnoreFile.WriteString(content); err != nil {
		return fmt.Errorf("failed to write to .gitignore file: %w", err)
	}

	return nil
}

func createConfigurationFiles(tools []domain.Tool, cliLocalMode bool) error {
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

func configFileTemplate(tools []domain.Tool) string {
	// Maps to track which tools are enabled
	toolsMap := make(map[string]bool)
	toolVersions := make(map[string]string)

	// Track needed runtimes
	needsNode := false
	needsPython := false
	needsDart := false
	needsJava := false

	// Default versions
	defaultVersions := map[string]string{
		ESLint:       "8.57.0",
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
		} else if tool.Uuid == PMD {
			needsJava = true
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
		if needsJava {
			sb.WriteString("    - java@17.0.10\n")
		}
	} else {
		// In local mode with no tools specified, include all runtimes
		sb.WriteString("    - node@22.2.0\n")
		sb.WriteString("    - python@3.11.11\n")
		sb.WriteString("    - dart@3.7.2\n")
		sb.WriteString("    - java@17.0.10\n")
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

	apiTools, err := tools.GetRepositoryTools(initFlags)
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
		Lizard:       "lizard",
		Semgrep:      "semgrep",
	}

	// Generate languages configuration based on API tools response
	if err := tools.CreateLanguagesConfigFile(apiTools, toolsConfigDir, uuidToName, initFlags); err != nil {
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

		apiToolConfigurations, err := codacyclient.GetRepositoryToolPatterns(initFlags, tool.Uuid)

		if err != nil {
			fmt.Println("Error unmarshaling tool configurations:", err)
			return err
		}

		createToolFileConfigurations(tool, apiToolConfigurations)
	}

	return nil
}

// map tool uuid to tool name
func createToolFileConfigurations(tool domain.Tool, patternConfiguration []domain.PatternConfiguration) error {
	toolsConfigDir := config.Config.ToolsConfigDirectory()
	switch tool.Uuid {
	case ESLint:
		err := tools.CreateEslintConfig(toolsConfigDir, patternConfiguration)
		if err != nil {
			return fmt.Errorf("failed to write eslint config: %v", err)
		}
		fmt.Println("ESLint configuration created based on Codacy settings. Ignoring plugin rules. ESLint plugins are not supported yet.")
	case Trivy:
		err := createTrivyConfigFile(patternConfiguration, toolsConfigDir)
		if err != nil {
			return fmt.Errorf("failed to create Trivy config: %v", err)
		}
		fmt.Println("Trivy configuration created based on Codacy settings")
	case PMD:
		err := createPMDConfigFile(patternConfiguration, toolsConfigDir)
		if err != nil {
			return fmt.Errorf("failed to create PMD config: %v", err)
		}
		fmt.Println("PMD configuration created based on Codacy settings")
	case PyLint:
		err := createPylintConfigFile(patternConfiguration, toolsConfigDir)
		if err != nil {
			return fmt.Errorf("failed to create Pylint config: %v", err)
		}
		fmt.Println("Pylint configuration created based on Codacy settings")
	case DartAnalyzer:
		err := createDartAnalyzerConfigFile(patternConfiguration, toolsConfigDir)
		if err != nil {
			return fmt.Errorf("failed to create Dart Analyzer config: %v", err)
		}
		fmt.Println("Dart configuration created based on Codacy settings")
	case Semgrep:
		err := createSemgrepConfigFile(patternConfiguration, toolsConfigDir)
		if err != nil {
			return fmt.Errorf("failed to create Semgrep config: %v", err)
		}
		fmt.Println("Semgrep configuration created based on Codacy settings")
	case Lizard:
		err := createLizardConfigFile(toolsConfigDir, patternConfiguration)
		if err != nil {
			return fmt.Errorf("failed to create Lizard config: %v", err)
		}
		fmt.Println("Lizard configuration created based on Codacy settings")
	}
	return nil
}

func createPMDConfigFile(config []domain.PatternConfiguration, toolsConfigDir string) error {
	pmdConfigurationString := tools.CreatePmdConfig(config)
	return os.WriteFile(filepath.Join(toolsConfigDir, "ruleset.xml"), []byte(pmdConfigurationString), utils.DefaultFilePerms)
}

func createPylintConfigFile(config []domain.PatternConfiguration, toolsConfigDir string) error {
	pylintConfigurationString := pylint.GeneratePylintRC(config)
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

func createLizardConfigFile(toolsConfigDir string, patternConfiguration []domain.PatternConfiguration) error {
	patterns := make([]domain.PatternDefinition, len(patternConfiguration))
	for i, pattern := range patternConfiguration {
		patterns[i] = pattern.PatternDefinition

	}
	err := lizard.CreateLizardConfig(toolsConfigDir, patterns)
	if err != nil {
		return fmt.Errorf("failed to create Lizard configuration: %w", err)
	}
	return nil
}

// buildDefaultConfigurationFiles creates default configuration files for enabled tools
func buildDefaultConfigurationFiles(toolsConfigDir string, enabledTools []domain.Tool) error {
	for _, tool := range enabledTools {
		patternsConfig, err := codacyclient.GetDefaultToolPatternsConfig(initFlags, tool.Uuid)
		if err != nil {
			return fmt.Errorf("failed to get default tool patterns config: %w", err)
		}
		switch tool.Uuid {
		case ESLint:
			if err := tools.CreateEslintConfig(toolsConfigDir, patternsConfig); err != nil {
				return fmt.Errorf("failed to create eslint config file: %v", err)
			}
		case Trivy:
			if err := createTrivyConfigFile(patternsConfig, toolsConfigDir); err != nil {
				return fmt.Errorf("failed to create default Trivy configuration: %w", err)
			}
		case PMD:
			if err := createPMDConfigFile(patternsConfig, toolsConfigDir); err != nil {
				return fmt.Errorf("failed to create default PMD configuration: %w", err)
			}
		case PyLint:
			if err := createPylintConfigFile(patternsConfig, toolsConfigDir); err != nil {
				return fmt.Errorf("failed to create default Pylint configuration: %w", err)
			}
		case DartAnalyzer:
			if err := createDartAnalyzerConfigFile(patternsConfig, toolsConfigDir); err != nil {
				return fmt.Errorf("failed to create default Dart Analyzer configuration: %w", err)
			}
		case Semgrep:
			if err := createSemgrepConfigFile(patternsConfig, toolsConfigDir); err != nil {
				return fmt.Errorf("failed to create default Semgrep configuration: %w", err)
			}
		case Lizard:
			if err := createLizardConfigFile(toolsConfigDir, patternsConfig); err != nil {
				return fmt.Errorf("failed to create default Lizard configuration: %w", err)
			}
		}
	}
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

// AvailableTools lists all tool UUIDs supported by Codacy CLI.
var AvailableTools = []string{
	ESLint,
	Trivy,
	PMD,
	PyLint,
	DartAnalyzer,
	Semgrep,
	Lizard,
}
