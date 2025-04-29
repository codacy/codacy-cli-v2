package tools

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"codacy/cli-v2/utils"

	"gopkg.in/yaml.v3"
)

const CodacyApiBase = "https://app.codacy.com"

// ToolLanguageInfo contains language and extension information for a tool
type ToolLanguageInfo struct {
	Name       string   `yaml:"name"`
	Languages  []string `yaml:"languages,flow"`
	Extensions []string `yaml:"extensions,flow"`
}

// LanguagesConfig represents the structure of the languages configuration file
type LanguagesConfig struct {
	Tools []ToolLanguageInfo `yaml:"tools"`
}

// CreateLanguagesConfigFile creates languages-config.yaml based on API response
func CreateLanguagesConfigFile(apiTools []Tool, toolsConfigDir string, toolIDMap map[string]string, apiToken string, provider string, organization string, repository string) error {
	// Map tool names to their language/extension information
	toolLanguageMap := map[string]ToolLanguageInfo{
		"cppcheck": {
			Name:       "cppcheck",
			Languages:  []string{"C", "CPP"},
			Extensions: []string{".c", ".cpp", ".cc", ".h", ".hpp"},
		},
		"pylint": {
			Name:       "pylint",
			Languages:  []string{"Python"},
			Extensions: []string{".py"},
		},
		"eslint": {
			Name:       "eslint",
			Languages:  []string{"JavaScript", "TypeScript", "JSX", "TSX"},
			Extensions: []string{".js", ".jsx", ".ts", ".tsx"},
		},
		"pmd": {
			Name:       "pmd",
			Languages:  []string{"Java", "JavaScript", "JSP", "Velocity", "XML", "Apex", "Scala", "Ruby", "VisualForce"},
			Extensions: []string{".java", ".js", ".jsp", ".vm", ".xml", ".cls", ".trigger", ".scala", ".rb", ".page", ".component"},
		},
		"trivy": {
			Name:       "trivy",
			Languages:  []string{"Multiple"},
			Extensions: []string{},
		},
		"dartanalyzer": {
			Name:       "dartanalyzer",
			Languages:  []string{"Dart"},
			Extensions: []string{".dart"},
		},
	}

	// Build a list of tool language info for enabled tools
	var configTools []ToolLanguageInfo

	repositoryLanguages, err := getRepositoryLanguages(apiToken, provider, organization, repository)
	if err != nil {
		return fmt.Errorf("failed to get repository languages: %w", err)
	}

	for _, tool := range apiTools {
		shortName, exists := toolIDMap[tool.Uuid]
		if !exists {
			// Skip tools we don't recognize
			continue
		}

		// Get language info for this tool
		langInfo, exists := toolLanguageMap[shortName]
		if exists {
			// Special case for Trivy - always include it
			if shortName == "trivy" {
				configTools = append(configTools, langInfo)
				continue
			}

			// Filter languages based on repository languages
			var filteredLanguages []string
			for _, lang := range langInfo.Languages {
				// Convert both to lowercase for case-insensitive comparison
				lowerLang := strings.ToLower(lang)
				if extensions, exists := repositoryLanguages[lowerLang]; exists && len(extensions) > 0 {
					filteredLanguages = append(filteredLanguages, lang)
				}
			}

			// Only add tool if it has languages that exist in the repository
			if len(filteredLanguages) > 0 {
				langInfo.Languages = filteredLanguages
				configTools = append(configTools, langInfo)
			}
		}
	}

	// If we have no tools or couldn't match any, include all known tools
	if len(configTools) == 0 {
		for _, langInfo := range toolLanguageMap {
			configTools = append(configTools, langInfo)
		}
	}

	// Create the config structure
	config := LanguagesConfig{
		Tools: configTools,
	}

	// Marshal to YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal languages config to YAML: %w", err)
	}

	// Write the file
	configPath := filepath.Join(toolsConfigDir, "languages-config.yaml")
	if err := os.WriteFile(configPath, data, utils.DefaultFilePerms); err != nil {
		return fmt.Errorf("failed to write languages config file: %w", err)
	}

	fmt.Println("Created languages configuration file based on enabled tools")
	return nil
}

// https://app.codacy.com/api/v3/organizations/gh/troubleshoot-codacy/repositories/eslint-test-examples/settings/languages
func getRepositoryLanguages(apiToken string, provider string, organization string, repository string) (map[string][]string, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	url := fmt.Sprintf("%s/api/v3/organizations/%s/%s/repositories/%s/settings/languages",
		CodacyApiBase,
		provider,
		organization,
		repository)

	// Create a new GET request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set the API token header
	req.Header.Set("api-token", apiToken)

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to get repository languages: status code %d", resp.StatusCode)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Define the response structure
	type LanguageResponse struct {
		Name           string   `json:"name"`
		CodacyDefaults []string `json:"codacyDefaults"`
		Extensions     []string `json:"extensions"`
		Enabled        bool     `json:"enabled"`
		Detected       bool     `json:"detected"`
	}

	type LanguagesResponse struct {
		Languages []LanguageResponse `json:"languages"`
	}

	var response LanguagesResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Create map to store language name -> combined extensions
	result := make(map[string][]string)

	// Filter and process languages
	for _, lang := range response.Languages {
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

			// Add to result map with lowercase key for case-insensitive matching
			result[strings.ToLower(lang.Name)] = extSlice
		}
	}

	return result, nil
}
