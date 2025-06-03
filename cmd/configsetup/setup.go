package configsetup

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	codacyclient "codacy/cli-v2/codacy-client"
	"codacy/cli-v2/config"
	"codacy/cli-v2/domain"
	"codacy/cli-v2/plugins"
	"codacy/cli-v2/tools"
	"codacy/cli-v2/tools/lizard"
	"codacy/cli-v2/tools/pylint"
	"codacy/cli-v2/utils"
)

func CreateLanguagesConfigFileLocal(toolsConfigDir string) error {
	content := `tools:
    - name: pylint
      languages: [Python]
      extensions: [.py]
    - name: eslint
      languages: [JavaScript, TypeScript, JSX, TSX]
      extensions: [.js, .jsx, .ts, .tsx]
    - name: pmd
      languages: [Java, JavaScript, JSP, Velocity, XML, Apex, Scala, Ruby, VisualForce]
      extensions: [.java, .js, .jsp, .vm, .xml, .cls, .trigger, .scala, .rb, .page, .component]
    - name: trivy
      languages: [Multiple]
      extensions: []
    - name: dartanalyzer
      languages: [Dart]
      extensions: [.dart]
    - name: lizard
      languages: [C, CPP, Java, "C#", JavaScript, TypeScript, VueJS, "Objective-C", Swift, Python, Ruby, "TTCN-3", PHP, Scala, GDScript, Golang, Lua, Rust, Fortran, Kotlin, Solidity, Erlang, Zig, Perl]
      extensions: [.c, .cpp, .cc, .h, .hpp, .java, .cs, .js, .jsx, .ts, .tsx, .vue, .m, .swift, .py, .rb, .ttcn, .php, .scala, .gd, .go, .lua, .rs, .f, .f90, .kt, .sol, .erl, .zig, .pl]
    - name: semgrep
      languages: [C, CPP, "C#", Generic, Go, Java, JavaScript, JSON, Kotlin, Python, TypeScript, Ruby, Rust, JSX, PHP, Scala, Swift, Terraform]
      extensions: [.c, .cpp, .h, .hpp, .cs, .go, .java, .js, .json, .kt, .py, .ts, .rb, .rs, .jsx, .php, .scala, .swift, .tf, .tfvars]
    - name: codacy-enigma-cli
      languages: [Multiple]
      extensions: []`

	return os.WriteFile(filepath.Join(toolsConfigDir, "languages-config.yaml"), []byte(content), utils.DefaultFilePerms)
}

func CreateGitIgnoreFile() error {
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

func CreateConfigurationFiles(tools []domain.Tool, cliLocalMode bool) error {
	configFile, err := os.Create(config.Config.ProjectConfigFile())
	if err != nil {
		return fmt.Errorf("failed to create project config file: %w", err)
	}
	defer configFile.Close()

	configContent := ConfigFileTemplate(tools)
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

// RuntimePluginConfig holds the structure of the runtime plugin.yaml file
type RuntimePluginConfig struct {
	Name           string `yaml:"name"`
	Description    string `yaml:"description"`
	DefaultVersion string `yaml:"default_version"`
}

func ConfigFileTemplate(tools []domain.Tool) string {
	// Maps to track which tools are enabled
	toolsMap := make(map[string]bool)
	toolVersions := make(map[string]string)

	toolsWithLatestVersion, _, _ := KeepToolsWithLatestVersion(tools)

	// Track needed runtimes
	neededRuntimes := make(map[string]bool)

	// Get tool versions from plugin configurations
	defaultVersions := plugins.GetToolVersions()

	// Get runtime versions all at once
	runtimeVersions := plugins.GetRuntimeVersions()

	// Get tool runtime dependencies
	runtimeDependencies := plugins.GetToolRuntimeDependencies()

	// Build map of enabled tools with their versions
	for _, tool := range toolsWithLatestVersion {
		toolsMap[tool.Uuid] = true
		if tool.Version != "" {
			toolVersions[tool.Uuid] = tool.Version
		} else {
			if meta, ok := domain.SupportedToolsMetadata[tool.Uuid]; ok {
				if defaultVersion, ok := defaultVersions[meta.Name]; ok {
					toolVersions[tool.Uuid] = defaultVersion
				}
			}
		}

		// Get the tool's runtime dependency
		if meta, ok := domain.SupportedToolsMetadata[tool.Uuid]; ok {
			if runtime, ok := runtimeDependencies[meta.Name]; ok {
				// Handle special case for dartanalyzer which can use either dart or flutter
				if meta.Name == "dartanalyzer" {
					// For now, default to dart runtime
					neededRuntimes["dart"] = true
				} else {
					neededRuntimes[runtime] = true
				}
			}
		}
	}

	// Start building the YAML content
	var sb strings.Builder
	sb.WriteString("runtimes:\n")

	// Only include runtimes needed by the enabled tools
	if len(tools) > 0 {
		// Create a sorted slice of runtimes
		var sortedRuntimes []string
		for runtime := range neededRuntimes {
			sortedRuntimes = append(sortedRuntimes, runtime)
		}
		sort.Strings(sortedRuntimes)

		// Write sorted runtimes
		for _, runtime := range sortedRuntimes {
			sb.WriteString(fmt.Sprintf("    - %s@%s\n", runtime, runtimeVersions[runtime]))
		}
	} else {
		// In local mode with no tools specified, include only the necessary runtimes
		supportedTools, err := plugins.GetSupportedTools()
		if err != nil {
			log.Printf("Warning: failed to get supported tools: %v", err)
			return sb.String()
		}

		// Get runtimes needed by supported tools
		for toolName := range supportedTools {
			if runtime, ok := runtimeDependencies[toolName]; ok {
				if toolName == "dartanalyzer" {
					neededRuntimes["dart"] = true
				} else {
					neededRuntimes[runtime] = true
				}
			}
		}
		var sortedRuntimes []string
		for runtime := range neededRuntimes {
			sortedRuntimes = append(sortedRuntimes, runtime)
		}
		sort.Strings(sortedRuntimes)
		for _, runtime := range sortedRuntimes {
			sb.WriteString(fmt.Sprintf("    - %s@%s\n", runtime, runtimeVersions[runtime]))
		}
	}

	sb.WriteString("tools:\n")

	if len(tools) > 0 {
		// Create a sorted slice of tool names
		var sortedTools []string
		for uuid, meta := range domain.SupportedToolsMetadata {
			if toolsMap[uuid] {
				sortedTools = append(sortedTools, meta.Name)
			}
		}
		sort.Strings(sortedTools)

		// Write sorted tools
		for _, name := range sortedTools {
			// Find the UUID for this tool name to get its version
			for uuid, meta := range domain.SupportedToolsMetadata {
				if meta.Name == name && toolsMap[uuid] {
					version := toolVersions[uuid]
					sb.WriteString(fmt.Sprintf("    - %s@%s\n", name, version))
					break
				}
			}
		}
	} else {
		// If no tools were specified (local mode), include all tools in sorted order
		var sortedTools []string

		// Get supported tools from plugin system
		supportedTools, err := plugins.GetSupportedTools()
		if err != nil {
			log.Printf("Warning: failed to get supported tools: %v", err)
			return sb.String()
		}

		// Convert map keys to slice and sort them
		for toolName := range supportedTools {
			if version, ok := defaultVersions[toolName]; ok {
				// Skip tools without a version
				if version != "" {
					sortedTools = append(sortedTools, toolName)
				}
			}
		}
		sort.Strings(sortedTools)

		// Write sorted tools
		for _, toolName := range sortedTools {
			if version, ok := defaultVersions[toolName]; ok {
				sb.WriteString(fmt.Sprintf("    - %s@%s\n", toolName, version))
			}
		}
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

func BuildRepositoryConfigurationFiles(flags domain.InitFlags) error {
	fmt.Println("Fetching repository configuration from codacy ...")

	toolsConfigDir := config.Config.ToolsConfigDirectory()

	// Create tools-configs directory if it doesn't exist
	if err := os.MkdirAll(toolsConfigDir, utils.DefaultDirPerms); err != nil {
		return fmt.Errorf("failed to create tools-configs directory: %w", err)
	}

	// Clear any previous configuration files
	if err := CleanConfigDirectory(toolsConfigDir); err != nil {
		return fmt.Errorf("failed to clean configuration directory: %w", err)
	}

	apiTools, err := tools.GetRepositoryTools(flags)
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
	if err := tools.CreateLanguagesConfigFile(toolsWithLatestVersion, toolsConfigDir, uuidToName, flags); err != nil {
		return fmt.Errorf("failed to create languages configuration file: %w", err)
	}

	// Filter out any tools that use configuration file
	configuredToolsWithUI := tools.FilterToolsByConfigUsage(toolsWithLatestVersion)

	// Create main config files with all enabled API tools
	err = CreateConfigurationFiles(toolsWithLatestVersion, false)
	if err != nil {
		log.Fatal(err)
	}

	// Only generate config files for tools not using their own config file
	for _, tool := range configuredToolsWithUI {

		apiToolConfigurations, err := codacyclient.GetRepositoryToolPatterns(flags, tool.Uuid)

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
func CleanConfigDirectory(toolsConfigDir string) error {
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
func BuildDefaultConfigurationFiles(toolsConfigDir string, flags domain.InitFlags) error {
	for uuid := range domain.SupportedToolsMetadata {
		patternsConfig, err := codacyclient.GetDefaultToolPatternsConfig(flags, uuid)
		if err != nil {
			return fmt.Errorf("failed to get default tool patterns config: %w", err)
		}
		switch uuid {
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
		case domain.PMD7, domain.ESLint9:
			continue
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
