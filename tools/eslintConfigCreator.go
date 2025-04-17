package tools

import (
	"codacy/cli-v2/domain"
	"encoding/json"
	"fmt"
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
