package eslint

import (
	"testing"

	"codacy/cli-v2/tools"
	"codacy/cli-v2/tools/eslint"

	"github.com/stretchr/testify/assert"
)

func testConfig(t *testing.T, configuration tools.ToolConfiguration, expected string) {
	actual := eslint.CreateEslintConfig(configuration)
	assert.Equal(t, expected, actual)
}

func TestCreateEslintConfigEmptyConfig(t *testing.T) {
	testConfig(t,
		tools.ToolConfiguration{},
		`export default [
    {
        rules: {
        }
    }
];`)
}

func TestCreateEslintConfigConfig1(t *testing.T) {
	testConfig(t,
		tools.ToolConfiguration{
			PatternsConfiguration: []tools.PatternConfiguration{
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
		tools.ToolConfiguration{
			PatternsConfiguration: []tools.PatternConfiguration{
				{
					PatternId: "ESLint8_semi",
					ParameterConfigurations: []tools.PatternParameterConfiguration{
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
		tools.ToolConfiguration{
			PatternsConfiguration: []tools.PatternConfiguration{
				{
					PatternId: "consistent-return",
					ParameterConfigurations: []tools.PatternParameterConfiguration{
						{
							Name:  "treatUndefinedAsUnspecified",
							Value: "false",
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
		tools.ToolConfiguration{
			PatternsConfiguration: []tools.PatternConfiguration{
				{
					PatternId: "consistent-return",
					ParameterConfigurations: []tools.PatternParameterConfiguration{
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

func TestCreateEslintConfigSupportPlugins(t *testing.T) {
	testConfig(t,
		tools.ToolConfiguration{
			PatternsConfiguration: []tools.PatternConfiguration{
				{
					PatternId: "plugin/consistent-return",
				},
			},
		},
		`export default [
    {
        rules: {
          "plugin/consistent-return": "error",
        }
    }
];`)
}
