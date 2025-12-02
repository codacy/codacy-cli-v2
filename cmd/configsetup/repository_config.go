// Package configsetup provides functions to build repository configuration
// files based on Codacy settings.
package configsetup

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	codacyclient "codacy/cli-v2/codacy-client"
	"codacy/cli-v2/config"
	"codacy/cli-v2/constants"
	"codacy/cli-v2/domain"
	"codacy/cli-v2/tools"
)

// BuildRepositoryConfigurationFiles downloads repository configuration from
// Codacy and generates local configuration files.
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
	if err := CreateConfigurationFiles(toolsWithLatestVersion, false, flags); err != nil {
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
	toolUUID := getToolUUIDByName(toolName)
	if toolUUID == "" {
		return fmt.Errorf("tool '%s' not found in supported tools", toolName)
	}

	patternsConfig, err := codacyclient.GetToolPatternsConfig(flags, toolUUID, true)
	if err != nil {
		return fmt.Errorf("failed to get default patterns: %w", err)
	}

	// Get the tool object to pass to createToolFileConfiguration
	tool := domain.Tool{Uuid: toolUUID}
	return createToolFileConfiguration(tool, patternsConfig)
}

// getToolUUIDByName finds the UUID for a tool given its name.
func getToolUUIDByName(toolName string) string {
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
