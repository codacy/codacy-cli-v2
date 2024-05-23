package tools

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func testConfig(t *testing.T, configuration toolConfiguration, expected string) {
	actual := CreateEslintConfig(configuration)
	assert.Equal(t, expected, actual)
}

func TestCreateEslintConfigEmptyConfig(t *testing.T) {
	testConfig(t,
		toolConfiguration{},
		`export default [
    {
        rules: {
        }
    }
];`)
}

func TestCreateEslintConfigConfig1(t *testing.T) {
	testConfig(t,
		toolConfiguration{
			patternsConfiguration: []patternConfiguration{
				{
					patternId: "semi",
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
		toolConfiguration{
			patternsConfiguration: []patternConfiguration{
				{
					patternId: "semi",
					paramenterConfiguration: []patternParameterConfiguration{
						{
							name:  "unnamedParam",
							value: "never",
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
		toolConfiguration{
			patternsConfiguration: []patternConfiguration{
				{
					patternId: "consistent-return",
					paramenterConfiguration: []patternParameterConfiguration{
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
		toolConfiguration{
			patternsConfiguration: []patternConfiguration{
				{
					patternId: "consistent-return",
					paramenterConfiguration: []patternParameterConfiguration{
						{
							name:  "treatUndefinedAsUnspecified",
							value: "false",
						},
						{
							name:  "unnamedParam",
							value: "foo",
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
