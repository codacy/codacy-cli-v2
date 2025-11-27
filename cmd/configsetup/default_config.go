// Package configsetup contains defaults and helpers to generate
// configuration for supported tools.
package configsetup

import (
	"fmt"
	"log"
	"strings"

	codacyclient "codacy/cli-v2/codacy-client"
	"codacy/cli-v2/config"
	"codacy/cli-v2/domain"
	"codacy/cli-v2/plugins"
	"codacy/cli-v2/tools"
)

// KeepToolsWithLatestVersion filters the tools to keep only the latest
// version of each tool family.
func KeepToolsWithLatestVersion(tools []domain.Tool) (
	toolsWithLatestVersion []domain.Tool,
	uuidToName map[string]string,
	familyToVersions map[string][]string,
) {
	latestTools := map[string]domain.Tool{}
	uuidToName = map[string]string{}
	seen := map[string][]domain.Tool{}

	for _, tool := range tools {
		processToolForLatest(tool, latestTools, uuidToName, seen)
	}

	familyToVersions = buildFamilyVersionMap(seen)

	for _, tool := range latestTools {
		toolsWithLatestVersion = append(toolsWithLatestVersion, tool)
	}

	return
}

// processToolForLatest updates the latest tool per family and tracking maps.
func processToolForLatest(tool domain.Tool, latestTools map[string]domain.Tool, uuidToName map[string]string, seen map[string][]domain.Tool) {
	meta, ok := domain.SupportedToolsMetadata[tool.Uuid]
	if !ok {
		return
	}

	seen[meta.Name] = append(seen[meta.Name], tool)

	current, exists := latestTools[meta.Name]
	if !exists || domain.SupportedToolsMetadata[current.Uuid].Priority > meta.Priority {
		latestTools[meta.Name] = tool
		uuidToName[tool.Uuid] = meta.Name
	}
}

// buildFamilyVersionMap builds a map of tool family to discovered versions.
func buildFamilyVersionMap(seen map[string][]domain.Tool) map[string][]string {
	familyToVersions := make(map[string][]string)
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
	return familyToVersions
}

// BuildDefaultConfigurationFiles creates default configuration files for all tools.
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

// CreateConfigurationFilesForDiscoveredTools creates tool configuration files for discovered tools.
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
	}
	// Local mode - create default configurations for discovered tools
	return createDefaultConfigurationsForSpecificTools(discoveredToolNames, toolsConfigDir, initFlags)
}

// createRemoteToolConfigurationsForDiscovered creates tool configurations for remote mode based on cloud settings.
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

// selectCorrectToolUUID selects the correct UUID for a tool based on its version.
func selectCorrectToolUUID(toolName string, defaultVersions map[string]string) string {
	version := defaultVersions[toolName]

	switch toolName {
	case "pmd":
		if strings.HasPrefix(version, "7.") {
			return domain.PMD7
		}
		return domain.PMD
	case "eslint":
		if strings.HasPrefix(version, "9.") {
			return domain.ESLint9
		}
		return domain.ESLint
	}

	// For other tools, find the first matching UUID
	for uuid, meta := range domain.SupportedToolsMetadata {
		if meta.Name == toolName {
			return uuid
		}
	}

	return ""
}

// createDefaultConfigurationsForSpecificTools creates default configurations for specific tools only.
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
		patternsConfig, err := codacyclient.GetToolPatternsConfig(initFlags, uuid, true)
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
