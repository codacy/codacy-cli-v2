package configsetup

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"codacy/cli-v2/domain"
	"codacy/cli-v2/plugins"
)

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
