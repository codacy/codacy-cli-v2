package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	codacyclient "codacy/cli-v2/codacy-client"
	"codacy/cli-v2/domain"
	"codacy/cli-v2/utils"

	"gopkg.in/yaml.v3"
)

// CreateLanguagesConfigFile creates languages-config.yaml based on API response
func CreateLanguagesConfigFile(apiTools []domain.Tool, toolsConfigDir string, toolIDMap map[string]string, initFlags domain.InitFlags) error {
	// Map tool names to their language/extension information
	toolLanguageMap := map[string]domain.ToolLanguageInfo{
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
		"lizard": {
			Name:       "lizard",
			Languages:  []string{"C", "CPP", "Java", "C#", "JavaScript", "TypeScript", "VueJS", "Objective-C", "Swift", "Python", "Ruby", "TTCN-3", "PHP", "Scala", "GDScript", "Golang", "Lua", "Rust", "Fortran", "Kotlin", "Solidity", "Erlang", "Zig", "Perl"},
			Extensions: []string{".c", ".cpp", ".cc", ".h", ".hpp", ".java", ".cs", ".js", ".jsx", ".ts", ".tsx", ".vue", ".m", ".swift", ".py", ".rb", ".ttcn", ".php", ".scala", ".gd", ".go", ".lua", ".rs", ".f", ".f90", ".kt", ".sol", ".erl", ".zig", ".pl"},
		},
		"semgrep": {
			Name:       "semgrep",
			Languages:  []string{"C", "CPP", "C#", "Generic", "Go", "Java", "JavaScript", "JSON", "Kotlin", "Python", "TypeScript", "Ruby", "Rust", "JSX", "PHP", "Scala", "Swift", "Terraform"},
			Extensions: []string{".c", ".cpp", ".h", ".hpp", ".cs", ".go", ".java", ".js", ".json", ".kt", ".py", ".ts", ".rb", ".rs", ".jsx", ".php", ".scala", ".swift", ".tf", ".tfvars"},
		},
		"pyrefly": {
			Name:       "pyrefly",
			Languages:  []string{"Python"},
			Extensions: []string{".py"},
		},
	}

	// Build a list of tool language info for enabled tools
	var configTools []domain.ToolLanguageInfo

	repositoryLanguages, err := getRepositoryLanguages(initFlags)
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
			var filteredExtensionsSet = make(map[string]struct{})
			for _, lang := range langInfo.Languages {
				lowerLang := strings.ToLower(lang)
				if extensions, exists := repositoryLanguages[lowerLang]; exists && len(extensions) > 0 {
					filteredLanguages = append(filteredLanguages, lang)
					for _, ext := range extensions {
						filteredExtensionsSet[ext] = struct{}{}
					}
				}
			}
			filteredExtensions := make([]string, 0, len(filteredExtensionsSet))
			for ext := range filteredExtensionsSet {
				filteredExtensions = append(filteredExtensions, ext)
			}
			slices.Sort(filteredExtensions)
			langInfo.Languages = filteredLanguages
			langInfo.Extensions = filteredExtensions

			// Only add tool if it has languages that exist in the repository
			if len(filteredLanguages) > 0 {
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
	if err := os.WriteFile(configPath, data, utils.DefaultFilePerms); err != nil {
		return fmt.Errorf("failed to write languages config file: %w", err)
	}

	fmt.Println("Created languages configuration file based on enabled tools")
	return nil
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
