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
			fmt.Println("ℹ️  No project token was specified, fetching codacy default configurations")
			noTools := []domain.Tool{}
			err := createConfigurationFiles(noTools, cliLocalMode)
			if err != nil {
				log.Fatal(err)
			}
			// Create default configuration files
			if err := buildDefaultConfigurationFiles(toolsConfigDir); err != nil {
				log.Fatal(err)
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

	toolsWithLatestVersion, _, _ := KeepToolsWithLatestVersion(tools)

	// Track needed runtimes
	needsNode := false
	needsPython := false
	needsDart := false
	needsJava := false

	// Default versions
	defaultVersions := map[string]string{
		domain.ESLint:       "8.57.0",
		domain.Trivy:        "0.59.1",
		domain.PyLint:       "3.3.6",
		domain.PMD:          "6.55.0",
		domain.DartAnalyzer: "3.7.2",
		domain.Semgrep:      "1.78.0",
		domain.Lizard:       "1.17.19",
	}

	// Build map of enabled tools with their versions
	for _, tool := range toolsWithLatestVersion {
		toolsMap[tool.Uuid] = true
		if tool.Version != "" {
			toolVersions[tool.Uuid] = tool.Version
		} else {
			toolVersions[tool.Uuid] = defaultVersions[tool.Uuid]
		}

		// Check if tool needs a runtime
		switch tool.Uuid {
		case domain.ESLint, domain.ESLint9:
			needsNode = true
		case domain.PyLint, domain.Lizard:
			needsPython = true
		case domain.DartAnalyzer:
			needsDart = true
		case domain.PMD, domain.PMD7:
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
		for uuid, meta := range domain.SupportedToolsMetadata {
			if toolsMap[uuid] {
				sb.WriteString(fmt.Sprintf("    - %s@%s\n", meta.Name, toolVersions[uuid]))
			}
		}

	} else {
		// If no tools were specified (local mode), include all defaults
		sb.WriteString(fmt.Sprintf("    - eslint@%s\n", defaultVersions[domain.ESLint]))
		sb.WriteString(fmt.Sprintf("    - trivy@%s\n", defaultVersions[domain.Trivy]))
		sb.WriteString(fmt.Sprintf("    - pylint@%s\n", defaultVersions[domain.PyLint]))
		sb.WriteString(fmt.Sprintf("    - pmd@%s\n", defaultVersions[domain.PMD]))
		sb.WriteString(fmt.Sprintf("    - dartanalyzer@%s\n", defaultVersions[domain.DartAnalyzer]))
		sb.WriteString(fmt.Sprintf("    - semgrep@%s\n", defaultVersions[domain.Semgrep]))
		sb.WriteString(fmt.Sprintf("    - lizard@%s\n", defaultVersions[domain.Lizard]))
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

	toolsWithLatestVersion, uuidToName, familyToVersions := KeepToolsWithLatestVersion(apiTools)

	for family, versions := range familyToVersions {
		if len(versions) > 1 {
			kept := ", "
			for _, tool := range toolsWithLatestVersion {
				if domain.SupportedToolsMetadata[tool.Uuid].Name == family {
					kept = tool.Version
					break
				}
			}
			fmt.Printf("⚠️  Multiple versions of '%s' detected: [%s], keeping %s\n", family, strings.Join(versions, ", "), kept)
		}
	}

	// Generate languages configuration based on API tools response
	if err := tools.CreateLanguagesConfigFile(toolsWithLatestVersion, toolsConfigDir, uuidToName, initFlags); err != nil {
		return fmt.Errorf("failed to create languages configuration file: %w", err)
	}

	// Filter out any tools that use configuration file
	configuredToolsWithUI := tools.FilterToolsByConfigUsage(toolsWithLatestVersion)

	// Create main config files with all enabled API tools
	err = createConfigurationFiles(toolsWithLatestVersion, false)
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
	case domain.ESLint, domain.ESLint9:
		err := tools.CreateEslintConfig(toolsConfigDir, patternConfiguration)
		if err != nil {
			return fmt.Errorf("failed to write eslint config: %v", err)
		}
		fmt.Println("ESLint configuration created based on Codacy settings. Ignoring plugin rules. ESLint plugins are not supported yet.")
	case domain.Trivy:
		err := createTrivyConfigFile(patternConfiguration, toolsConfigDir)
		if err != nil {
			return fmt.Errorf("failed to create Trivy config: %v", err)
		}
		fmt.Println("Trivy configuration created based on Codacy settings")
	case domain.PMD:
		err := createPMDConfigFile(patternConfiguration, toolsConfigDir)
		if err != nil {
			return fmt.Errorf("failed to create PMD config: %v", err)
		}
	case domain.PMD7:
		err := createPMD7ConfigFile(patternConfiguration, toolsConfigDir)
		if err != nil {
			return fmt.Errorf("failed to create PMD7 config: %v", err)
		}
		fmt.Println("PMD7 configuration created based on Codacy settings")
	case domain.PyLint:
		err := createPylintConfigFile(patternConfiguration, toolsConfigDir)
		if err != nil {
			return fmt.Errorf("failed to create Pylint config: %v", err)
		}
		fmt.Println("Pylint configuration created based on Codacy settings")
	case domain.DartAnalyzer:
		err := createDartAnalyzerConfigFile(patternConfiguration, toolsConfigDir)
		if err != nil {
			return fmt.Errorf("failed to create Dart Analyzer config: %v", err)
		}
		fmt.Println("Dart configuration created based on Codacy settings")
	case domain.Semgrep:
		err := createSemgrepConfigFile(patternConfiguration, toolsConfigDir)
		if err != nil {
			return fmt.Errorf("failed to create Semgrep config: %v", err)
		}
		fmt.Println("Semgrep configuration created based on Codacy settings")
	case domain.Lizard:
		err := createLizardConfigFile(toolsConfigDir, patternConfiguration)
		if err != nil {
			return fmt.Errorf("failed to create Lizard config: %v", err)
		}
		fmt.Println("Lizard configuration created based on Codacy settings")
	}
	return nil
}

func createPMDConfigFile(config []domain.PatternConfiguration, toolsConfigDir string) error {
	pmdConfigurationString := tools.CreatePmd6Config(config)
	return os.WriteFile(filepath.Join(toolsConfigDir, "ruleset.xml"), []byte(pmdConfigurationString), utils.DefaultFilePerms)
}

func createPMD7ConfigFile(config []domain.PatternConfiguration, toolsConfigDir string) error {
	pmdConfigurationString := tools.CreatePmd7Config(config)
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

// buildDefaultConfigurationFiles creates default configuration files for all tools
func buildDefaultConfigurationFiles(toolsConfigDir string) error {
	for tool := range domain.SupportedToolsMetadata {
		patternsConfig, err := codacyclient.GetDefaultToolPatternsConfig(initFlags, tool)
		if err != nil {
			return fmt.Errorf("failed to get default tool patterns config: %w", err)
		}
		switch tool {
		case domain.ESLint:
			if err := tools.CreateEslintConfig(toolsConfigDir, patternsConfig); err != nil {
				return fmt.Errorf("failed to create eslint config file: %v", err)
			}
		case domain.Trivy:
			if err := createTrivyConfigFile(patternsConfig, toolsConfigDir); err != nil {
				return fmt.Errorf("failed to create default Trivy configuration: %w", err)
			}
		case domain.PMD:
			if err := createPMDConfigFile(patternsConfig, toolsConfigDir); err != nil {
				return fmt.Errorf("failed to create default PMD configuration: %w", err)
			}
		case domain.PMD7, domain.ESLint9:
			continue
		case domain.PyLint:
			if err := createPylintConfigFile(patternsConfig, toolsConfigDir); err != nil {
				return fmt.Errorf("failed to create default Pylint configuration: %w", err)
			}
		case domain.DartAnalyzer:
			if err := createDartAnalyzerConfigFile(patternsConfig, toolsConfigDir); err != nil {
				return fmt.Errorf("failed to create default Dart Analyzer configuration: %w", err)
			}
		case domain.Semgrep:
			if err := createSemgrepConfigFile(patternsConfig, toolsConfigDir); err != nil {
				return fmt.Errorf("failed to create default Semgrep configuration: %w", err)
			}
		case domain.Lizard:
			if err := createLizardConfigFile(toolsConfigDir, patternsConfig); err != nil {
				return fmt.Errorf("failed to create default Lizard configuration: %w", err)
			}
		}
	}
	return nil
}

// KeepToolsWithLatestVersion filters the tools to keep only the latest version of each tool family.
func KeepToolsWithLatestVersion(tools []domain.Tool) (
	toolsWithLatestVersion []domain.Tool,
	uuidToName map[string]string,
	familyToVersions map[string][]string,
) {
	latestTools := map[string]domain.Tool{}
	uuidToName = map[string]string{}
	seen := map[string][]domain.Tool{}
	familyToVersions = map[string][]string{}

	for _, tool := range tools {
		meta, ok := domain.SupportedToolsMetadata[tool.Uuid]
		if !ok {
			continue
		}

		// Track all tools seen per family
		seen[meta.Name] = append(seen[meta.Name], tool)

		// Pick the best version
		current, exists := latestTools[meta.Name]
		if !exists || domain.SupportedToolsMetadata[current.Uuid].Priority > meta.Priority {
			latestTools[meta.Name] = tool
			uuidToName[tool.Uuid] = meta.Name
		}
	}

	// Populate final list and version map for logging
	for family, tools := range seen {
		var versions []string
		for _, t := range tools {
			v := t.Version
			if v == "" {
				v = "(unknown)"
			}
			versions = append(versions, v)
		}
		familyToVersions[family] = versions
	}

	for _, tool := range latestTools {
		toolsWithLatestVersion = append(toolsWithLatestVersion, tool)
	}

	return
}
