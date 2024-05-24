package tools

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func testConfig(t *testing.T, configuration ToolConfiguration, expected string) {
	actual := CreateEslintConfig(configuration)
	assert.Equal(t, expected, actual)
}

func TestCreateEslintConfigEmptyConfig(t *testing.T) {
	testConfig(t,
		ToolConfiguration{},
		`export default [
    {
        rules: {
        }
    }
];`)
}

func TestCreateEslintConfigConfig1(t *testing.T) {
	testConfig(t,
		ToolConfiguration{
			PatternsConfiguration: []PatternConfiguration{
				{
					PatternId: "ESLint8_semi",
				},
			},
		},
		`export default [
    {
        rules: {
          "semi": "error",
        }
    }
];`)
}

func TestCreateEslintConfigUnnamedParam(t *testing.T) {
	testConfig(t,
		ToolConfiguration{
			PatternsConfiguration: []PatternConfiguration{
				{
					PatternId: "ESLint8_semi",
					ParamenterConfigurations: []PatternParameterConfiguration{
						{
							Name:  "unnamedParam",
							Value: "never",
						},
					},
				},
			},
		},
		`export default [
    {
        rules: {
          "semi": ["error", "never"],
        }
    }
];`)
}

func TestCreateEslintConfigNamedParam(t *testing.T) {
	testConfig(t,
		ToolConfiguration{
			patternsConfiguration: []PatternConfiguration{
				{
					patternId: "consistent-return",
					paramenterConfiguration: []PatternParameterConfiguration{
						{
							name:  "treatUndefinedAsUnspecified",
							value: "false",
						},
					},
				},
			},
		},
		`export default [
    {
        rules: {
          "consistent-return": ["error", {"treatUndefinedAsUnspecified": false}],
        }
    }
];`)
}

func TestCreateEslintConfigUnnamedAndNamedParam(t *testing.T) {
	testConfig(t,
		ToolConfiguration{
			PatternsConfiguration: []PatternConfiguration{
				{
					PatternId: "consistent-return",
					ParamenterConfigurations: []PatternParameterConfiguration{
						{
							Name:  "treatUndefinedAsUnspecified",
							Value: "false",
						},
						{
							Name:  "unnamedParam",
							Value: "foo",
						},
					},
				},
			},
		},
		`export default [
    {
        rules: {
          "consistent-return": ["error", "foo", {"treatUndefinedAsUnspecified": false}],
        }
    }
];`)
}
