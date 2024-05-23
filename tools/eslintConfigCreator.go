package tools

import "fmt"

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

func CreateEslintConfig(configuration toolConfiguration) string {
	result := `export default [
    {
        rules: {
`

	for _, patternConfiguration := range configuration.PatternsConfiguration() {
		var unnamedParam *string

		for _, parameter := range patternConfiguration.paramenterConfiguration {
			if parameter.name == "unnamedParam" {
				unnamedParam = &parameter.value
			}
		}

		result += "          "

		if unnamedParam == nil {
			result += fmt.Sprintf(`"%s": "error",`, patternConfiguration.patternId)
			result += "\n"
		} else {
			result += fmt.Sprintf(`"%s": ["error", "%s"],`, patternConfiguration.patternId, *unnamedParam)
			result += "\n"
		}
	}

	result += `        }
    }
];`

	return result
}
