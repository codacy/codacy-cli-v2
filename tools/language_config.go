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

//
// This file is responsible for building the languages-config.yaml file.
//

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

	// Create map of language name to files
	languageFilesMap := make(map[string][]string)
	for _, langTool := range languageTools {
		languageFilesMap[strings.ToLower(langTool.Name)] = langTool.Files
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

		// Build extensions and files from API language data
		extensionsSet := make(map[string]struct{})
		filesSet := make(map[string]struct{})

		for _, apiLang := range tool.Languages {
			lowerLang := strings.ToLower(apiLang)
			if extensions, exists := languageExtensionsMap[lowerLang]; exists {
				for _, ext := range extensions {
					extensionsSet[ext] = struct{}{}
				}
			}
			if files, exists := languageFilesMap[lowerLang]; exists {
				for _, file := range files {
					filesSet[file] = struct{}{}
				}
			}
		}

		// Convert sets to sorted slices
		for ext := range extensionsSet {
			configTool.Extensions = append(configTool.Extensions, ext)
		}
		slices.Sort(configTool.Extensions)
		for file := range filesSet {
			configTool.Files = append(configTool.Files, file)
		}
		slices.Sort(configTool.Files)

		// Sort languages alphabetically
		slices.Sort(configTool.Languages)

		result[toolName] = configTool
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
			Files:      []string{},
		}

		// Use only languages that exist in the repository
		extensionsSet := make(map[string]struct{})
		filesSet := make(map[string]struct{})

		for _, lang := range tool.Languages {
			lowerLang := strings.ToLower(lang)
			if repoLang, exists := repositoryLanguages[lowerLang]; exists {
				// Check if this language has either extensions or files
				hasExtensions := len(repoLang.Extensions) > 0
				hasFiles := len(repoLang.Files) > 0

				if hasExtensions || hasFiles {
					configTool.Languages = append(configTool.Languages, lang)

					// Add repository-specific extensions if they exist
					if hasExtensions {
						for _, ext := range repoLang.Extensions {
							extensionsSet[ext] = struct{}{}
						}
					}

					// Add repository-specific files if they exist
					if hasFiles {
						for _, file := range repoLang.Files {
							filesSet[file] = struct{}{}
						}
					}
				}
			}
		}

		// Convert extensions set to sorted slice
		for ext := range extensionsSet {
			configTool.Extensions = append(configTool.Extensions, ext)
		}
		slices.Sort(configTool.Extensions)

		// Convert files set to sorted slice
		for file := range filesSet {
			configTool.Files = append(configTool.Files, file)
		}
		slices.Sort(configTool.Files)

		// Sort languages alphabetically
		slices.Sort(configTool.Languages)

		// Add the tool (even if it has no languages - this is what repository configured)
		configTools = append(configTools, configTool)
	}

	return configTools, nil
}

func getRepositoryLanguages(initFlags domain.InitFlags) (map[string]domain.Language, error) {
	response, err := codacyclient.GetRepositoryLanguages(initFlags)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository languages: %w", err)
	}

	// Create map to store language name -> Language struct
	result := make(map[string]domain.Language)

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

			// Combine and deduplicate files
			files := make(map[string]struct{})
			for _, file := range lang.DefaultFiles {
				files[file] = struct{}{}
			}

			// Convert extension map to slice
			extSlice := make([]string, 0, len(extensions))
			for ext := range extensions {
				extSlice = append(extSlice, ext)
			}
			slices.Sort(extSlice)

			// Convert files map to slice
			fileSlice := make([]string, 0, len(files))
			for file := range files {
				fileSlice = append(fileSlice, file)
			}
			slices.Sort(fileSlice)

			// Add to result map with lowercase key for case-insensitive matching
			result[strings.ToLower(lang.Name)] = domain.Language{
				Name:       lang.Name,
				Extensions: extSlice,
				Files:      fileSlice,
			}
		}
	}

	return result, nil
}
