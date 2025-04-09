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
        "id": "PyLintPython3_E0117"
      },
      "parameters": []
    },
    {
      "patternDefinition": {
        "id": "PyLintPython3_E1206"
      },
      "parameters": []
    },
    {
      "patternDefinition": {
        "id": "PyLintPython3_E1302"
      },
      "parameters": []
    },
    {
      "patternDefinition": {
        "id": "PyLintPython3_E0240"
      },
      "parameters": []
    },
    {
      "patternDefinition": {
        "id": "PyLintPython3_E1125"
      },
      "parameters": []
    },
    {
      "patternDefinition": {
        "id": "PyLintPython3_E0710"
      },
      "parameters": []
    },
    {
      "patternDefinition": {
        "id": "PyLintPython3_E0107"
      },
      "parameters": []
    },
    {
      "patternDefinition": {
        "id": "PyLintPython3_E1111"
      },
      "parameters": []
    },
    {
      "patternDefinition": {
        "id": "PyLintPython3_E1304"
      },
      "parameters": []
    },
    {
      "patternDefinition": {
        "id": "PyLintPython3_E1126"
      },
      "parameters": []
    },
    {
      "patternDefinition": {
        "id": "PyLintPython3_E1205"
      },
      "parameters": []
    },
    {
      "patternDefinition": {
        "id": "PyLintPython3_E1300"
      },
      "parameters": []
    },
    {
      "patternDefinition": {
        "id": "PyLintPython3_E0114"
      },
      "parameters": []
    },
    {
      "patternDefinition": {
        "id": "PyLintPython3_E0601"
      },
      "parameters": []
    },
    {
      "patternDefinition": {
        "id": "PyLintPython3_E0106"
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
        "id": "PyLintPython3_E0112"
      },
      "parameters": []
    },
    {
      "patternDefinition": {
        "id": "PyLintPython3_E1120"
      },
      "parameters": []
    },
    {
      "patternDefinition": {
        "id": "PyLintPython3_E0302"
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
        "id": "PyLintPython3_E1127"
      },
      "parameters": []
    },
    {
      "patternDefinition": {
        "id": "PyLintPython3_C0105"
      },
      "parameters": []
    },
    {
      "patternDefinition": {
        "id": "PyLintPython3_E1200"
      },
      "parameters": []
    },
    {
      "patternDefinition": {
        "id": "PyLintPython3_E0105"
      },
      "parameters": []
    },
    {
      "patternDefinition": {
        "id": "PyLintPython3_E0711"
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
        "id": "PyLintPython3_E0202"
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
        "id": "PyLintPython3_E0104"
      },
      "parameters": []
    },
    {
      "patternDefinition": {
        "id": "PyLintPython3_E1132"
      },
      "parameters": []
    },
    {
      "patternDefinition": {
        "id": "PyLintPython3_C0103"
      },
      "parameters": []
    },
    {
      "patternDefinition": {
        "id": "PyLintPython3_E1305"
      },
      "parameters": []
    },
    {
      "patternDefinition": {
        "id": "PyLintPython3_E0203"
      },
      "parameters": []
    },
    {
      "patternDefinition": {
        "id": "PyLintPython3_E1201"
      },
      "parameters": []
    },
    {
      "patternDefinition": {
        "id": "PyLintPython3_E0113"
      },
      "parameters": []
    },
    {
      "patternDefinition": {
        "id": "PyLintPython3_E1303"
      },
      "parameters": []
    },
    {
      "patternDefinition": {
        "id": "PyLintPython3_E1003"
      },
      "parameters": []
    },
    {
      "patternDefinition": {
        "id": "PyLintPython3_C0104"
      },
      "parameters": []
    },
    {
      "patternDefinition": {
        "id": "PyLintPython3_E1301"
      },
      "parameters": []
    },
    {
      "patternDefinition": {
        "id": "PyLintPython3_E1102"
      },
      "parameters": []
    },
    {
      "patternDefinition": {
        "id": "PyLintPython3_E0116"
      },
      "parameters": []
    },
    {
      "patternDefinition": {
        "id": "PyLintPython3_E0712"
      },
      "parameters": []
    },
    {
      "patternDefinition": {
        "id": "PyLintPython3_E0238"
      },
      "parameters": []
    },
    {
      "patternDefinition": {
        "id": "PyLintPython3_E0702"
      },
      "parameters": []
    },
    {
      "patternDefinition": {
        "id": "PyLintPython3_E0704"
      },
      "parameters": []
    },
    {
      "patternDefinition": {
        "id": "PyLintPython3_E0236"
      },
      "parameters": []
    },
    {
      "patternDefinition": {
        "id": "PyLintPython3_E0102"
      },
      "parameters": []
    },
    {
      "patternDefinition": {
        "id": "PyLintPython3_E0301"
      },
      "parameters": []
    },
    {
      "patternDefinition": {
        "id": "PyLintPython3_E0603"
      },
      "parameters": []
    },
    {
      "patternDefinition": {
        "id": "PyLintPython3_E1124"
      },
      "parameters": []
    },
    {
      "patternDefinition": {
        "id": "PyLintPython3_E0604"
      },
      "parameters": []
    },
    {
      "patternDefinition": {
        "id": "PyLintPython3_E0110"
      },
      "parameters": []
    },
    {
      "patternDefinition": {
        "id": "PyLintPython3_E1123"
      },
      "parameters": []
    },
    {
      "patternDefinition": {
        "id": "PyLintPython3_W0122"
      },
      "parameters": []
    },
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
        "id": "PyLintPython3_E0701"
      },
      "parameters": []
    },
    {
      "patternDefinition": {
        "id": "PyLintPython3_E1306"
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
