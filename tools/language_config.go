package tools

import (
	"fmt"
	"os"
	"path/filepath"

	"codacy/cli-v2/utils"

	"gopkg.in/yaml.v3"
)

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
func CreateLanguagesConfigFile(apiTools []Tool, toolsConfigDir string, toolIDMap map[string]string) error {
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

	for _, tool := range apiTools {
		shortName, exists := toolIDMap[tool.Uuid]
		if !exists {
			// Skip tools we don't recognize
			continue
		}

		// Get language info for this tool
		langInfo, exists := toolLanguageMap[shortName]
		if exists {
			configTools = append(configTools, langInfo)
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
