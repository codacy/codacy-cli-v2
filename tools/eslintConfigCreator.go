package tools

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
		result += "          \""
		result += patternConfiguration.patternId
		result += "\": \"error\",\n"
	}

	result += `        }
    }
];`

	return result
}
