package tools

import (
	"codacy/cli-v2/domain"
	"os"
	"path/filepath"
	"strings"
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

func TestCreateEslintConfigWithPlugins(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "eslint-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Set up the tools config directory
	configDir := filepath.Join(tmpDir, ".codacy", "tools-configs")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create tools config directory: %v", err)
	}

	tests := []struct {
		name           string
		patterns       []domain.PatternConfiguration
		expectedOutput []string
	}{
		{
			name: "Basic ESLint config with plugins",
			patterns: []domain.PatternConfiguration{
				{
					PatternDefinition: domain.PatternDefinition{
						Id: "ESLint8_jest_no-unsanitized_rule1",
					},
				},
			},
			expectedOutput: []string{
				"eslint-plugin-jest@^28.11.0",
				"eslint-plugin-no-unsanitized@^4.0.2",
			},
		},
		{
			name: "Scoped ESLint config with plugins",
			patterns: []domain.PatternConfiguration{
				{
					PatternDefinition: domain.PatternDefinition{
						Id: "ESLint8_@angular-eslint_rule1",
					},
				},
			},
			expectedOutput: []string{
				"@angular-eslint/eslint-plugin@^17.4.1",
			},
		},
		{
			name:           "Empty config",
			patterns:       []domain.PatternConfiguration{},
			expectedOutput: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run the function
			plugins := CreateEslintConfigWithPlugins(tt.patterns)

			// Compare the output
			if len(plugins) != len(tt.expectedOutput) {
				t.Errorf("Expected %d plugins, got %d", len(tt.expectedOutput), len(plugins))
				return
			}

			for i, plugin := range plugins {
				// Remove quotes if present
				plugin = strings.Trim(plugin, "\"")
				if plugin != tt.expectedOutput[i] {
					t.Errorf("Plugin %d: expected %q, got %q", i, tt.expectedOutput[i], plugin)
				}
			}
		})
	}
}
