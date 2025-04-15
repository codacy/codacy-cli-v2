package tools

import (
	"codacy/cli-v2/config"
	"codacy/cli-v2/domain"
	"codacy/cli-v2/utils"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func quoteWhenIsNotJson(value string) string {
	var data interface{}
	err := json.Unmarshal([]byte(value), &data)
	if err == nil {
		// the value was a json value
		return value
	} else {
		// the value was a string literal.
		return "\"" + value + "\""
	}
}

func CreateEslintConfig(configuration []domain.PatternConfiguration) string {
	result := `export default [
    {
        rules: {
`

	for _, patternConfiguration := range configuration {
		rule := strings.TrimPrefix(patternConfiguration.PatternDefinition.Id, "ESLint8_")

		const tempstring = "TEMPORARYSTRING"
		rule = strings.ReplaceAll(rule, "__", tempstring)
		rule = strings.ReplaceAll(rule, "_", "/")
		rule = strings.ReplaceAll(rule, tempstring, "_")

		parametersString := ""

		for _, parameter := range patternConfiguration.Parameters {
			if parameter.Name == "unnamedParam" {
				parametersString += quoteWhenIsNotJson(parameter.Value)
			}
		}

		// build named parameters json object
		namedParametersString := ""
		for _, parameter := range patternConfiguration.Parameters {

			if parameter.Name != "unnamedParam" {
				if len(namedParametersString) == 0 {
					namedParametersString += "{"
				} else {
					namedParametersString += ", "
				}
				namedParametersString += fmt.Sprintf("\"%s\": %s", parameter.Name, quoteWhenIsNotJson(parameter.Value))
			}
		}
		if len(namedParametersString) > 0 {
			namedParametersString += "}"
		}

		if parametersString != "" && namedParametersString != "" {
			parametersString = fmt.Sprintf("%s, %s", parametersString, namedParametersString)
		} else {
			parametersString += namedParametersString
		}

		result += "          "

		if parametersString == "" {
			result += fmt.Sprintf(`"%s": "error",`, rule)
			result += "\n"
		} else {
			result += fmt.Sprintf(`"%s": ["error", %s],`, rule, parametersString)
			result += "\n"
		}
	}

	result += `        }
    }
];`

	return result
}

func CreateEslintConfigWithPlugins(patterns []domain.PatternConfiguration) []string {
	pluginMap := make(map[string]bool)

	for _, pattern := range patterns {
		parts := strings.Split(pattern.PatternDefinition.Id, "_")
		// Skip first part (ESLint8) and last part (rule name)
		// Only look at middle parts that could be plugin names
		for i := 1; i < len(parts)-1; i++ {
			if parts[i] != "" {
				pluginMap[parts[i]] = true
			}
		}
	}

	// Convert unique plugins from map to slice
	plugins := make([]string, 0, len(pluginMap))
	// Create YAML file content
	yamlContent := "plugins:\n"
	for plugin := range pluginMap {
		var fullPluginName string
		if strings.HasPrefix(plugin, "@") {
			// For scoped packages
			fullPluginName = plugin + "/eslint-plugin"
		} else {
			// Add eslint-plugin- prefix
			fullPluginName = "eslint-plugin-" + plugin
		}
		version := utils.EslintPlugins[fullPluginName]
		pluginWithVersion := fmt.Sprintf("\"%s@%s\"", fullPluginName, version)
		plugins = append(plugins, pluginWithVersion)
		yamlContent += fmt.Sprintf("  - %s\n", pluginWithVersion)
	}

	// Ensure the directory exists
	if err := os.MkdirAll(config.Config.ToolsConfigDirectory(), 0755); err != nil {
		log.Printf("Error creating tools config directory: %v", err)
		return plugins
	}

	// Write YAML content to file with proper permissions
	filePath := filepath.Join(config.Config.ToolsConfigDirectory(), "eslint_plugins.yaml")
	if err := os.WriteFile(filePath, []byte(yamlContent), 0644); err != nil {
		log.Printf("Error writing eslint_plugins.yaml: %v", err)
		return plugins
	}

	// Verify the file was written correctly
	if _, err := os.Stat(filePath); err != nil {
		log.Printf("Error verifying eslint_plugins.yaml: %v", err)
		return plugins
	}

	return plugins
}
