package tools

import (
	"codacy/cli-v2/domain"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

// Sample rules YAML content for testing
const sampleRulesYAML = `rules:
  - id: rule1
    pattern: |
      $X
    message: "Test rule 1"
    languages: [go]
    severity: INFO
  - id: rule2
    pattern: |
      $Y
    message: "Test rule 2"
    languages: [javascript]
    severity: WARNING
  - id: rule3
    pattern-either:
      - pattern: "foo()"
      - pattern: "bar()"
    message: "Test rule 3"
    languages: [python]
    severity: ERROR
`

// TestFilterRulesFromFile tests the FilterRulesFromFile function
func TestFilterRulesFromFile(t *testing.T) {
	// Create a temporary rules file
	tempDir := t.TempDir()
	rulesFile := filepath.Join(tempDir, "rules.yaml")
	err := os.WriteFile(rulesFile, []byte(sampleRulesYAML), 0644)
	assert.NoError(t, err)

	// Test case 1: Filter with enabled rules
	config := []domain.PatternConfiguration{
		{
			Enabled: true,
			PatternDefinition: domain.PatternDefinition{
				Id:      "Semgrep_rule1",
				Enabled: true,
			},
		},
		{
			Enabled: true,
			PatternDefinition: domain.PatternDefinition{
				Id:      "Semgrep_rule3",
				Enabled: true,
			},
		},
	}

	result, err := FilterRulesFromFile(rulesFile, config)
	assert.NoError(t, err)

	// Parse the result and verify the rules
	var parsedRules semgrepRulesFile
	err = yaml.Unmarshal(result, &parsedRules)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(parsedRules.Rules))

	// Check that it contains rule1 and rule3 but not rule2
	ruleIDs := map[string]bool{}
	for _, rule := range parsedRules.Rules {
		id, _ := rule["id"].(string)
		ruleIDs[id] = true
	}
	assert.True(t, ruleIDs["rule1"])
	assert.False(t, ruleIDs["rule2"])
	assert.True(t, ruleIDs["rule3"])

	// Test case 2: No enabled rules should return an error
	noEnabledConfig := []domain.PatternConfiguration{
		{
			Enabled: false,
			PatternDefinition: domain.PatternDefinition{
				Id:      "Semgrep_rule1",
				Enabled: false,
			},
		},
	}

	_, err = FilterRulesFromFile(rulesFile, noEnabledConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no matching rules found")

	// Test case 3: Non-existent rules file should return an error
	_, err = FilterRulesFromFile(filepath.Join(tempDir, "nonexistent.yaml"), config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read rules file")
}

// TestGetSemgrepConfig tests the GetSemgrepConfig function
func TestGetSemgrepConfig(t *testing.T) {
	// Create a temporary rules file
	tempDir := t.TempDir()
	testRulesFile := filepath.Join(tempDir, "rules.yaml")
	err := os.WriteFile(testRulesFile, []byte(sampleRulesYAML), 0644)
	assert.NoError(t, err)

	// Create a mock executable path that points to our temp directory
	originalGetExecutablePath := getExecutablePath
	getExecutablePath = func() (string, error) {
		return filepath.Join(tempDir, "test-executable"), nil
	}
	defer func() {
		getExecutablePath = originalGetExecutablePath
	}()

	// Create the plugins directory structure
	pluginsDir := filepath.Join(tempDir, "plugins", "tools", "semgrep")
	err = os.MkdirAll(pluginsDir, 0755)
	assert.NoError(t, err)

	// Copy our test file to the plugins directory
	err = os.WriteFile(filepath.Join(pluginsDir, "rules.yaml"), []byte(sampleRulesYAML), 0644)
	assert.NoError(t, err)

	// Test with valid configuration
	config := []domain.PatternConfiguration{
		{
			Enabled: true,
			PatternDefinition: domain.PatternDefinition{
				Id:      "Semgrep_rule1",
				Enabled: true,
			},
		},
	}

	result, err := GetSemgrepConfig(config)
	assert.NoError(t, err)

	var parsedRules semgrepRulesFile
	err = yaml.Unmarshal(result, &parsedRules)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(parsedRules.Rules))

	// Test with empty configuration
	_, err = GetSemgrepConfig([]domain.PatternConfiguration{})
	assert.Error(t, err)
}

// TestGetDefaultSemgrepConfig tests the GetDefaultSemgrepConfig function
func TestGetDefaultSemgrepConfig(t *testing.T) {
	// Create a temporary rules file
	tempDir := t.TempDir()
	testRulesFile := filepath.Join(tempDir, "rules.yaml")
	err := os.WriteFile(testRulesFile, []byte(sampleRulesYAML), 0644)
	assert.NoError(t, err)

	// Create a mock executable path that points to our temp directory
	originalGetExecutablePath := getExecutablePath
	getExecutablePath = func() (string, error) {
		return filepath.Join(tempDir, "test-executable"), nil
	}
	defer func() {
		getExecutablePath = originalGetExecutablePath
	}()

	// Create the plugins directory structure
	pluginsDir := filepath.Join(tempDir, "plugins", "tools", "semgrep")
	err = os.MkdirAll(pluginsDir, 0755)
	assert.NoError(t, err)

	// Copy our test file to the plugins directory
	err = os.WriteFile(filepath.Join(pluginsDir, "rules.yaml"), []byte(sampleRulesYAML), 0644)
	assert.NoError(t, err)

	// Test getting default config
	result, err := GetDefaultSemgrepConfig()
	assert.NoError(t, err)

	var parsedRules semgrepRulesFile
	err = yaml.Unmarshal(result, &parsedRules)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(parsedRules.Rules))

	// Test when rules.yaml doesn't exist
	os.Remove(filepath.Join(pluginsDir, "rules.yaml"))
	_, err = GetDefaultSemgrepConfig()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rules.yaml not found")
}
