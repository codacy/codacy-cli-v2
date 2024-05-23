package tools

import (
	"encoding/json"
	"fmt"
)

type patternParameterConfiguration struct {
	name  string
	value string
}

func (c *patternParameterConfiguration) Name() string {
	return c.name
}

func (c *patternParameterConfiguration) Value() string {
	return c.value
}

type patternConfiguration struct {
	patternId               string
	paramenterConfiguration []patternParameterConfiguration
}

func (c *patternConfiguration) PatternId() string {
	return c.patternId
}
func (c *patternConfiguration) ParamenterConfiguration() []patternParameterConfiguration {
	return c.paramenterConfiguration
}

type toolConfiguration struct {
	patternsConfiguration []patternConfiguration
}

func (c *toolConfiguration) PatternsConfiguration() []patternConfiguration {
	return c.patternsConfiguration
}

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

func CreateEslintConfig(configuration toolConfiguration) string {
	result := `export default [
    {
        rules: {
`

	for _, patternConfiguration := range configuration.PatternsConfiguration() {
		parametersString := ""

		for _, parameter := range patternConfiguration.paramenterConfiguration {
			if parameter.name == "unnamedParam" {
				parametersString += quoteWhenIsNotJson(parameter.value)
			}
		}

		// build named parameters json object
		namedParametersString := ""
		for _, parameter := range patternConfiguration.paramenterConfiguration {

			if parameter.name != "unnamedParam" {
				if len(namedParametersString) == 0 {
					namedParametersString += "{"
				} else {
					namedParametersString += ", "
				}
				namedParametersString += fmt.Sprintf("\"%s\": %s", parameter.name, quoteWhenIsNotJson(parameter.value))
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
			result += fmt.Sprintf(`"%s": "error",`, patternConfiguration.patternId)
			result += "\n"
		} else {
			result += fmt.Sprintf(`"%s": ["error", %s],`, patternConfiguration.patternId, parametersString)
			result += "\n"
		}
	}

	result += `        }
    }
];`

	return result
}
