package tools

import (
	"fmt"
	"strings"
)

func CreateEslintConfig(configuration ToolConfiguration) string {
	result := `export default [
    {
        rules: {
`

	for _, patternConfiguration := range configuration.PatternsConfiguration {

		rule := strings.TrimPrefix(patternConfiguration.PatternId, "ESLint8_")

		const tempstring = "TEMPORARYSTRING"
		rule = strings.ReplaceAll(rule, "__", tempstring)
		rule = strings.ReplaceAll(rule, "_", "/")
		rule = strings.ReplaceAll(rule, tempstring, "_")

		if strings.Contains(rule, "/") {
			continue
		}

		var unnamedParam *string

		for _, parameter := range patternConfiguration.ParamenterConfigurations {
			if parameter.Name == "unnamedParam" {
				unnamedParam = &parameter.Value
			}
		}

		result += "          "

		if unnamedParam == nil {
			result += fmt.Sprintf(`"%s": "error",`, rule)
			result += "\n"
		} else {
			result += fmt.Sprintf(`"%s": ["error", "%s"],`, rule, *unnamedParam)
			result += "\n"
		}
	}

	result += `        }
    }
];`

	return result
}
