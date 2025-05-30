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

// Tool UUID constants
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

// Map tool UUIDs to their names
var toolNameMap = map[string]string{
	ESLint:       "eslint",
	Trivy:        "trivy",
	PyLint:       "pylint",
	PMD:          "pmd",
	DartAnalyzer: "dartanalyzer",
	Semgrep:      "semgrep",
	Lizard:       "lizard",
}

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

func CreateConfigurationFiles(toolsToUse []domain.Tool, cliLocalMode bool) error {
	configFile, err := os.Create(config.Config.ProjectConfigFile())
	if err != nil {
		return fmt.Errorf("failed to create project config file: %w", err)
	}
	defer configFile.Close()

	configContent := ConfigFileTemplate(toolsToUse)
	_, err = configFile.WriteString(configContent)
	if err != nil {
		return fmt.Errorf("failed to write project config file: %w", err)
	}

	cliConfigFile, err := os.Create(config.Config.CliConfigFile())
	if err != nil {
		return fmt.Errorf("failed to create CLI config file: %w", err)
	}
	defer cliConfigFile.Close()

	cliConfigContent := CliConfigFileTemplate(cliLocalMode)
	_, err = cliConfigFile.WriteString(cliConfigContent)
	if err != nil {
		return fmt.Errorf("failed to write CLI config file: %w", err)
	}

	return nil
}

func ConfigFileTemplate(toolsToUse []domain.Tool) string {
	toolsMap := make(map[string]bool)
	toolVersions := make(map[string]string)
	neededRuntimes := make(map[string]bool)
	defaultVersions := plugins.GetToolVersions()
	runtimeVersions := plugins.GetRuntimeVersions()
	runtimeDependencies := plugins.GetToolRuntimeDependencies()

	for _, tool := range toolsToUse {
		toolsMap[tool.Uuid] = true
		if tool.Version != "" {
			toolVersions[tool.Uuid] = tool.Version
		} else {
			toolName := toolNameMap[tool.Uuid]
			if defaultVersion, ok := defaultVersions[toolName]; ok {
				toolVersions[tool.Uuid] = defaultVersion
			}
		}

		toolName := toolNameMap[tool.Uuid]
		if toolName != "" {
			if runtime, ok := runtimeDependencies[toolName]; ok {
				if toolName == "dartanalyzer" {
					neededRuntimes["dart"] = true
				} else {
					neededRuntimes[runtime] = true
				}
			}
		}
	}

	var sb strings.Builder
	sb.WriteString("runtimes:\n")

	if len(toolsToUse) > 0 {
		var sortedRuntimes []string
		for runtime := range neededRuntimes {
			sortedRuntimes = append(sortedRuntimes, runtime)
		}
		sort.Strings(sortedRuntimes)
		for _, runtime := range sortedRuntimes {
			sb.WriteString(fmt.Sprintf("    - %s@%s\n", runtime, runtimeVersions[runtime]))
		}
	} else {
		supportedTools, err := plugins.GetSupportedTools()
		if err != nil {
			log.Printf("Warning: failed to get supported tools: %v", err)
			return sb.String()
		}
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
	if len(toolsToUse) > 0 {
		var sortedTools []string
		for uuid, name := range toolNameMap {
			if toolsMap[uuid] {
				sortedTools = append(sortedTools, name)
			}
		}
		sort.Strings(sortedTools)
		for _, name := range sortedTools {
			for uuid, toolNameLookup := range toolNameMap {
				if toolNameLookup == name && toolsMap[uuid] {
					version := toolVersions[uuid]
					sb.WriteString(fmt.Sprintf("    - %s@%s\n", name, version))
					break
				}
			}
		}
	} else {
		var sortedTools []string
		supportedTools, err := plugins.GetSupportedTools()
		if err != nil {
			log.Printf("Warning: failed to get supported tools: %v", err)
			return sb.String()
		}
		for toolName := range supportedTools {
			if version, ok := defaultVersions[toolName]; ok {
				if version != "" {
					sortedTools = append(sortedTools, toolName)
				}
			}
		}
		sort.Strings(sortedTools)
		for _, toolName := range sortedTools {
			if version, ok := defaultVersions[toolName]; ok {
				sb.WriteString(fmt.Sprintf("    - %s@%s\n", toolName, version))
			}
		}
	}
	return sb.String()
}

func CliConfigFileTemplate(cliLocalMode bool) string {
	var cliModeString string
	if cliLocalMode {
		cliModeString = "local"
	} else {
		cliModeString = "remote"
	}
	return fmt.Sprintf(`mode: %s`, cliModeString)
}

func BuildRepositoryConfigurationFiles(initFlags domain.InitFlags) error {
	fmt.Println("Fetching repository configuration from codacy ...")
	toolsConfigDir := config.Config.ToolsConfigDirectory()
	if err := os.MkdirAll(toolsConfigDir, utils.DefaultDirPerms); err != nil {
		return fmt.Errorf("failed to create tools-configs directory: %w", err)
	}
	if err := CleanConfigDirectory(toolsConfigDir); err != nil {
		return fmt.Errorf("failed to clean configuration directory: %w", err)
	}

	apiTools, err := tools.GetRepositoryTools(initFlags)
	if err != nil {
		return err
	}

	uuidToName := map[string]string{
		ESLint:       "eslint",
		Trivy:        "trivy",
		PyLint:       "pylint",
		PMD:          "pmd",
		DartAnalyzer: "dartanalyzer",
		Lizard:       "lizard",
		Semgrep:      "semgrep",
	}

	if err := tools.CreateLanguagesConfigFile(apiTools, toolsConfigDir, uuidToName, initFlags); err != nil {
		return fmt.Errorf("failed to create languages configuration file: %w", err)
	}

	configuredToolsWithUI := tools.FilterToolsByConfigUsage(apiTools)
	err = CreateConfigurationFiles(apiTools, false)
	if err != nil {
		log.Fatal(err)
	}

	for _, tool := range configuredToolsWithUI {
		apiToolConfigurations, err := codacyclient.GetRepositoryToolPatterns(initFlags, tool.Uuid)
		if err != nil {
			fmt.Println("Error unmarshaling tool configurations:", err)
			return err
		}
		CreateToolFileConfigurations(tool, apiToolConfigurations, initFlags)
	}
	return nil
}

func CreateToolFileConfigurations(tool domain.Tool, patternConfiguration []domain.PatternConfiguration, initFlags domain.InitFlags) error {
	toolsConfigDir := config.Config.ToolsConfigDirectory()
	switch tool.Uuid {
	case ESLint:
		err := tools.CreateEslintConfig(toolsConfigDir, patternConfiguration)
		if err != nil {
			return fmt.Errorf("failed to write eslint config: %v", err)
		}
		fmt.Println("ESLint configuration created based on Codacy settings. Ignoring plugin rules. ESLint plugins are not supported yet.")
	case Trivy:
		err := CreateTrivyConfigFile(patternConfiguration, toolsConfigDir)
		if err != nil {
			return fmt.Errorf("failed to create Trivy config: %v", err)
		}
		fmt.Println("Trivy configuration created based on Codacy settings")
	case PMD:
		err := CreatePMDConfigFile(patternConfiguration, toolsConfigDir)
		if err != nil {
			return fmt.Errorf("failed to create PMD config: %v", err)
		}
		fmt.Println("PMD configuration created based on Codacy settings")
	case PyLint:
		err := CreatePylintConfigFile(patternConfiguration, toolsConfigDir)
		if err != nil {
			return fmt.Errorf("failed to create Pylint config: %v", err)
		}
		fmt.Println("Pylint configuration created based on Codacy settings")
	case DartAnalyzer:
		err := CreateDartAnalyzerConfigFile(patternConfiguration, toolsConfigDir)
		if err != nil {
			return fmt.Errorf("failed to create Dart Analyzer config: %v", err)
		}
		fmt.Println("Dart configuration created based on Codacy settings")
	case Semgrep:
		err := CreateSemgrepConfigFile(patternConfiguration, toolsConfigDir)
		if err != nil {
			return fmt.Errorf("failed to create Semgrep config: %v", err)
		}
		fmt.Println("Semgrep configuration created based on Codacy settings")
	case Lizard:
		err := CreateLizardConfigFile(toolsConfigDir, patternConfiguration)
		if err != nil {
			return fmt.Errorf("failed to create Lizard config: %v", err)
		}
		fmt.Println("Lizard configuration created based on Codacy settings")
	}
	return nil
}

func CreatePMDConfigFile(config []domain.PatternConfiguration, toolsConfigDir string) error {
	pmdConfigurationString := tools.CreatePmdConfig(config)
	return os.WriteFile(filepath.Join(toolsConfigDir, "ruleset.xml"), []byte(pmdConfigurationString), utils.DefaultFilePerms)
}

func CreatePylintConfigFile(configPatterns []domain.PatternConfiguration, toolsConfigDir string) error {
	pylintConfigurationString := pylint.GeneratePylintRC(configPatterns)
	return os.WriteFile(filepath.Join(toolsConfigDir, "pylint.rc"), []byte(pylintConfigurationString), utils.DefaultFilePerms)
}

func CreateTrivyConfigFile(configPatterns []domain.PatternConfiguration, toolsConfigDir string) error {
	trivyConfigurationString := tools.CreateTrivyConfig(configPatterns)
	return os.WriteFile(filepath.Join(toolsConfigDir, "trivy.yaml"), []byte(trivyConfigurationString), utils.DefaultFilePerms)
}

func CreateDartAnalyzerConfigFile(configPatterns []domain.PatternConfiguration, toolsConfigDir string) error {
	dartAnalyzerConfigurationString := tools.CreateDartAnalyzerConfig(configPatterns)
	return os.WriteFile(filepath.Join(toolsConfigDir, "analysis_options.yaml"), []byte(dartAnalyzerConfigurationString), utils.DefaultFilePerms)
}

func CreateSemgrepConfigFile(configPatterns []domain.PatternConfiguration, toolsConfigDir string) error {
	configData, err := tools.GetSemgrepConfig(configPatterns)
	if err != nil {
		return fmt.Errorf("failed to create Semgrep config: %v", err)
	}
	return os.WriteFile(filepath.Join(toolsConfigDir, "semgrep.yaml"), configData, utils.DefaultFilePerms)
}

func CleanConfigDirectory(toolsConfigDir string) error {
	if _, err := os.Stat(toolsConfigDir); os.IsNotExist(err) {
		return nil
	}
	entries, err := os.ReadDir(toolsConfigDir)
	if err != nil {
		return fmt.Errorf("failed to read config directory: %w", err)
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			filePath := filepath.Join(toolsConfigDir, entry.Name())
			if err := os.Remove(filePath); err != nil {
				return fmt.Errorf("failed to remove file %s: %w", filePath, err)
			}
		}
	}
	fmt.Println("Cleaned previous configuration files")
	return nil
}

func CreateLizardConfigFile(toolsConfigDir string, patternConfiguration []domain.PatternConfiguration) error {
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

func BuildDefaultConfigurationFiles(toolsConfigDir string, initFlags domain.InitFlags) error {
	for _, tool := range AvailableTools {
		patternsConfig, err := codacyclient.GetDefaultToolPatternsConfig(initFlags, tool)
		if err != nil {
			return fmt.Errorf("failed to get default tool patterns config: %w", err)
		}
		switch tool {
		case ESLint:
			if err := tools.CreateEslintConfig(toolsConfigDir, patternsConfig); err != nil {
				return fmt.Errorf("failed to create eslint config file: %v", err)
			}
		case Trivy:
			if err := CreateTrivyConfigFile(patternsConfig, toolsConfigDir); err != nil {
				return fmt.Errorf("failed to create default Trivy configuration: %w", err)
			}
		case PMD:
			if err := CreatePMDConfigFile(patternsConfig, toolsConfigDir); err != nil {
				return fmt.Errorf("failed to create default PMD configuration: %w", err)
			}
		case PyLint:
			if err := CreatePylintConfigFile(patternsConfig, toolsConfigDir); err != nil {
				return fmt.Errorf("failed to create default Pylint configuration: %w", err)
			}
		case DartAnalyzer:
			if err := CreateDartAnalyzerConfigFile(patternsConfig, toolsConfigDir); err != nil {
				return fmt.Errorf("failed to create default Dart Analyzer configuration: %w", err)
			}
		case Semgrep:
			if err := CreateSemgrepConfigFile(patternsConfig, toolsConfigDir); err != nil {
				return fmt.Errorf("failed to create default Semgrep configuration: %w", err)
			}
		case Lizard:
			if err := CreateLizardConfigFile(toolsConfigDir, patternsConfig); err != nil {
				return fmt.Errorf("failed to create default Lizard configuration: %w", err)
			}
		}
	}
	return nil
}
