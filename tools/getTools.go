package tools

import (
	codacyClient "codacy/cli-v2/codacy-client"
	"codacy/cli-v2/domain"
	"codacy/cli-v2/utils/logger"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

func enrichToolsWithVersion(tools []domain.Tool) ([]domain.Tool, error) {
	toolsVersions, err := codacyClient.GetToolsVersions()
	if err != nil {
		return nil, err
	}

	// Create a map of tool UUIDs to versions
	versionMap := make(map[string]string)
	for _, tool := range toolsVersions {
		versionMap[tool.Uuid] = tool.Version
	}

	// Enrich the input tools with versions
	for i, tool := range tools {
		if version, exists := versionMap[tool.Uuid]; exists {
			tools[i].Version = version
		}
	}

	return tools, nil
}

func GetRepositoryTools(initFlags domain.InitFlags) ([]domain.Tool, error) {
	tools, err := codacyClient.GetRepositoryTools(initFlags)
	if err != nil {
		return nil, err
	}

	// Log supported tool UUIDs and names
	supportedTools := make([]map[string]string, 0, len(domain.SupportedToolsMetadata))
	for uuid, meta := range domain.SupportedToolsMetadata {
		supportedTools = append(supportedTools, map[string]string{
			"uuid": uuid,
			"name": meta.Name,
		})
	}
	logger.Info("Supported tools:", logrus.Fields{
		"supported_tools": supportedTools,
	})

	var enabledTools []domain.Tool
	var unsupportedTools []string

	for _, tool := range tools {
		if tool.Settings.Enabled {
			if _, supported := domain.SupportedToolsMetadata[tool.Uuid]; supported {
				enabledTools = append(enabledTools, tool)
			} else {
				unsupportedTools = append(unsupportedTools, tool.Name)
			}
		}
	}

	if len(unsupportedTools) > 0 {
		fmt.Printf("Warning: Some tools are not supported: %s\n", strings.Join(unsupportedTools, ", "))
	}

	return enrichToolsWithVersion(enabledTools)
}

// FilterToolsByConfigUsage filters out tools that use their own configuration files
// Returns only tools that need configuration to be generated for them (UsesConfigurationFile = false)
func FilterToolsByConfigUsage(tools []domain.Tool) []domain.Tool {
	var filtered []domain.Tool
	for _, tool := range tools {
		if !tool.Settings.UsesConfigurationFile {
			filtered = append(filtered, tool)
		} else {
			fmt.Printf("Skipping config generation for %s - configured to use repo's config file\n", tool.Name)
		}
	}
	return filtered
}
