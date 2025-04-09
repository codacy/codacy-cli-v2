package pylint

import (
	"codacy/cli-v2/domain"
	"codacy/cli-v2/tools/pylint"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testData = `{
  "patterns": [
    {
      "patternDefinition": {
        "id": "PyLintPython3_C0301"
      },
      "parameters": [
        {
          "name": "max-line-length",
          "value": "125"
        }
      ]
    },
    {
      "patternDefinition": {
        "id": "PyLintPython3_C0103"
      },
      "parameters": []
    },
    {
      "patternDefinition": {
        "id": "PyLintPython3_C0112"
      },
      "parameters": []
    },
    {
      "patternDefinition": {
        "id": "PyLintPython3_C0114"
      },
      "parameters": []
    },
    {
      "patternDefinition": {
        "id": "PyLintPython3_E0100"
      },
      "parameters": []
    },
    {
      "patternDefinition": {
        "id": "PyLintPython3_E0101"
      },
      "parameters": []
    },
    {
      "patternDefinition": {
        "id": "PyLintPython3_E0102"
      },
      "parameters": []
    }
  ]
}`

type testDataWrapper struct {
	Patterns []domain.PatternConfiguration `json:"patterns"`
}

func TestGeneratePylintRCFromTestData(t *testing.T) {
	var wrapper testDataWrapper
	err := json.Unmarshal([]byte(testData), &wrapper)
	assert.NoError(t, err)

	rcContent := pylint.GeneratePylintRC(wrapper.Patterns)
	assert.NotEmpty(t, rcContent)

	// Check if the content contains the expected sections
	assert.Contains(t, rcContent, "[MASTER]")
	assert.Contains(t, rcContent, "[MESSAGES CONTROL]")
	assert.Contains(t, rcContent, "disable=all")

	// Check if the patterns are enabled
	for _, pattern := range wrapper.Patterns {
		patternID := strings.Split(pattern.PatternDefinition.Id, "_")[1]
		assert.Contains(t, rcContent, patternID)
	}

	// Check if the parameter is set correctly
	assert.Contains(t, rcContent, "max-line-length=125")
}

func TestGeneratePylintRCDefault(t *testing.T) {
	rcContent := pylint.GeneratePylintRCDefault()
	assert.NotEmpty(t, rcContent)

	// Check if the content contains the expected sections
	assert.Contains(t, rcContent, "[MASTER]")
	assert.Contains(t, rcContent, "[MESSAGES CONTROL]")
	assert.Contains(t, rcContent, "disable=all")

	// Check if default patterns are enabled
	for _, patternID := range pylint.DefaultPatterns {
		assert.Contains(t, rcContent, patternID)
	}
}
