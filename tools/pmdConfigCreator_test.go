package tools

import (
	"encoding/xml"
	"os"
	"path/filepath"
	"strings"
	"testing"

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
	config := ToolConfiguration{
		PatternsConfiguration: []PatternConfiguration{
			{
				PatternId: "java/codestyle/AtLeastOneConstructor",
				ParameterConfigurations: []PatternParameterConfiguration{
					{
						Name:  "enabled",
						Value: "true",
					},
				},
			},
			{
				PatternId: "java/design/UnusedPrivateField",
				ParameterConfigurations: []PatternParameterConfiguration{
					{
						Name:  "enabled",
						Value: "true",
					},
				},
			},
			{
				PatternId: "java/design/LoosePackageCoupling",
				ParameterConfigurations: []PatternParameterConfiguration{
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
	config := ToolConfiguration{
		PatternsConfiguration: []PatternConfiguration{
			{
				PatternId: "java/codestyle/AtLeastOneConstructor",
				ParameterConfigurations: []PatternParameterConfiguration{
					{
						Name:  "enabled",
						Value: "false",
					},
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

	config := ToolConfiguration{
		PatternsConfiguration: []PatternConfiguration{},
	}

	obtainedConfig := CreatePmdConfig(config)

	assert.Contains(t, obtainedConfig, `name="Default PMD Ruleset"`, "XML should contain the correct ruleset name")
	assert.Contains(t, obtainedConfig, `xmlns="http://pmd.sourceforge.net/ruleset/2.0.0"`, "XML should contain the correct xmlns")
	assert.Contains(t, obtainedConfig, `xsi:schemaLocation="http://pmd.sourceforge.net/ruleset/2.0.0 http://pmd.sourceforge.net/ruleset_2_0_0.xsd"`, "XML should contain the correct schema location")
	assert.Contains(t, obtainedConfig, `<rule ref="category/java/bestpractices.xml/AvoidReassigningParameters"/>`, "XML should contain expected rules")
	assert.Contains(t, obtainedConfig, `<rule ref="category/java/codestyle.xml/ClassNamingConventions">`, "XML should contain rules with properties")
}
