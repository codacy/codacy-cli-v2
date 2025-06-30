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
	"codacy/cli-v2/constants"
	"codacy/cli-v2/domain"
	"codacy/cli-v2/plugins"
	"codacy/cli-v2/tools"
	"codacy/cli-v2/tools/lizard"
	"codacy/cli-v2/tools/pylint"
	reviveTool "codacy/cli-v2/tools/revive"
	"codacy/cli-v2/utils"

	"gopkg.in/yaml.v3"
)

// ToolConfigCreator defines the interface for tool configuration creators
type ToolConfigCreator interface {
	CreateConfig(toolsConfigDir string, patterns []domain.PatternConfiguration) error
	GetConfigFileName() string
	GetToolName() string
}

// toolConfigRegistry maps tool UUIDs to their configuration creators
var toolConfigRegistry = map[string]ToolConfigCreator{
	domain.ESLint:       &eslintConfigCreator{},
	domain.ESLint9:      &eslintConfigCreator{},
	domain.Trivy:        &trivyConfigCreator{},
	domain.PMD:          &pmdConfigCreator{},
	domain.PMD7:         &pmd7ConfigCreator{},
	domain.PyLint:       &pylintConfigCreator{},
	domain.DartAnalyzer: &dartAnalyzerConfigCreator{},
	domain.Semgrep:      &semgrepConfigCreator{},
	domain.Lizard:       &lizardConfigCreator{},
	domain.Revive:       &reviveConfigCreator{},
}

// writeConfigFile is a helper function to write configuration files with consistent error handling
func writeConfigFile(filePath string, content []byte) error {
	return os.WriteFile(filePath, content, constants.DefaultFilePerms)
}

// eslintConfigCreator implements ToolConfigCreator for ESLint
type eslintConfigCreator struct{}

func (e *eslintConfigCreator) CreateConfig(toolsConfigDir string, patterns []domain.PatternConfiguration) error {
	err := tools.CreateEslintConfig(toolsConfigDir, patterns)
	if err == nil {
		fmt.Println("ESLint configuration created based on Codacy settings. Ignoring plugin rules. ESLint plugins are not supported yet.")
	}
	return err
}

func (e *eslintConfigCreator) GetConfigFileName() string { return "eslint.config.mjs" }
func (e *eslintConfigCreator) GetToolName() string       { return "ESLint" }

// trivyConfigCreator implements ToolConfigCreator for Trivy
type trivyConfigCreator struct{}

func (t *trivyConfigCreator) CreateConfig(toolsConfigDir string, patterns []domain.PatternConfiguration) error {
	configString := tools.CreateTrivyConfig(patterns)
	err := writeConfigFile(filepath.Join(toolsConfigDir, constants.TrivyConfigFileName), []byte(configString))
	if err == nil {
		fmt.Println("Trivy configuration created based on Codacy settings")
	}
	return err
}

func (t *trivyConfigCreator) GetConfigFileName() string { return constants.TrivyConfigFileName }
func (t *trivyConfigCreator) GetToolName() string       { return "Trivy" }

// pmdConfigCreator implements ToolConfigCreator for PMD
type pmdConfigCreator struct{}

func (p *pmdConfigCreator) CreateConfig(toolsConfigDir string, patterns []domain.PatternConfiguration) error {
	configString := tools.CreatePmd6Config(patterns)
	return writeConfigFile(filepath.Join(toolsConfigDir, constants.PMDConfigFileName), []byte(configString))
}

func (p *pmdConfigCreator) GetConfigFileName() string { return constants.PMDConfigFileName }
func (p *pmdConfigCreator) GetToolName() string       { return "PMD" }

// pmd7ConfigCreator implements ToolConfigCreator for PMD7
type pmd7ConfigCreator struct{}

func (p *pmd7ConfigCreator) CreateConfig(toolsConfigDir string, patterns []domain.PatternConfiguration) error {
	configString := tools.CreatePmd7Config(patterns)
	err := writeConfigFile(filepath.Join(toolsConfigDir, constants.PMDConfigFileName), []byte(configString))
	if err == nil {
		fmt.Println("PMD7 configuration created based on Codacy settings")
	}
	return err
}

func (p *pmd7ConfigCreator) GetConfigFileName() string { return constants.PMDConfigFileName }
func (p *pmd7ConfigCreator) GetToolName() string       { return "PMD7" }

// pylintConfigCreator implements ToolConfigCreator for Pylint
type pylintConfigCreator struct{}

func (p *pylintConfigCreator) CreateConfig(toolsConfigDir string, patterns []domain.PatternConfiguration) error {
	configString := pylint.GeneratePylintRC(patterns)
	err := writeConfigFile(filepath.Join(toolsConfigDir, constants.PylintConfigFileName), []byte(configString))
	if err == nil {
		fmt.Println("Pylint configuration created based on Codacy settings")
	}
	return err
}

func (p *pylintConfigCreator) GetConfigFileName() string { return constants.PylintConfigFileName }
func (p *pylintConfigCreator) GetToolName() string       { return "Pylint" }

// dartAnalyzerConfigCreator implements ToolConfigCreator for Dart Analyzer
type dartAnalyzerConfigCreator struct{}

func (d *dartAnalyzerConfigCreator) CreateConfig(toolsConfigDir string, patterns []domain.PatternConfiguration) error {
	configString := tools.CreateDartAnalyzerConfig(patterns)
	err := writeConfigFile(filepath.Join(toolsConfigDir, constants.DartAnalyzerConfigFileName), []byte(configString))
	if err == nil {
		fmt.Println("Dart configuration created based on Codacy settings")
	}
	return err
}

func (d *dartAnalyzerConfigCreator) GetConfigFileName() string {
	return constants.DartAnalyzerConfigFileName
}
func (d *dartAnalyzerConfigCreator) GetToolName() string { return "Dart Analyzer" }

// semgrepConfigCreator implements ToolConfigCreator for Semgrep
type semgrepConfigCreator struct{}

func (s *semgrepConfigCreator) CreateConfig(toolsConfigDir string, patterns []domain.PatternConfiguration) error {
	configData, err := tools.GetSemgrepConfig(patterns)
	if err != nil {
		return fmt.Errorf("failed to create Semgrep config: %v", err)
	}
	err = writeConfigFile(filepath.Join(toolsConfigDir, constants.SemgrepConfigFileName), configData)
	if err == nil {
		fmt.Println("Semgrep configuration created based on Codacy settings")
	}
	return err
}

func (s *semgrepConfigCreator) GetConfigFileName() string { return constants.SemgrepConfigFileName }
func (s *semgrepConfigCreator) GetToolName() string       { return "Semgrep" }

// lizardConfigCreator implements ToolConfigCreator for Lizard
type lizardConfigCreator struct{}

func (l *lizardConfigCreator) CreateConfig(toolsConfigDir string, patterns []domain.PatternConfiguration) error {
	patternDefinitions := make([]domain.PatternDefinition, len(patterns))
	for i, pattern := range patterns {
		patternDefinitions[i] = pattern.PatternDefinition
	}
	err := lizard.CreateLizardConfig(toolsConfigDir, patternDefinitions)
	if err != nil {
		return fmt.Errorf("failed to create Lizard configuration: %w", err)
	}
	fmt.Println("Lizard configuration created based on Codacy settings")
	return nil
}

func (l *lizardConfigCreator) GetConfigFileName() string { return "lizard.json" }
func (l *lizardConfigCreator) GetToolName() string       { return "Lizard" }

// reviveConfigCreator implements ToolConfigCreator for Revive
type reviveConfigCreator struct{}

func (r *reviveConfigCreator) CreateConfig(toolsConfigDir string, patterns []domain.PatternConfiguration) error {
	err := createReviveConfigFile(patterns, toolsConfigDir)
	if err == nil {
		fmt.Println("Revive configuration created based on Codacy settings")
	}
	return err
}

func (r *reviveConfigCreator) GetConfigFileName() string { return "revive.toml" }
func (r *reviveConfigCreator) GetToolName() string       { return "Revive" }

func CreateLanguagesConfigFileLocal(toolsConfigDir string) error {
	// Build tool language configurations from API
	configTools, err := tools.BuildLanguagesConfigFromAPI()
	if err != nil {
		return fmt.Errorf("failed to build languages config from API: %w", err)
	}

	// Create the config structure
	config := domain.LanguagesConfig{
		Tools: configTools,
	}

	// Marshal to YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal languages config to YAML: %w", err)
	}

	return writeConfigFile(filepath.Join(toolsConfigDir, constants.LanguagesConfigFileName), data)
}

func CreateGitIgnoreFile() error {
	gitIgnorePath := filepath.Join(config.Config.LocalCodacyDirectory(), constants.GitIgnoreFileName)
	content := "# Codacy CLI\ntools-configs/\n.gitignore\ncli-config.yaml\nlogs/\n"
	return writeConfigFile(gitIgnorePath, []byte(content))
}

func CreateConfigurationFiles(tools []domain.Tool, cliLocalMode bool) error {
	// Create project config file
	configContent := ConfigFileTemplate(tools)
	if err := writeConfigFile(config.Config.ProjectConfigFile(), []byte(configContent)); err != nil {
		return fmt.Errorf("failed to write project config file: %w", err)
	}

	// Create CLI config file
	cliConfigContent := buildCliConfigContent(cliLocalMode)
	if err := writeConfigFile(config.Config.CliConfigFile(), []byte(cliConfigContent)); err != nil {
		return fmt.Errorf("failed to write CLI config file: %w", err)
	}

	return nil
}

// buildCliConfigContent creates the CLI configuration content
func buildCliConfigContent(cliLocalMode bool) string {
	mode := "remote"
	if cliLocalMode {
		mode = "local"
	}
	return fmt.Sprintf("mode: %s", mode)
}

// RuntimePluginConfig holds the structure of the runtime plugin.yaml file
type RuntimePluginConfig struct {
	Name           string `yaml:"name"`
	Description    string `yaml:"description"`
	DefaultVersion string `yaml:"default_version"`
}

func ConfigFileTemplate(tools []domain.Tool) string {
	toolsMap := make(map[string]bool)
	toolVersions := make(map[string]string)
	neededRuntimes := make(map[string]bool)

	toolsWithLatestVersion, _, _ := KeepToolsWithLatestVersion(tools)

	// Get versions and runtime dependencies
	defaultVersions := plugins.GetToolVersions()
	runtimeVersions := plugins.GetRuntimeVersions()
	runtimeDependencies := plugins.GetToolRuntimeDependencies()

	// Process enabled tools
	for _, tool := range toolsWithLatestVersion {
		toolsMap[tool.Uuid] = true
		toolVersions[tool.Uuid] = getToolVersion(tool, defaultVersions)
		addRequiredRuntime(tool.Uuid, neededRuntimes, runtimeDependencies)
	}

	var sb strings.Builder

	// Build runtimes section
	buildRuntimesSection(&sb, tools, neededRuntimes, runtimeVersions, runtimeDependencies)

	// Build tools section
	buildToolsSection(&sb, tools, toolsMap, toolVersions, defaultVersions)

	return sb.String()
}

// getToolVersion returns the version for a tool, preferring tool.Version over default
func getToolVersion(tool domain.Tool, defaultVersions map[string]string) string {
	if tool.Version != "" {
		return tool.Version
	}
	if meta, ok := domain.SupportedToolsMetadata[tool.Uuid]; ok {
		if defaultVersion, ok := defaultVersions[meta.Name]; ok {
			return defaultVersion
		}
	}
	return ""
}

// addRequiredRuntime adds the runtime requirement for a tool
func addRequiredRuntime(toolUuid string, neededRuntimes map[string]bool, runtimeDependencies map[string]string) {
	if meta, ok := domain.SupportedToolsMetadata[toolUuid]; ok {
		if runtime, ok := runtimeDependencies[meta.Name]; ok {
			if meta.Name == "dartanalyzer" {
				// For dartanalyzer, default to dart runtime
				neededRuntimes["dart"] = true
			} else {
				neededRuntimes[runtime] = true
			}
		}
	}
}

// buildRuntimesSection builds the runtimes section of the configuration
func buildRuntimesSection(sb *strings.Builder, tools []domain.Tool, neededRuntimes map[string]bool, runtimeVersions map[string]string, runtimeDependencies map[string]string) {
	sb.WriteString("runtimes:\n")

	if len(tools) == 0 {
		// In local mode with no tools specified, include all necessary runtimes
		addAllSupportedRuntimes(neededRuntimes, runtimeDependencies)
	}

	writeRuntimesList(sb, neededRuntimes, runtimeVersions)
}

// addAllSupportedRuntimes adds all runtimes needed by supported tools
func addAllSupportedRuntimes(neededRuntimes map[string]bool, runtimeDependencies map[string]string) {
	supportedTools, err := plugins.GetSupportedTools()
	if err != nil {
		log.Printf("Warning: failed to get supported tools: %v", err)
		return
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
}

// writeRuntimesList writes the sorted runtimes list to the string builder
func writeRuntimesList(sb *strings.Builder, neededRuntimes map[string]bool, runtimeVersions map[string]string) {
	var sortedRuntimes []string
	for runtime := range neededRuntimes {
		sortedRuntimes = append(sortedRuntimes, runtime)
	}
	sort.Strings(sortedRuntimes)

	for _, runtime := range sortedRuntimes {
		sb.WriteString(fmt.Sprintf("    - %s@%s\n", runtime, runtimeVersions[runtime]))
	}
}

// buildToolsSection builds the tools section of the configuration
func buildToolsSection(sb *strings.Builder, tools []domain.Tool, toolsMap map[string]bool, toolVersions map[string]string, defaultVersions map[string]string) {
	sb.WriteString("tools:\n")

	if len(tools) > 0 {
		writeEnabledTools(sb, toolsMap, toolVersions)
	} else {
		writeAllSupportedTools(sb, defaultVersions)
	}
}

// writeEnabledTools writes the enabled tools to the string builder
func writeEnabledTools(sb *strings.Builder, toolsMap map[string]bool, toolVersions map[string]string) {
	var sortedTools []string
	for uuid, meta := range domain.SupportedToolsMetadata {
		if toolsMap[uuid] {
			sortedTools = append(sortedTools, meta.Name)
		}
	}
	sort.Strings(sortedTools)

	for _, name := range sortedTools {
		for uuid, meta := range domain.SupportedToolsMetadata {
			if meta.Name == name && toolsMap[uuid] {
				version := toolVersions[uuid]
				sb.WriteString(fmt.Sprintf("    - %s@%s\n", name, version))
				break
			}
		}
	}
}

// writeAllSupportedTools writes all supported tools to the string builder
func writeAllSupportedTools(sb *strings.Builder, defaultVersions map[string]string) {
	supportedTools, err := plugins.GetSupportedTools()
	if err != nil {
		log.Printf("Warning: failed to get supported tools: %v", err)
		return
	}

	var sortedTools []string
	for toolName := range supportedTools {
		if version, ok := defaultVersions[toolName]; ok && version != "" {
			sortedTools = append(sortedTools, toolName)
		}
	}
	sort.Strings(sortedTools)

	for _, toolName := range sortedTools {
		if version, ok := defaultVersions[toolName]; ok {
			sb.WriteString(fmt.Sprintf("    - %s@%s\n", toolName, version))
		}
	}
}

func BuildRepositoryConfigurationFiles(flags domain.InitFlags) error {
	fmt.Println("Fetching repository configuration from codacy ...")

	toolsConfigDir := config.Config.ToolsConfigDirectory()

	// Create tools-configs directory if it doesn't exist
	if err := os.MkdirAll(toolsConfigDir, constants.DefaultDirPerms); err != nil {
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

	logVersionConflicts(familyToVersions, toolsWithLatestVersion)

	// Create main config files with all enabled API tools (including cli-config.yaml)
	if err := CreateConfigurationFiles(toolsWithLatestVersion, false); err != nil {
		return err
	}

	// Generate languages configuration based on API tools response (after cli-config.yaml is created)
	if err := tools.CreateLanguagesConfigFile(toolsWithLatestVersion, toolsConfigDir, uuidToName, flags); err != nil {
		return fmt.Errorf("failed to create languages configuration file: %w", err)
	}

	// Filter out any tools that use configuration file
	configuredToolsWithUI := tools.FilterToolsByConfigUsage(toolsWithLatestVersion)

	// Generate config files for tools not using their own config file
	return createToolConfigurationFiles(configuredToolsWithUI, flags)
}

// logVersionConflicts logs warnings about multiple versions of the same tool family
func logVersionConflicts(familyToVersions map[string][]string, toolsWithLatestVersion []domain.Tool) {
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
}

// createToolConfigurationFiles creates configuration files for the given tools
func createToolConfigurationFiles(tools []domain.Tool, flags domain.InitFlags) error {
	for _, tool := range tools {
		apiToolConfigurations, err := codacyclient.GetRepositoryToolPatterns(flags, tool.Uuid)
		if err != nil {
			fmt.Println("Error unmarshaling tool configurations:", err)
			return err
		}

		if err := createToolFileConfiguration(tool, apiToolConfigurations); err != nil {
			return err
		}
	}
	return nil
}

// CreateToolConfigurationFile creates a configuration file for a single tool
func CreateToolConfigurationFile(toolName string, flags domain.InitFlags) error {
	// Find the tool UUID by tool name
	toolUuid := getToolUuidByName(toolName)
	if toolUuid == "" {
		return fmt.Errorf("tool '%s' not found in supported tools", toolName)
	}

	patternsConfig, err := codacyclient.GetDefaultToolPatternsConfig(flags, toolUuid)
	if err != nil {
		return fmt.Errorf("failed to get default patterns: %w", err)
	}

	// Get the tool object to pass to createToolFileConfiguration
	tool := domain.Tool{Uuid: toolUuid}
	return createToolFileConfiguration(tool, patternsConfig)
}

// getToolUuidByName finds the UUID for a tool given its name
func getToolUuidByName(toolName string) string {
	for uuid, toolInfo := range domain.SupportedToolsMetadata {
		if toolInfo.Name == toolName {
			return uuid
		}
	}
	return ""
}

// createToolFileConfiguration creates a configuration file for a single tool using the registry
func createToolFileConfiguration(tool domain.Tool, patternConfiguration []domain.PatternConfiguration) error {
	creator, exists := toolConfigRegistry[tool.Uuid]
	if !exists {
		// Tool doesn't have a configuration creator - this is not an error
		return nil
	}

	toolsConfigDir := config.Config.ToolsConfigDirectory()

	// Ensure the tools-configs directory exists
	if err := os.MkdirAll(toolsConfigDir, constants.DefaultDirPerms); err != nil {
		return fmt.Errorf("failed to create tools-configs directory: %w", err)
	}

	return creator.CreateConfig(toolsConfigDir, patternConfiguration)
}

// CleanConfigDirectory removes all previous configuration files in the tools-configs directory
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

func createReviveConfigFile(config []domain.PatternConfiguration, toolsConfigDir string) error {
	reviveConfigurationString := reviveTool.GenerateReviveConfig(config)
	return os.WriteFile(filepath.Join(toolsConfigDir, "revive.toml"), []byte(reviveConfigurationString), utils.DefaultFilePerms)
}

// BuildDefaultConfigurationFiles creates default configuration files for all tools
func BuildDefaultConfigurationFiles(toolsConfigDir string, flags domain.InitFlags) error {
	// Get default tool versions to determine correct UUIDs
	defaultVersions := plugins.GetToolVersions()

	// Get unique tool names from metadata
	toolNames := make(map[string]struct{})
	for _, meta := range domain.SupportedToolsMetadata {
		toolNames[meta.Name] = struct{}{}
	}

	// Convert tool names to correct UUIDs based on versions
	var allUUIDs []string
	for toolName := range toolNames {
		uuid := selectCorrectToolUUID(toolName, defaultVersions)
		if uuid != "" {
			allUUIDs = append(allUUIDs, uuid)
		}
	}

	return createToolConfigurationsForUUIDs(allUUIDs, toolsConfigDir, flags)
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

// CreateConfigurationFilesForDiscoveredTools creates tool configuration files for discovered tools
func CreateConfigurationFilesForDiscoveredTools(discoveredToolNames map[string]struct{}, toolsConfigDir string, initFlags domain.InitFlags) error {
	// Determine CLI mode
	currentCliMode, err := config.Config.GetCliMode()
	if err != nil {
		log.Printf("Warning: Could not determine CLI mode: %v. Assuming local mode for tool configuration creation.", err)
		currentCliMode = "local" // Default to local
	}

	if currentCliMode == "remote" && initFlags.ApiToken != "" {
		// Remote mode - create configurations based on cloud repository settings
		return createRemoteToolConfigurationsForDiscovered(discoveredToolNames, initFlags)
	} else {
		// Local mode - create default configurations for discovered tools
		return createDefaultConfigurationsForSpecificTools(discoveredToolNames, toolsConfigDir, initFlags)
	}
}

// createRemoteToolConfigurationsForDiscovered creates tool configurations for remote mode based on cloud settings
func createRemoteToolConfigurationsForDiscovered(discoveredToolNames map[string]struct{}, initFlags domain.InitFlags) error {
	// Get repository tools from API
	apiTools, err := tools.GetRepositoryTools(initFlags)
	if err != nil {
		return fmt.Errorf("failed to get repository tools from cloud: %w", err)
	}

	// Filter to only tools that were discovered and enabled in cloud
	var enabledDiscoveredTools []domain.Tool
	for _, tool := range apiTools {
		if tool.Settings.Enabled {
			if meta, ok := domain.SupportedToolsMetadata[tool.Uuid]; ok {
				if _, discovered := discoveredToolNames[meta.Name]; discovered {
					enabledDiscoveredTools = append(enabledDiscoveredTools, tool)
				}
			}
		}
	}

	if len(enabledDiscoveredTools) == 0 {
		fmt.Println("No discovered tools are enabled in cloud configuration.")
		return nil
	}

	// Filter out tools that use their own configuration files
	configuredTools := tools.FilterToolsByConfigUsage(enabledDiscoveredTools)

	fmt.Printf("Creating configurations for %d discovered tools enabled in cloud...\n", len(configuredTools))

	// Create configuration files for each tool using existing logic
	return createToolConfigurationFiles(configuredTools, initFlags)
}

// selectCorrectToolUUID selects the correct UUID for a tool based on its version
func selectCorrectToolUUID(toolName string, defaultVersions map[string]string) string {
	version, hasVersion := defaultVersions[toolName]

	// Special case for PMD: choose PMD7 UUID for version 7.x, PMD UUID for version 6.x
	if toolName == "pmd" && hasVersion {
		if strings.HasPrefix(version, "7.") {
			return domain.PMD7
		} else {
			return domain.PMD
		}
	}

	// Special case for ESLint: choose ESLint9 UUID for version 9.x, ESLint UUID for older versions
	if toolName == "eslint" && hasVersion {
		if strings.HasPrefix(version, "9.") {
			return domain.ESLint9
		} else {
			return domain.ESLint
		}
	}

	// For other tools, find the first matching UUID
	for uuid, meta := range domain.SupportedToolsMetadata {
		if meta.Name == toolName {
			return uuid
		}
	}

	return ""
}

// createDefaultConfigurationsForSpecificTools creates default configurations for specific tools only
func createDefaultConfigurationsForSpecificTools(discoveredToolNames map[string]struct{}, toolsConfigDir string, initFlags domain.InitFlags) error {
	fmt.Printf("Creating default configurations for %d discovered tools...\n", len(discoveredToolNames))

	// Get default tool versions to determine correct UUIDs
	defaultVersions := plugins.GetToolVersions()

	// Convert tool names to UUIDs, selecting the correct UUID based on version
	var discoveredUUIDs []string
	for toolName := range discoveredToolNames {
		uuid := selectCorrectToolUUID(toolName, defaultVersions)
		if uuid != "" {
			discoveredUUIDs = append(discoveredUUIDs, uuid)
		}
	}

	if len(discoveredUUIDs) == 0 {
		log.Printf("Warning: No recognized tools found among discovered tools")
		return nil
	}

	// Create configurations for discovered tools only
	return createToolConfigurationsForUUIDs(discoveredUUIDs, toolsConfigDir, initFlags)
}

// createToolConfigurationsForUUIDs creates tool configurations for specific UUIDs
func createToolConfigurationsForUUIDs(uuids []string, toolsConfigDir string, initFlags domain.InitFlags) error {
	for _, uuid := range uuids {
		patternsConfig, err := codacyclient.GetDefaultToolPatternsConfig(initFlags, uuid)
		if err != nil {
			logToolConfigWarning(uuid, "Failed to get default patterns", err)
			continue
		}

		if err := createSingleToolConfiguration(uuid, patternsConfig, toolsConfigDir); err != nil {
			logToolConfigWarning(uuid, "Failed to create configuration", err)
			continue
		}

		// Print success message
		if meta, ok := domain.SupportedToolsMetadata[uuid]; ok {
			fmt.Printf("Created %s configuration\n", meta.Name)
		}
	}

	return nil
}

// logToolConfigWarning logs a warning message for tool configuration issues
func logToolConfigWarning(uuid, message string, err error) {
	if meta, ok := domain.SupportedToolsMetadata[uuid]; ok {
		log.Printf("Warning: %s for %s: %v", message, meta.Name, err)
	} else {
		log.Printf("Warning: %s for UUID %s: %v", message, uuid, err)
	}
}

// createSingleToolConfiguration creates a single tool configuration file based on UUID using the registry
func createSingleToolConfiguration(uuid string, patternsConfig []domain.PatternConfiguration, toolsConfigDir string) error {
	creator, exists := toolConfigRegistry[uuid]
	if !exists {
		if meta, ok := domain.SupportedToolsMetadata[uuid]; ok {
			return fmt.Errorf("configuration creation not implemented for tool %s", meta.Name)
		}
		return fmt.Errorf("configuration creation not implemented for UUID %s", uuid)
	}

	return creator.CreateConfig(toolsConfigDir, patternsConfig)
}
