package tools

import (
	"codacy/cli-v2/domain"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func testConfig(t *testing.T, configuration []domain.PatternConfiguration, expected string) {
	// Create a temporary directory for the config file
	tempDir, err := os.MkdirTemp("", "eslint-config-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Call the function with the temporary directory
	err = CreateEslintConfig(tempDir, configuration)
	assert.NoError(t, err)

	// Read the generated file
	configPath := filepath.Join(tempDir, "eslint.config.mjs")
	content, err := os.ReadFile(configPath)
	assert.NoError(t, err)

	// Compare the content
	assert.Equal(t, expected, string(content))
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
          "semi": ["error"],
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
					Id: "ESLint8_consistent-return",
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
					Id: "ESLint8_consistent-return",
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

func TestCreateEslintConfigDoNotSupportPlugins(t *testing.T) {
	testConfig(t,
		[]domain.PatternConfiguration{
			{
				PatternDefinition: domain.PatternDefinition{
					Id: "ESLint8_plugin_consistent-return",
				},
			},
		},
		`export default [
    {
        rules: {
        }
    }
];`)
}

func TestCreateEslintConfigWithDefaultValues(t *testing.T) {
	testConfig(t,
		[]domain.PatternConfiguration{
			{
				PatternDefinition: domain.PatternDefinition{
					Id: "ESLint8_no-fallthrough",
					Parameters: []domain.ParameterConfiguration{
						{
							Name:    "commentPattern",
							Default: "",
						},
						{
							Name:    "allowEmptyCase",
							Default: "false",
						},
					},
				},
				Parameters: []domain.ParameterConfiguration{
					{
						Name:  "commentPattern",
						Value: "", // Empty value with empty default - should be skipped
					},
					{
						Name:  "allowEmptyCase",
						Value: "", // Empty value with default "false" - should use default
					},
				},
			},
		},
		`export default [
    {
        rules: {
          "no-fallthrough": ["error", {"allowEmptyCase": false}],
        }
    }
];`)
}

func TestCreateEslintConfigWithUnnamedDefaultValues(t *testing.T) {
	testConfig(t,
		[]domain.PatternConfiguration{
			{
				PatternDefinition: domain.PatternDefinition{
					Id: "ESLint8_no-inner-declarations",
					Parameters: []domain.ParameterConfiguration{
						{
							Name:    "unnamedParam",
							Default: "functions",
						},
					},
				},
				Parameters: []domain.ParameterConfiguration{
					{
						Name:  "unnamedParam",
						Value: "", // Empty value with default "functions" - should use default
					},
				},
			},
		},
		`export default [
    {
        rules: {
          "no-inner-declarations": ["error", "functions"],
        }
    }
];`)
}
