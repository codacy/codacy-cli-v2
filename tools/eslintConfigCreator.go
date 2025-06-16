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

func removePlugins(rule string) (string, bool) {
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
		return rule, true
	}

	return rule, false
}

func buildNamedParametersString(parameters []domain.ParameterConfiguration, patternDefinition domain.PatternDefinition) string {
	namedParametersString := ""
	for _, parameter := range parameters {
		if parameter.Name != "unnamedParam" {
			paramValue := parameter.Value

			// If value is empty, look for default in pattern definition
			if paramValue == "" {
				for _, paramDef := range patternDefinition.Parameters {
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
	return namedParametersString
}

func writeFile(filePath string, content string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		return fmt.Errorf("failed to write content: %v", err)
	}

	return nil
}

// rulesWithoutOptions contains ESLint rules that don't accept any configuration options
var rulesWithoutOptions = map[string]bool{
	"no-misleading-character-class": true,
	"constructor-super":             true,
	"for-direction":                 true,
	"no-async-promise-executor":     true,
	"no-case-declarations":          true,
	"no-class-assign":               true,
	"no-compare-neg-zero":           true,
	"no-const-assign":               true,
	"no-control-regex":              true,
	"no-debugger":                   true,
	"no-delete-var":                 true,
	"no-dupe-args":                  true,
	"no-dupe-class-members":         true,
	"no-dupe-else-if":               true,
	"no-dupe-keys":                  true,
	"no-duplicate-case":             true,
	"no-empty-character-class":      true,
	"no-ex-assign":                  true,
	"no-extra-semi":                 true,
	"no-func-assign":                true,
	"no-global-assign":              true,
	"no-import-assign":              true,
	"no-invalid-regexp":             true,
	"no-loss-of-precision":          true,
	"no-mixed-spaces-and-tabs":      true,
	"no-new-symbol":                 true,
	"no-nonoctal-decimal-escape":    true,
	"no-obj-calls":                  true,
	"no-octal":                      true,
	"no-prototype-builtins":         true,
	"no-regex-spaces":               true,
	"no-setter-return":              true,
	"no-shadow-restricted-names":    true,
	"no-sparse-arrays":              true,
	"no-this-before-super":          true,
	"no-unexpected-multiline":       true,
	"no-unreachable":                true,
	"no-unsafe-finally":             true,
	"no-unused-labels":              true,
	"no-useless-backreference":      true,
	"no-useless-catch":              true,
	"no-useless-escape":             true,
	"no-with":                       true,
	"require-yield":                 true,
}

func CreateEslintConfig(toolsConfigDir string, configuration []domain.PatternConfiguration) error {
	result := `export default [
    {
        rules: {
`

	for _, patternConfiguration := range configuration {
		ruleId := patternConfiguration.PatternDefinition.Id
		ruleId = strings.TrimPrefix(ruleId, "ESLint8_")
		ruleId = strings.TrimPrefix(ruleId, "ESLint9_")
		rule := ruleId

		rule, skipRule := removePlugins(rule)
		if skipRule {
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

		// Check if this rule accepts options
		if rulesWithoutOptions[rule] {
			// Rule doesn't accept options, only use error level
			result += "          "
			result += fmt.Sprintf(`"%s": ["error"],`, rule)
			result += "\n"
		} else {
			// Use the new helper method to build named parameters JSON object
			namedParametersString := buildNamedParametersString(patternConfiguration.Parameters, patternConfiguration.PatternDefinition)

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
	}

	result += `        }
    }
];`

	eslintConfigFilePath := filepath.Join(toolsConfigDir, "eslint.config.mjs")
	return writeFile(eslintConfigFilePath, result)
}
