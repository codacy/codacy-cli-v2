package tools

import (
	"encoding/xml"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"codacy/cli-v2/domain"

	"github.com/stretchr/testify/assert"
)

// PMDRuleset represents the structure of a PMD ruleset XML
type PMDRuleset struct {
	XMLName     xml.Name  `xml:"ruleset"`
	Name        string    `xml:"name,attr"`
	Xmlns       string    `xml:"xmlns,attr"`
	XmlnsXsi    string    `xml:"xmlns:xsi,attr"`
	SchemaLoc   string    `xml:"xsi:schemaLocation,attr"`
	Description string    `xml:"description,omitempty"`
	Rules       []PMDRule `xml:"rule"`
}

type PMDRule struct {
	Ref        string         `xml:"ref,attr"`
	Properties *PMDProperties `xml:"properties,omitempty"`
}

type PMDProperties struct {
	Properties []PMDProperty `xml:"property"`
}

type PMDProperty struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

func TestCreatePmdConfig(t *testing.T) {
	// Setup test configuration with patterns
	config := []domain.PatternConfiguration{
		{
			PatternDefinition: domain.PatternDefinition{
				Id: "java/codestyle/AtLeastOneConstructor",
			},
			Parameters: []domain.ParameterConfiguration{
				{
					Name:  "enabled",
					Value: "true",
				},
			},
		},
		{
			PatternDefinition: domain.PatternDefinition{
				Id: "java/design/UnusedPrivateField",
			},
			Parameters: []domain.ParameterConfiguration{
				{
					Name:  "enabled",
					Value: "true",
				},
			},
		},
		{
			PatternDefinition: domain.PatternDefinition{
				Id: "java/design/LoosePackageCoupling",
			},
			Parameters: []domain.ParameterConfiguration{
				{
					Name:  "enabled",
					Value: "true",
				},
				{
					Name:  "packages",
					Value: "java.util,java.io",
				},
			},
		},
	}

	// Generate PMD config
	generatedConfig := CreatePmdConfig(config)

	// Read expected ruleset
	expectedRulesetPath := filepath.Join("testdata", "repositories", "pmd", "expected-ruleset.xml")
	expectedRulesetBytes, err := os.ReadFile(expectedRulesetPath)
	if err != nil {
		t.Fatalf("Failed to read expected ruleset: %v", err)
	}

	// Normalize both configurations by removing whitespace differences
	normalizeXML := func(xml string) string {
		// Remove all whitespace between tags
		xml = strings.ReplaceAll(xml, ">\n", ">")
		xml = strings.ReplaceAll(xml, ">\t", ">")
		xml = strings.ReplaceAll(xml, ">\r", ">")
		xml = strings.ReplaceAll(xml, "\n<", "<")
		xml = strings.ReplaceAll(xml, "\t<", "<")
		xml = strings.ReplaceAll(xml, "\r<", "<")
		// Remove multiple spaces
		xml = strings.Join(strings.Fields(xml), " ")
		return xml
	}

	normalizedGenerated := normalizeXML(generatedConfig)
	normalizedExpected := normalizeXML(string(expectedRulesetBytes))

	// Compare the normalized configurations
	assert.Equal(t, normalizedExpected, normalizedGenerated, "Generated PMD config should match expected ruleset")
}

func TestCreatePmdConfigWithDisabledRules(t *testing.T) {
	config := []domain.PatternConfiguration{
		{
			PatternDefinition: domain.PatternDefinition{
				Id: "java/codestyle/AtLeastOneConstructor",
			},
			Parameters: []domain.ParameterConfiguration{
				{
					Name:  "enabled",
					Value: "false",
				},
			},
		},
	}

	obtainedConfig := CreatePmdConfig(config)

	var ruleset PMDRuleset
	err := xml.Unmarshal([]byte(obtainedConfig), &ruleset)
	if err != nil {
		t.Fatalf("Failed to parse generated XML: %v", err)
	}

	for _, rule := range ruleset.Rules {
		assert.NotContains(t, rule.Ref, "AtLeastOneConstructor", "Disabled rule should not appear in config")
	}
}

func TestCreatePmdConfigEmpty(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	tmpDir := t.TempDir()
	err = os.MkdirAll(filepath.Join(tmpDir, "plugins", "tools", "pmd"), 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory structure: %v", err)
	}

	defaultRuleset := `<?xml version="1.0" encoding="UTF-8"?>
<ruleset name="Default PMD Ruleset"
         xmlns="http://pmd.sourceforge.net/ruleset/2.0.0"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://pmd.sourceforge.net/ruleset/2.0.0 http://pmd.sourceforge.net/ruleset_2_0_0.xsd">
    <!-- Default rules here -->
</ruleset>`

	err = os.WriteFile(filepath.Join(tmpDir, "plugins", "tools", "pmd", "default-ruleset.xml"), []byte(defaultRuleset), 0644)
	if err != nil {
		t.Fatalf("Failed to write test ruleset: %v", err)
	}

	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}
	defer os.Chdir(cwd)

	config := []domain.PatternConfiguration{}

	obtainedConfig := CreatePmdConfig(config)

	assert.Contains(t, obtainedConfig, `name="Default PMD Ruleset"`, "XML should contain the correct ruleset name")
	assert.Contains(t, obtainedConfig, `xmlns="http://pmd.sourceforge.net/ruleset/2.0.0"`, "XML should contain the correct xmlns")
	assert.Contains(t, obtainedConfig, `xsi:schemaLocation="http://pmd.sourceforge.net/ruleset/2.0.0 http://pmd.sourceforge.net/ruleset_2_0_0.xsd"`, "XML should contain the correct schema location")
	assert.Contains(t, obtainedConfig, `<rule ref="category/java/bestpractices.xml/AvoidReassigningParameters"/>`, "XML should contain expected rules")
	assert.Contains(t, obtainedConfig, `<rule ref="category/java/codestyle.xml/ClassNamingConventions">`, "XML should contain rules with properties")
}

func TestCreatePmdConfigEmptyParameterValues(t *testing.T) {
	config := []domain.PatternConfiguration{
		{
			PatternDefinition: domain.PatternDefinition{
				Id: "java/codestyle/ClassNamingConventions",
			},
			Parameters: []domain.ParameterConfiguration{
				{
					Name:  "enabled",
					Value: "true",
				},
				{
					Name:  "testClassPattern",
					Value: "", // Empty value should be skipped
				},
				{
					Name:  "abstractClassPattern",
					Value: "Abstract.*", // Non-empty value should be included
				},
				{
					Name:  "classPattern",
					Value: "", // Empty value should be skipped
				},
			},
		},
	}

	obtainedConfig := CreatePmdConfig(config)

	var ruleset PMDRuleset
	err := xml.Unmarshal([]byte(obtainedConfig), &ruleset)
	if err != nil {
		t.Fatalf("Failed to parse generated XML: %v", err)
	}

	// Find the ClassNamingConventions rule
	var classNamingRule *PMDRule
	for i, rule := range ruleset.Rules {
		if strings.Contains(rule.Ref, "ClassNamingConventions") {
			classNamingRule = &ruleset.Rules[i]
			break
		}
	}

	// Rule should exist
	assert.NotNil(t, classNamingRule, "ClassNamingConventions rule should exist")

	// Properties should exist
	assert.NotNil(t, classNamingRule.Properties, "Properties section should exist")

	// Should only contain the non-empty parameter
	foundAbstractPattern := false
	for _, prop := range classNamingRule.Properties.Properties {
		// Should not find empty value parameters
		assert.NotEqual(t, "testClassPattern", prop.Name, "Empty parameter should be skipped")
		assert.NotEqual(t, "classPattern", prop.Name, "Empty parameter should be skipped")

		// Should find non-empty parameter
		if prop.Name == "abstractClassPattern" {
			foundAbstractPattern = true
			assert.Equal(t, "Abstract.*", prop.Value)
		}
	}

	assert.True(t, foundAbstractPattern, "Non-empty parameter should be included")
}

func TestCreatePmdConfigWithDefaultParameterValues(t *testing.T) {
	config := []domain.PatternConfiguration{
		{
			PatternDefinition: domain.PatternDefinition{
				Id: "PMD_category_pom_errorprone_InvalidDependencyTypes",
				Parameters: []domain.ParameterConfiguration{
					{
						Name:    "validTypes",
						Default: "pom,jar,maven-plugin,ejb,war,ear,rar,par",
					},
				},
			},
			Parameters: []domain.ParameterConfiguration{
				{
					Name:  "enabled",
					Value: "true",
				},
				{
					Name:  "validTypes",
					Value: "", // Empty value should use default
				},
			},
		},
	}

	obtainedConfig := CreatePmdConfig(config)

	var ruleset PMDRuleset
	err := xml.Unmarshal([]byte(obtainedConfig), &ruleset)
	if err != nil {
		t.Fatalf("Failed to parse generated XML: %v", err)
	}

	// Find the InvalidDependencyTypes rule
	var found bool
	for _, rule := range ruleset.Rules {
		if strings.Contains(rule.Ref, "InvalidDependencyTypes") {
			found = true
			// Rule should have properties
			assert.NotNil(t, rule.Properties, "Rule should have properties")

			// Properties should contain validTypes with default value
			foundParam := false
			for _, prop := range rule.Properties.Properties {
				if prop.Name == "validTypes" {
					foundParam = true
					assert.Equal(t, "pom,jar,maven-plugin,ejb,war,ear,rar,par", prop.Value, "Property should use default value")
				}
			}
			assert.True(t, foundParam, "validTypes parameter should be included with default value")
		}
	}
	assert.True(t, found, "InvalidDependencyTypes rule should exist")
}

func TestDefaultParametersIncludedWhenNotSpecified(t *testing.T) {
	config := []domain.PatternConfiguration{
		{
			PatternDefinition: domain.PatternDefinition{
				Id: "PMD_category_ecmascript_codestyle_AssignmentInOperand",
				Parameters: []domain.ParameterConfiguration{
					{
						Name:    "allowWhile",
						Default: "false",
					},
					{
						Name:    "allowIf",
						Default: "false",
					},
					{
						Name:    "allowTernaryResults",
						Default: "false",
					},
					{
						Name:    "allowTernary",
						Default: "false",
					},
					{
						Name:    "allowFor",
						Default: "false",
					},
					{
						Name:    "allowIncrementDecrement",
						Default: "false",
					},
				},
			},
			Parameters: []domain.ParameterConfiguration{
				{
					Name:  "enabled",
					Value: "true",
				},
				// No other parameters specified - should use defaults
			},
		},
	}

	obtainedConfig := CreatePmdConfig(config)

	var ruleset PMDRuleset
	err := xml.Unmarshal([]byte(obtainedConfig), &ruleset)
	if err != nil {
		t.Fatalf("Failed to parse generated XML: %v", err)
	}

	// Find the AssignmentInOperand rule
	var found bool
	for _, rule := range ruleset.Rules {
		if strings.Contains(rule.Ref, "AssignmentInOperand") {
			found = true
			// Rule should have properties
			assert.NotNil(t, rule.Properties, "Rule should have properties")

			// Check all the default parameters are included
			expectedParams := map[string]string{
				"allowWhile":              "false",
				"allowIf":                 "false",
				"allowTernaryResults":     "false",
				"allowTernary":            "false",
				"allowFor":                "false",
				"allowIncrementDecrement": "false",
			}

			for _, prop := range rule.Properties.Properties {
				expectedValue, exists := expectedParams[prop.Name]
				assert.True(t, exists, "Unexpected parameter: %s", prop.Name)
				assert.Equal(t, expectedValue, prop.Value, "Parameter %s has wrong value", prop.Name)
				delete(expectedParams, prop.Name)
			}

			// All expected parameters should have been found
			assert.Empty(t, expectedParams, "Not all default parameters were included")
		}
	}
	assert.True(t, found, "AssignmentInOperand rule should exist")
}
