package tools

import (
	"codacy/cli-v2/domain"
	"encoding/json"
	"fmt"
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

func CreateEslintConfig(toolsConfigDir string, configuration []domain.PatternConfiguration) error {
	result := `export default [
    {
        rules: {
`

	for _, patternConfiguration := range configuration {
		rule := strings.TrimPrefix(patternConfiguration.PatternDefinition.Id, "ESLint8_")

		// Handle plugin rules (those containing _)
		if strings.Contains(rule, "_") {
			parts := strings.Split(rule, "_")
			if len(parts) >= 2 {
				// First part is the plugin name, last part is the rule name
				plugin := parts[0]
				ruleName := parts[len(parts)-1]

				// Handle scoped packages
				if strings.HasPrefix(plugin, "@") {
					rule = fmt.Sprintf("%s/%s", plugin, ruleName)
				} else {
					// For non-scoped packages, add eslint-plugin- prefix
					rule = fmt.Sprintf("%s/%s", "eslint-plugin-"+plugin, ruleName)
				}
			}
		}

		// Skip any rule that contains a plugin (contains "/")
		if strings.Contains(rule, "/") {
			continue
		}

		parametersString := ""

		// Find default value for unnamedParam if needed
		defaultUnnamedParamValue := ""
		for _, paramDef := range patternConfiguration.PatternDefinition.Parameters {
			if paramDef.Name == "unnamedParam" {
				defaultUnnamedParamValue = paramDef.Default
				break
			}
		}

		// Process parameters
		foundUnnamedParam := false
		for _, parameter := range patternConfiguration.Parameters {
			if parameter.Name == "unnamedParam" {
				foundUnnamedParam = true
				// If value is empty but we have a default, use the default
				if parameter.Value == "" && defaultUnnamedParamValue != "" {
					parametersString += quoteWhenIsNotJson(defaultUnnamedParamValue)
				} else if parameter.Value != "" {
					parametersString += quoteWhenIsNotJson(parameter.Value)
				}
			}
		}

		// If we found an unnamed param with empty value but have a default, use it
		if foundUnnamedParam && parametersString == "" && defaultUnnamedParamValue != "" {
			parametersString += quoteWhenIsNotJson(defaultUnnamedParamValue)
		}

		// build named parameters json object
		namedParametersString := ""
		for _, parameter := range patternConfiguration.Parameters {
			if parameter.Name != "unnamedParam" {
				paramValue := parameter.Value

				// If value is empty, look for default in pattern definition
				if paramValue == "" {
					for _, paramDef := range patternConfiguration.PatternDefinition.Parameters {
						if paramDef.Name == parameter.Name && paramDef.Default != "" {
							paramValue = paramDef.Default
							break
						}
					}
				}

				// Skip only if both value and default are empty
				if paramValue == "" {
					continue
				}

				if len(namedParametersString) == 0 {
					namedParametersString += "{"
				} else {
					namedParametersString += ", "
				}

				if paramValue == "true" || paramValue == "false" {
					namedParametersString += fmt.Sprintf("\"%s\": %s", parameter.Name, paramValue)
				} else {
					namedParametersString += fmt.Sprintf("\"%s\": %s", parameter.Name, quoteWhenIsNotJson(paramValue))
				}
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
			result += fmt.Sprintf(`"%s": ["error"],`, rule)
			result += "\n"
		} else {
			result += fmt.Sprintf(`"%s": ["error", %s],`, rule, parametersString)
			result += "\n"
		}
	}

	result += `        }
    }
];`
	eslintConfigFile, err := os.Create(filepath.Join(toolsConfigDir, "eslint.config.mjs"))
	if err != nil {
		return fmt.Errorf("failed to create eslint config file: %v", err)
	}
	defer eslintConfigFile.Close()

	_, err = eslintConfigFile.WriteString(result)
	if err != nil {
		return fmt.Errorf("failed to write eslint config: %v", err)
	}

	return nil
}
