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
