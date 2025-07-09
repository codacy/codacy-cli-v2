package tools

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	codacyclient "codacy/cli-v2/codacy-client"
	"codacy/cli-v2/config"
	"codacy/cli-v2/constants"
	"codacy/cli-v2/domain"
	"codacy/cli-v2/utils/logger"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// buildToolLanguageInfoFromAPI builds tool language information from API data
// This is the core shared logic used by both GetToolLanguageMappingFromAPI and buildToolLanguageConfigFromAPI
func buildToolLanguageInfoFromAPI() (map[string]domain.ToolLanguageInfo, error) {
	// Get all tools from API with their languages
	allTools, err := codacyclient.GetToolsVersions()
	if err != nil {
		return nil, fmt.Errorf("failed to get tools from API: %w", err)
	}

	// Get language file extensions from API
	languageTools, err := codacyclient.GetLanguageTools()
	if err != nil {
		return nil, fmt.Errorf("failed to get language tools from API: %w", err)
	}

	// Create map of language name to file extensions
	languageExtensionsMap := make(map[string][]string)
	for _, langTool := range languageTools {
		languageExtensionsMap[strings.ToLower(langTool.Name)] = langTool.FileExtensions
	}

	// Build tool language configurations from API data
	result := make(map[string]domain.ToolLanguageInfo)
	supportedToolNames := make(map[string]bool)

	// Get supported tool names from metadata
	for _, meta := range domain.SupportedToolsMetadata {
		supportedToolNames[meta.Name] = true
	}

	// Group tools by name and keep only supported ones
	toolsByName := make(map[string]domain.Tool)
	for _, tool := range allTools {
		if supportedToolNames[strings.ToLower(tool.ShortName)] {
			// Keep the tool with latest version (first one in the response)
			if _, exists := toolsByName[strings.ToLower(tool.ShortName)]; !exists {
				toolsByName[strings.ToLower(tool.ShortName)] = tool
			}
		}
	}

	// Build configuration for each supported tool
	for toolName, tool := range toolsByName {
		configTool := domain.ToolLanguageInfo{
			Name:       toolName,
			Languages:  tool.Languages,
			Extensions: []string{},
		}

		// Build extensions from API language data
		extensionsSet := make(map[string]struct{})
		for _, apiLang := range tool.Languages {
			lowerLang := strings.ToLower(apiLang)
			if extensions, exists := languageExtensionsMap[lowerLang]; exists {
				for _, ext := range extensions {
					extensionsSet[ext] = struct{}{}
				}
			}
		}

		// Convert set to sorted slice
		for ext := range extensionsSet {
			configTool.Extensions = append(configTool.Extensions, ext)
		}
		slices.Sort(configTool.Extensions)

		// Sort languages alphabetically
		slices.Sort(configTool.Languages)

		result[toolName] = configTool
	}

	// Fallback: Add Pyrefly mapping if not present in API
	if _, ok := result["pyrefly"]; !ok {
		result["pyrefly"] = domain.ToolLanguageInfo{
			Name:       "pyrefly",
			Languages:  []string{"Python"},
			Extensions: []string{".py"},
		}
	}
	return result, nil
}

// GetToolLanguageMappingFromAPI gets the tool language mapping from the public API
//
// TODO: cache this with TTL time
func GetToolLanguageMappingFromAPI() (map[string]domain.ToolLanguageInfo, error) {
	return buildToolLanguageInfoFromAPI()
}

// GetDefaultToolLanguageMapping returns the default mapping of tools to their supported languages and file extensions
// This function now uses the public API instead of hardcoded mappings.
func GetDefaultToolLanguageMapping() map[string]domain.ToolLanguageInfo {
	// Try to get the mapping from API, fallback to hardcoded only if API fails
	apiMapping, err := GetToolLanguageMappingFromAPI()
	if err != nil {
		logger.Error("Failed to get tool language mapping from API", logrus.Fields{
			"error": err,
		})
		// print fatal error and exit
		log.Fatalf("Failed to get tool language mapping from API: %v", err)
	}
	return apiMapping
}

// buildToolLanguageConfigFromAPI builds tool language configuration using only API data
func buildToolLanguageConfigFromAPI() ([]domain.ToolLanguageInfo, error) {
	// Use the shared logic to get tool info map
	toolInfoMap, err := buildToolLanguageInfoFromAPI()
	if err != nil {
		return nil, err
	}

	// Convert map to slice
	var configTools []domain.ToolLanguageInfo
	for _, toolInfo := range toolInfoMap {
		configTools = append(configTools, toolInfo)
	}

	// Sort tools by name for consistent output
	sort.Slice(configTools, func(i, j int) bool {
		return configTools[i].Name < configTools[j].Name
	})

	return configTools, nil
}

// BuildLanguagesConfigFromAPI builds the tool language configuration from API data
func BuildLanguagesConfigFromAPI() ([]domain.ToolLanguageInfo, error) {
	return buildToolLanguageConfigFromAPI()
}

// CreateLanguagesConfigFile creates languages-config.yaml based on API response
func CreateLanguagesConfigFile(apiTools []domain.Tool, toolsConfigDir string, toolIDMap map[string]string, initFlags domain.InitFlags) error {
	// Check if we're in remote mode
	currentCliMode, err := config.Config.GetCliMode()
	if err != nil {
		// If we can't determine mode, default to local behavior
		currentCliMode = "local"
	}
	isRemoteMode := currentCliMode == "remote"

	var configTools []domain.ToolLanguageInfo

	if isRemoteMode {
		// Remote mode: Use GetRepositoryLanguages as the single source of truth
		configTools, err = buildRemoteModeLanguagesConfig(apiTools, toolIDMap, initFlags)
		if err != nil {
			return fmt.Errorf("failed to build remote mode languages config: %w", err)
		}
	} else {
		// Local mode: Show all supported tools
		configTools, err = buildToolLanguageConfigFromAPI()
		if err != nil {
			return fmt.Errorf("failed to build local mode languages config: %w", err)
		}
	}

	// Sort tools by name for consistent output
	sort.Slice(configTools, func(i, j int) bool {
		return configTools[i].Name < configTools[j].Name
	})

	// Create the config structure
	config := domain.LanguagesConfig{
		Tools: configTools,
	}

	// Marshal to YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal languages config to YAML: %w", err)
	}

	// Write the file
	configPath := filepath.Join(toolsConfigDir, "languages-config.yaml")
	if err := os.WriteFile(configPath, data, constants.DefaultFilePerms); err != nil {
		return fmt.Errorf("failed to write languages config file: %w", err)
	}

	fmt.Println("Created languages configuration file based on API data")
	return nil
}

// buildRemoteModeLanguagesConfig builds the languages config for remote mode using repository languages as source of truth
func buildRemoteModeLanguagesConfig(apiTools []domain.Tool, toolIDMap map[string]string, initFlags domain.InitFlags) ([]domain.ToolLanguageInfo, error) {
	// Get language file extensions from API
	languageTools, err := codacyclient.GetLanguageTools()
	if err != nil {
		return nil, fmt.Errorf("failed to get language tools from API: %w", err)
	}

	// Create map of language name to file extensions
	languageExtensionsMap := make(map[string][]string)
	for _, langTool := range languageTools {
		languageExtensionsMap[strings.ToLower(langTool.Name)] = langTool.FileExtensions
	}

	// Get repository languages - this is the single source of truth for remote mode
	repositoryLanguages, err := getRepositoryLanguages(initFlags)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository languages: %w", err)
	}

	var configTools []domain.ToolLanguageInfo

	for _, tool := range apiTools {
		shortName, exists := toolIDMap[tool.Uuid]
		if !exists {
			// Skip tools we don't recognize
			continue
		}

		configTool := domain.ToolLanguageInfo{
			Name:       shortName,
			Languages:  []string{},
			Extensions: []string{},
		}

		// Use only languages that exist in the repository
		extensionsSet := make(map[string]struct{})

		for _, lang := range tool.Languages {
			lowerLang := strings.ToLower(lang)
			if repoExts, exists := repositoryLanguages[lowerLang]; exists && len(repoExts) > 0 {
				configTool.Languages = append(configTool.Languages, lang)
				// Add repository-specific extensions
				for _, ext := range repoExts {
					extensionsSet[ext] = struct{}{}
				}
			}
		}

		// Convert extensions set to sorted slice
		for ext := range extensionsSet {
			configTool.Extensions = append(configTool.Extensions, ext)
		}
		slices.Sort(configTool.Extensions)

		// Sort languages alphabetically
		slices.Sort(configTool.Languages)

		// Add the tool (even if it has no languages - this is what repository configured)
		configTools = append(configTools, configTool)
	}

	return configTools, nil
}

func getRepositoryLanguages(initFlags domain.InitFlags) (map[string][]string, error) {
	response, err := codacyclient.GetRepositoryLanguages(initFlags)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository languages: %w", err)
	}

	// Create map to store language name -> combined extensions
	result := make(map[string][]string)

	// Filter and process languages
	for _, lang := range response {
		if lang.Enabled && lang.Detected {
			// Combine and deduplicate extensions
			extensions := make(map[string]struct{})
			for _, ext := range lang.CodacyDefaults {
				extensions[ext] = struct{}{}
			}
			for _, ext := range lang.Extensions {
				extensions[ext] = struct{}{}
			}

			// Convert map to slice
			extSlice := make([]string, 0, len(extensions))
			for ext := range extensions {
				extSlice = append(extSlice, ext)
			}

			// Sort extensions for consistent ordering in the config file
			slices.Sort(extSlice)

			// Add to result map with lowercase key for case-insensitive matching
			result[strings.ToLower(lang.Name)] = extSlice
		}
	}

	return result, nil
}
