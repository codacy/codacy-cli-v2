package pylint

import (
	"codacy/cli-v2/tools/pylint"
	"codacy/cli-v2/tools/types"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testData = `{
  "uuid": "31677b6d-4ae0-4f56-8041-606a8d7a8e61",
  "isEnabled": true,
  "notEdited": false,
  "patterns": [
    {
      "internalId": "PyLintPython3_E0117",
      "parameters": []
    },
    {
      "internalId": "PyLintPython3_E1206",
      "parameters": []
    },
    {
      "internalId": "PyLintPython3_E1302",
      "parameters": []
    },
    {
      "internalId": "PyLintPython3_E0240",
      "parameters": []
    },
    {
      "internalId": "PyLintPython3_E1125",
      "parameters": []
    },
    {
      "internalId": "PyLintPython3_E0710",
      "parameters": []
    },
    {
      "internalId": "PyLintPython3_E0107",
      "parameters": []
    },
    {
      "internalId": "PyLintPython3_E1111",
      "parameters": []
    },
    {
      "internalId": "PyLintPython3_E1304",
      "parameters": []
    },
    {
      "internalId": "PyLintPython3_E1126",
      "parameters": []
    },
    {
      "internalId": "PyLintPython3_E1205",
      "parameters": []
    },
    {
      "internalId": "PyLintPython3_E1300",
      "parameters": []
    },
    {
      "internalId": "PyLintPython3_E0114",
      "parameters": []
    },
    {
      "internalId": "PyLintPython3_E0601",
      "parameters": []
    },
    {
      "internalId": "PyLintPython3_E0106",
      "parameters": []
    },
    {
      "internalId": "PyLintPython3_E0100",
      "parameters": []
    },
    {
      "internalId": "PyLintPython3_E0112",
      "parameters": []
    },
    {
      "internalId": "PyLintPython3_E1120",
      "parameters": []
    },
    {
      "internalId": "PyLintPython3_E0302",
      "parameters": []
    },
    {
      "internalId": "PyLintPython3_C0114",
      "parameters": []
    },
    {
      "internalId": "PyLintPython3_E1127",
      "parameters": []
    },
    {
      "internalId": "PyLintPython3_C0105",
      "parameters": []
    },
    {
      "internalId": "PyLintPython3_E1200",
      "parameters": []
    },
    {
      "internalId": "PyLintPython3_E0105",
      "parameters": []
    },
    {
      "internalId": "PyLintPython3_E0711",
      "parameters": []
    },
    {
      "internalId": "PyLintPython3_C0112",
      "parameters": []
    },
    {
      "internalId": "PyLintPython3_E0202",
      "parameters": []
    },
    {
      "internalId": "PyLintPython3_E0101",
      "parameters": []
    },
    {
      "internalId": "PyLintPython3_E0104",
      "parameters": []
    },
    {
      "internalId": "PyLintPython3_E1132",
      "parameters": []
    },
    {
      "internalId": "PyLintPython3_C0103",
      "parameters": []
    },
    {
      "internalId": "PyLintPython3_E1305",
      "parameters": []
    },
    {
      "internalId": "PyLintPython3_E0203",
      "parameters": []
    },
    {
      "internalId": "PyLintPython3_E1201",
      "parameters": []
    },
    {
      "internalId": "PyLintPython3_E0113",
      "parameters": []
    },
    {
      "internalId": "PyLintPython3_E1303",
      "parameters": []
    },
    {
      "internalId": "PyLintPython3_E1003",
      "parameters": []
    },
    {
      "internalId": "PyLintPython3_C0104",
      "parameters": []
    },
    {
      "internalId": "PyLintPython3_E1301",
      "parameters": []
    },
    {
      "internalId": "PyLintPython3_E1102",
      "parameters": []
    },
    {
      "internalId": "PyLintPython3_E0116",
      "parameters": []
    },
    {
      "internalId": "PyLintPython3_E0712",
      "parameters": []
    },
    {
      "internalId": "PyLintPython3_E0238",
      "parameters": []
    },
    {
      "internalId": "PyLintPython3_E0702",
      "parameters": []
    },
    {
      "internalId": "PyLintPython3_E0704",
      "parameters": []
    },
    {
      "internalId": "PyLintPython3_E0236",
      "parameters": []
    },
    {
      "internalId": "PyLintPython3_E0102",
      "parameters": []
    },
    {
      "internalId": "PyLintPython3_E0301",
      "parameters": []
    },
    {
      "internalId": "PyLintPython3_E0603",
      "parameters": []
    },
    {
      "internalId": "PyLintPython3_E1124",
      "parameters": []
    },
    {
      "internalId": "PyLintPython3_E0604",
      "parameters": []
    },
    {
      "internalId": "PyLintPython3_E0110",
      "parameters": []
    },
    {
      "internalId": "PyLintPython3_E1123",
      "parameters": []
    },
    {
      "internalId": "PyLintPython3_W0122",
      "parameters": []
    },
    {
      "internalId": "PyLintPython3_C0301",
      "parameters": [
        {
          "name": "max-line-length",
          "value": "125"
        }
      ]
    },
    {
      "internalId": "PyLintPython3_E0701",
      "parameters": []
    },
    {
      "internalId": "PyLintPython3_E1306",
      "parameters": []
    },
    {
      "internalId": "PyLintPython3_E0108",
      "parameters": []
    },
    {
      "internalId": "PyLintPython3_E0239",
      "parameters": []
    },
    {
      "internalId": "PyLintPython3_E0103",
      "parameters": []
    },
    {
      "internalId": "PyLintPython3_E0241",
      "parameters": []
    },
    {
      "internalId": "PyLintPython3_E1121",
      "parameters": []
    },
    {
      "internalId": "PyLintPython3_E0211",
      "parameters": []
    },
    {
      "internalId": "PyLintPython3_E0115",
      "parameters": []
    }
  ]
}`

func TestGeneratePylintRCFromTestData(t *testing.T) {
	var config types.ToolConfiguration
	err := json.Unmarshal([]byte(testData), &config)
	assert.NoError(t, err)

	rcContent := pylint.GeneratePylintRC(config)
	t.Logf("Generated Pylint RC content:\n%s", rcContent)
	assert.NotEmpty(t, rcContent)

	// Verify basic structure
	assert.Contains(t, rcContent, "[MASTER]")
	assert.Contains(t, rcContent, "[MESSAGES CONTROL]")
	assert.Contains(t, rcContent, "enable=")

	// Verify specific pattern with parameters
	assert.Contains(t, rcContent, "max-line-length=125")
}

func TestGeneratePylintRCDefault(t *testing.T) {
	rcContent := pylint.GeneratePylintRCDefault()

	// Print the generated content for inspection
	fmt.Println("Generated Pylint RC Content:")
	fmt.Println(rcContent)

	// Basic validation
	if rcContent == "" {
		t.Error("Generated RC content is empty")
	}

	// Check for required sections
	requiredSections := []string{"[MASTER]", "[MESSAGES CONTROL]"}
	for _, section := range requiredSections {
		if !strings.Contains(rcContent, section) {
			t.Errorf("Missing required section: %s", section)
		}
	}

	// Check if default patterns are enabled
	if !strings.Contains(rcContent, "enable=") {
		t.Error("Missing enable line")
	}

	// Check if any parameters are present
	if !strings.Contains(rcContent, "=") {
		t.Error("No parameters found in the configuration")
	}
}
