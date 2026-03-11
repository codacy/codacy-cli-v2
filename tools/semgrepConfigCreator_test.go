package tools

import (
	"codacy/cli-v2/domain"
	"codacy/cli-v2/plugins/tools/opengrep/embedded"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

// TestFilterRulesFromFile tests the FilterRulesFromFile function
func TestFilterRulesFromFile(t *testing.T) {
	// Get the actual rules file content
	rulesData := embedded.GetOpengrepRules()

	// Test case 1: Filter with enabled rules
	config := []domain.PatternConfiguration{
		{
			Enabled: true,
			PatternDefinition: domain.PatternDefinition{
				Id:      "Opengrep_ai.csharp.detect-openai.detect-openai",
				Enabled: true,
			},
		},
	}

	result, err := FilterRulesFromFile(rulesData, config)
	assert.NoError(t, err)

	// Parse the result and verify we got filtered rules
	var parsedRules opengrepRulesFile
	err = yaml.Unmarshal(result, &parsedRules)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(parsedRules.Rules))

	// Test case 2: No enabled rules should return an error
	noEnabledConfig := []domain.PatternConfiguration{
		{
			Enabled: false,
			PatternDefinition: domain.PatternDefinition{
				Id:      "Opengrep_nonexistent",
				Enabled: false,
			},
		},
	}

	_, err = FilterRulesFromFile(rulesData, noEnabledConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no matching rules found")

	// Test case 3: Invalid YAML should return an error
	_, err = FilterRulesFromFile([]byte("invalid yaml"), config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse rules file")
}

// TestGetOpengrepConfig tests the GetOpengrepConfig function
func TestGetOpengrepConfig(t *testing.T) {
	// Test with valid configuration
	config := []domain.PatternConfiguration{
		{
			Enabled: true,
			PatternDefinition: domain.PatternDefinition{
				Id:      "Opengrep_ai.csharp.detect-openai.detect-openai",
				Enabled: true,
			},
		},
	}

	result, err := GetOpengrepConfig(config)
	assert.NoError(t, err)

	var parsedRules opengrepRulesFile
	err = yaml.Unmarshal(result, &parsedRules)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(parsedRules.Rules))

	// Test with empty configuration (should return all rules)
	result, err = GetOpengrepConfig([]domain.PatternConfiguration{})
	assert.NoError(t, err)
	err = yaml.Unmarshal(result, &parsedRules)
	assert.NoError(t, err)
	assert.True(t, len(parsedRules.Rules) > 0)
}

// TestGetDefaultOpengrepConfig tests the GetDefaultOpengrepConfig function
func TestGetDefaultOpengrepConfig(t *testing.T) {
	// Test getting default config
	result, err := GetDefaultOpengrepConfig()
	assert.NoError(t, err)

	var parsedRules opengrepRulesFile
	err = yaml.Unmarshal(result, &parsedRules)
	assert.NoError(t, err)
	assert.True(t, len(parsedRules.Rules) > 0)
}
