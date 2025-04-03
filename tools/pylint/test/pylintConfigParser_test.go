package pylint

import (
	pylint "codacy/cli-v2/tools/pylint/src"
	"codacy/cli-v2/tools/types"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testJSONData = `{
	"data": [
	{
		"patternDefinition": {
			"id": "PyLintPython3_C0301",
			"category": "CodeStyle",
			"level": "Info",
			"severityLevel": "Info",
			"enabled": false,
			"parameters": [
				{
					"name": "max-line-length",
					"default": "120",
					"description": "Maximum number of characters on a single line."
				}
			],
			"title": "line-too-long (C0301)",
			"description": "Line too long (%s/%s) Used when a line is longer than a given number of characters.",
			"explanation": "# Line Too Long (C0301)\n\nLine too long (%s/%s) Used when a line is longer than a given number of\ncharacters.",
			"languages": [
				"Python"
			],
			"timeToFix": 5
		},
		"enabled": true,
		"isCustom": true,
		"parameters": [
			{
				"name": "max-line-length",
				"value": "145"
			}
		],
		"enabledBy": []
	},
	{
		"patternDefinition": {
			"id": "PyLintPython3_C0209",
			"category": "CodeStyle",
			"level": "Info",
			"severityLevel": "Info",
			"enabled": false,
			"parameters": [],
			"title": "consider-using-f-string (C0209)",
			"description": "Formatting a regular string which could be an f-string Used when we detect a string that is being formatted with format() or % which could potentially be an f-string. The use of f-strings is preferred. Requires Python 3.6 and py-version >= 3.6.",
			"explanation": "# Consider Using F String (C0209)\n\nFormatting a regular string which could be an f-string Used when we\ndetect a string that is being formatted with format() or % which could\npotentially be an f-string. The use of f-strings is preferred. Requires\nPython 3.6 and py-version \\>= 3.6.",
			"languages": [
				"Python"
			],
			"timeToFix": 5
		},
		"enabled": false,
		"isCustom": false,
		"parameters": [],
		"enabledBy": []
	},
	{
		"patternDefinition": {
			"id": "PyLintPython3_C0103",
			"category": "CodeStyle",
			"level": "Info",
			"severityLevel": "Info",
			"enabled": false,
			"parameters": [
				{
					"name": "argument-rgx",
					"default": "[a-z_][a-z0-9_]{2,30}$",
					"description": "Regular expression matching correct argument names. Overrides argument- naming-style."
				},
				{
					"name": "attr-rgx",
					"default": "[a-z_][a-z0-9_]{2,30}$",
					"description": "Regular expression matching correct attribute names."
				},
				{
					"name": "class-attribute-rgx",
					"default": "([A-Za-z_][A-Za-z0-9_]{2,30}|(__.*__))$",
					"description": "Regular expression matching correct class attribute names."
				},
				{
					"name": "class-rgx",
					"default": "[A-Z_][a-zA-Z0-9]+$",
					"description": "Regular expression matching correct class names."
				},
				{
					"name": "const-rgx",
					"default": "(([A-Z_][A-Z0-9_]*)|(__.*__))$",
					"description": "Regular expression matching correct constant names."
				},
				{
					"name": "function-rgx",
					"default": "[a-z_][a-z0-9_]{2,30}$",
					"description": "Regular expression matching correct function names."
				},
				{
					"name": "inlinevar-rgx",
					"default": "[A-Za-z0-9_]*$",
					"description": "Regular expression matching correct inline iteration names."
				},
				{
					"name": "method-rgx",
					"default": "[a-z_][a-z0-9_]{2,30}$",
					"description": "Regular expression matching correct method names."
				},
				{
					"name": "module-rgx",
					"default": "(([a-z_][a-z0-9_]*)|([A-Z][a-zA-Z0-9]+))$",
					"description": "Regular expression matching correct module names."
				},
				{
					"name": "variable-rgx",
					"default": "[a-z_][a-z0-9_]{2,30}$",
					"description": "Regular expression matching correct variable names."
				}
			],
			"title": "invalid-name (C0103)",
			"description": "%s name \"%s\" doesn't conform to %s Used when the name doesn't conform to naming rules associated to its type (constant, variable, class...).",
			"explanation": "# Invalid Name (C0103)\n\n%s name \"%s\" doesn't conform to %s Used when the name doesn't conform to\nnaming rules associated to its type (constant, variable, class...).",
			"languages": [
				"Python"
			],
			"timeToFix": 5
		},
		"enabled": true,
		"isCustom": true,
		"parameters": [],
		"enabledBy": []
	}
	]
}`

var complexTestJSONData = `{
	"data": [
		{
			"patternDefinition": {
				"id": "PyLintPython3_C0301",
				"parameters": [
					{
						"name": "max-line-length",
						"default": "120",
						"description": "Maximum number of characters on a single line."
					}
				]
			},
			"enabled": true,
			"parameters": [
				{
					"name": "max-line-length",
					"value": "145"
				}
			]
		},
		{
			"patternDefinition": {
				"id": "PyLintPython3_C0209",
				"parameters": []
			},
			"enabled": true,
			"parameters": []
		},
		{
			"patternDefinition": {
				"id": "PyLintPython3_R0903",
				"parameters": [
					{
						"name": "max-args",
						"default": "5",
						"description": "Maximum number of arguments for function / method"
					}
				]
			},
			"enabled": true,
			"parameters": []
		},
		{
			"patternDefinition": {
				"id": "PyLintPython3_C0111",
				"parameters": [
					{
						"name": "max-doc-length",
						"default": "100",
						"description": "Maximum number of characters in a docstring"
					}
				]
			},
			"enabled": true,
			"parameters": [
				{
					"name": "max-doc-length",
					"value": "80"
				}
			]
		},
		{
			"patternDefinition": {
				"id": "PyLintPython3_R0904",
				"parameters": []
			},
			"enabled": true,
			"parameters": []
		},
		{
			"patternDefinition": {
				"id": "PyLintPython3_R0912",
				"parameters": [
					{
						"name": "max-branches",
						"default": "12",
						"description": "Maximum number of branches for function / method body"
					}
				]
			},
			"enabled": true,
			"parameters": [
				{
					"name": "max-branches",
					"value": "15"
				}
			]
		},
		{
			"patternDefinition": {
				"id": "PyLintPython3_R0916",
				"parameters": []
			},
			"enabled": true,
			"parameters": []
		},
		{
			"patternDefinition": {
				"id": "PyLintPython3_C0103",
				"parameters": [
					{
						"name": "method-naming-style",
						"default": "any",
						"description": "Regular expression which should only match function or class names that do not require a docstring"
					}
				]
			},
			"enabled": true,
			"parameters": [
				{
					"name": "method-naming-style",
					"value": "snake_case"
				}
			]
		},
		{
			"patternDefinition": {
				"id": "PyLintPython3_C0114",
				"parameters": [
					{
						"name": "docstring-min-length",
						"default": "10",
						"description": "Minimum length of docstring"
					}
				]
			},
			"enabled": true,
			"parameters": [
				{
					"name": "docstring-min-length",
					"value": "20"
				}
			]
		},
		{
			"patternDefinition": {
				"id": "PyLintPython3_C0117",
				"parameters": []
			},
			"enabled": true,
			"parameters": []
		},
		{
			"patternDefinition": {
				"id": "PyLintPython3_C0302",
				"parameters": [
					{
						"name": "max-module-lines",
						"default": "1000",
						"description": "Maximum number of lines in a module"
					}
				]
			},
			"enabled": true,
			"parameters": [
				{
					"name": "max-module-lines",
					"value": "2000"
				}
			]
		},
		{
			"patternDefinition": {
				"id": "PyLintPython3_C0303",
				"parameters": [
					{
						"name": "trailing-whitespace",
						"default": "0",
						"description": "Maximum number of trailing whitespace characters"
					}
				]
			},
			"enabled": true,
			"parameters": []
		},
		{
			"patternDefinition": {
				"id": "PyLintPython3_C0304",
				"parameters": []
			},
			"enabled": true,
			"parameters": []
		},
		{
			"patternDefinition": {
				"id": "PyLintPython3_C0305",
				"parameters": []
			},
			"enabled": true,
			"parameters": []
		},
		{
			"patternDefinition": {
				"id": "PyLintPython3_C0306",
				"parameters": []
			},
			"enabled": true,
			"parameters": []
		}
	]
}`

// =============================================
// Parsing Tests
// =============================================
// These tests verify the parsing functionality of the Pylint configuration:
// - TestParsePylintPatternsFromJSON: Verifies that JSON data is correctly parsed into PylintPatternConfiguration
// - TestExtractPatternID: Verifies that pattern IDs are correctly extracted from full pattern names
// - TestFilterEnabledPatterns: Verifies that only enabled patterns are returned from a list

func TestParsePylintPatternsFromJSON(t *testing.T) {
	patterns, err := pylint.ParsePylintPatternsFromJSON([]byte(testJSONData))
	assert.NoError(t, err)
	assert.Len(t, patterns, 3)

	pattern := patterns[0]
	assert.Equal(t, "C0301", pattern.Id)
	assert.True(t, pattern.Enabled)

	// Fetch the right parameters (since user-defined parameters are preferred)
	assert.Len(t, pattern.Parameters, 1)
	assert.Equal(t, "max-line-length", pattern.Parameters[0].Name)
	assert.Equal(t, "145", pattern.Parameters[0].Value)

	pattern2 := patterns[1]
	assert.Equal(t, "C0209", pattern2.Id)
	assert.False(t, pattern2.Enabled)
	assert.Len(t, pattern2.Parameters, 0)

	pattern3 := patterns[2]
	assert.Equal(t, "C0103", pattern3.Id)
	assert.True(t, pattern3.Enabled)
	assert.Len(t, pattern3.Parameters, 10)
	assert.Equal(t, "argument-rgx", pattern3.Parameters[0].Name)
	assert.Equal(t, "[a-z_][a-z0-9_]{2,30}$", pattern3.Parameters[0].Value)
	assert.Equal(t, "attr-rgx", pattern3.Parameters[1].Name)
	assert.Equal(t, "[a-z_][a-z0-9_]{2,30}$", pattern3.Parameters[1].Value)
	assert.Equal(t, "class-attribute-rgx", pattern3.Parameters[2].Name)
	assert.Equal(t, "([A-Za-z_][A-Za-z0-9_]{2,30}|(__.*__))$", pattern3.Parameters[2].Value)
}

func TestFilterEnabledPatterns(t *testing.T) {
	patterns, err := pylint.ParsePylintPatternsFromJSON([]byte(testJSONData))
	assert.NoError(t, err)

	// Test filtering enabled patterns
	enabledPatterns := pylint.FilterEnabledPatterns(patterns)
	assert.Len(t, enabledPatterns, 2)
	assert.Equal(t, "C0301", enabledPatterns[0].Id)
	assert.True(t, enabledPatterns[0].Enabled)

	// Verify parameters are preserved
	assert.Len(t, enabledPatterns[0].Parameters, 1)
	assert.Equal(t, "max-line-length", enabledPatterns[0].Parameters[0].Name)
	assert.Equal(t, "145", enabledPatterns[0].Parameters[0].Value)

}

func TestGeneratePylintRC(t *testing.T) {
	// First parse the JSON data
	patterns, err := pylint.ParsePylintPatternsFromJSON([]byte(testJSONData))
	assert.NoError(t, err)
	assert.Len(t, patterns, 3)

	// Then filter enabled patterns
	enabledPatterns := pylint.FilterEnabledPatterns(patterns)
	assert.Len(t, enabledPatterns, 2)
	assert.Equal(t, "C0301", enabledPatterns[0].Id)

	// Finally generate the pylintrc content
	rcContentNonEmptyParameters := pylint.GeneratePylintRC(enabledPatterns)

	// Verify the pylintrc content matches the expected format
	expectedContent := `[MASTER]
ignore=CVS
persistent=yes
load-plugins=

[MESSAGES CONTROL]
disable=all
enable=C0301,C0103

[FORMAT]
max-line-length=145

[BASIC]
argument-rgx=[a-z_][a-z0-9_]{2,30}$
attr-rgx=[a-z_][a-z0-9_]{2,30}$
class-attribute-rgx=([A-Za-z_][A-Za-z0-9_]{2,30}|(__.*__))$
class-rgx=[A-Z_][a-zA-Z0-9]+$
const-rgx=(([A-Z_][A-Z0-9_]*)|(__.*__))$
function-rgx=[a-z_][a-z0-9_]{2,30}$
inlinevar-rgx=[A-Za-z0-9_]*$
method-rgx=[a-z_][a-z0-9_]{2,30}$
module-rgx=(([a-z_][a-z0-9_]*)|([A-Z][a-zA-Z0-9]+))$
variable-rgx=[a-z_][a-z0-9_]{2,30}$

`
	assert.Equal(t, expectedContent, rcContentNonEmptyParameters)

	// Test with a pattern that has no parameters
	emptyPatterns := []types.PylintPatternConfiguration{
		{
			Id:         "PyLintPython3_C0209",
			Parameters: []types.PylintPatternParameterConfiguration{},
			Enabled:    true,
		},
	}
	rcContentEmptyParameters := pylint.GeneratePylintRC(emptyPatterns)
	assert.Contains(t, rcContentEmptyParameters, "enable=C0209")       // Single line with all enabled patterns
	assert.NotContains(t, rcContentEmptyParameters, "max-line-length") // Should not include any parameters
}

func TestComplexPatternSet(t *testing.T) {
	// Parse the complex test data
	patterns, err := pylint.ParsePylintPatternsFromJSON([]byte(complexTestJSONData))
	assert.NoError(t, err)
	assert.Len(t, patterns, 15)

	// Filter enabled patterns
	enabledPatterns := pylint.FilterEnabledPatterns(patterns)
	assert.Len(t, enabledPatterns, 15)

	// Generate pylintrc content
	rcContent := pylint.GeneratePylintRC(enabledPatterns)

	t.Logf("Generated pylintrc content:\n%s", rcContent)

	// Verify all enabled patterns are present in a single line
	expectedEnableLine := "enable=C0301,C0209,R0903,C0111,R0904,R0912,R0916,C0103,C0114,C0117,C0302,C0303,C0304,C0305,C0306"
	assert.Contains(t, rcContent, expectedEnableLine)

	// Verify parameters for patterns with user-defined values
	assert.Contains(t, rcContent, "max-line-length=145")               // C0301
	assert.NotContains(t, rcContent, "max-doc-length=80")              //C0111- should not be there since it's not mapped to any section
	assert.Contains(t, rcContent, "max-branches=15")                   // R0912
	assert.NotContains(t, rcContent, "method-naming-style=snake_case") // C0103- should not be there since it's not mapped to any section
	assert.Contains(t, rcContent, "docstring-min-length=20")           // C0114
	assert.Contains(t, rcContent, "max-module-lines=2000")             // C0302

	// Verify parameters for patterns with only default values
	assert.Contains(t, rcContent, "max-args=5")               // R0903
	assert.NotContains(t, rcContent, "trailing-whitespace=0") // C0303- should not be there since it's not mapped to any section

	// Verify patterns without parameters don't have any parameter settings
	assert.NotContains(t, rcContent, "C0209=")
	assert.NotContains(t, rcContent, "R0904=")
	assert.NotContains(t, rcContent, "R0916=")
	assert.NotContains(t, rcContent, "C0117=")
	assert.NotContains(t, rcContent, "C0304=")
	assert.NotContains(t, rcContent, "C0305=")
	assert.NotContains(t, rcContent, "C0306=")
}
