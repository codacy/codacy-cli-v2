package tools

import (
	"codacy/cli-v2/domain"
	"testing"

	"github.com/stretchr/testify/assert"
)

func testConfig(t *testing.T, configuration []domain.PatternConfiguration, expected string) {
	actual := CreateEslintConfig(configuration)
	assert.Equal(t, expected, actual)
}

func TestCreateEslintConfigEmptyConfig(t *testing.T) {
	testConfig(t,
		[]domain.PatternConfiguration{},
		`export default [
    {
        rules: {
        }
    }
];`)
}

func TestCreateEslintConfigConfig1(t *testing.T) {
	testConfig(t,
		[]domain.PatternConfiguration{
			{
				PatternDefinition: domain.PatternDefinition{
					Id: "ESLint8_semi",
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
		[]domain.PatternConfiguration{
			{
				PatternDefinition: domain.PatternDefinition{
					Id: "ESLint8_semi",
				},
				Parameters: []domain.ParameterConfiguration{
					{
						Name:  "unnamedParam",
						Value: "never",
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
		[]domain.PatternConfiguration{
			{
				PatternDefinition: domain.PatternDefinition{
					Id: "consistent-return",
				},
				Parameters: []domain.ParameterConfiguration{
					{
						Name:  "treatUndefinedAsUnspecified",
						Value: "false",
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
		[]domain.PatternConfiguration{
			{
				PatternDefinition: domain.PatternDefinition{
					Id: "consistent-return",
				},
				Parameters: []domain.ParameterConfiguration{
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
		[]domain.PatternConfiguration{
			{
				PatternDefinition: domain.PatternDefinition{
					Id: "plugin/consistent-return",
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
