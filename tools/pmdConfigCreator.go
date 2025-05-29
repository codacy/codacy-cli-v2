package tools

import (
	"codacy/cli-v2/domain"
	_ "embed"
	"encoding/xml"
	"fmt"
	"strings"
)

//go:embed pmd/default-ruleset.xml
var defaultPMDRuleset string

// Parameter represents a rule parameter
type Parameter struct {
	Name  string
	Value string
}

// Rule represents a PMD rule
type Rule struct {
	PatternID  string
	Parameters []Parameter
	Enabled    bool
}

// RuleSet represents the PMD ruleset XML structure
type RuleSet struct {
	XMLName     xml.Name `xml:"ruleset"`
	Name        string   `xml:"name,attr"`
	Description string   `xml:"description"`
	Rules       []Rule
}

// DeprecatedReferences maps deprecated pattern IDs to their new versions
var DeprecatedReferences = map[string]string{}

// Add PMD headers
var (
	pmd6Header = `<?xml version="1.0"?>
<ruleset name="Codacy PMD Ruleset"
    xmlns="http://pmd.sourceforge.net/ruleset/2.0.0"
    xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
    xsi:schemaLocation="http://pmd.sourceforge.net/ruleset/2.0.0 https://pmd.sourceforge.io/ruleset_2_0_0.xsd">
    <description>Codacy PMD Ruleset</description>`

	pmd7Header = `<?xml version="1.0" encoding="UTF-8"?>
<ruleset name="Codacy PMD 7 Ruleset"
    xmlns="https://pmd.github.io/ruleset/2.0.0"
    xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
    xsi:schemaLocation="https://pmd.github.io/ruleset/2.0.0 https://pmd.github.io/schemas/pmd-7.0.0.xsd">
    <description>Codacy PMD 7 Ruleset</description>`
)

// CreatePmd6Config creates a PMD 6 configuration from the provided tool configuration
func CreatePmd6Config(configuration []domain.PatternConfiguration) string {
	return createPmdConfigGeneric(configuration, pmd6Header, convertPatternIDToPMD, false)
}

// CreatePmd7Config creates a PMD 7 configuration from the provided tool configuration
func CreatePmd7Config(configuration []domain.PatternConfiguration) string {
	return createPmdConfigGeneric(configuration, pmd7Header, convertPatternIDToPMD7, true)
}

// prefixPatternID adds the appropriate prefix to a pattern ID
func prefixPatternID(patternID string) string {
	parts := strings.Split(patternID, "_")
	switch len(parts) {
	case 2:
		return fmt.Sprintf("category_java_%s_%s", parts[0], parts[1])
	case 3:
		return fmt.Sprintf("category_%s_%s_%s", parts[0], parts[1], parts[2])
	case 4:
		return fmt.Sprintf("%s_%s_%s_%s", parts[0], parts[1], parts[2], parts[3])
	default:
		return patternID
	}
}

// convertPatternIDToPMD converts a Codacy pattern ID to PMD 6 format
func convertPatternIDToPMD(patternID string) (string, error) {
	if newID, ok := DeprecatedReferences[patternID]; ok {
		patternID = newID
	}

	var parts []string
	if strings.Contains(patternID, "/") {
		parts = strings.Split(patternID, "/")
	} else {
		id := strings.TrimPrefix(patternID, "PMD_")
		parts = strings.Split(id, "_")
		if parts[0] == "category" {
			parts = parts[1:]
		}
	}

	if len(parts) < 3 {
		return "", fmt.Errorf("invalid pattern ID format: %s", patternID)
	}

	language := parts[0]
	category := parts[1]
	rule := strings.Join(parts[2:], "_")

	return fmt.Sprintf("category/%s/%s.xml/%s", language, category, rule), nil
}

// convertPatternIDToPMD7 converts a Codacy pattern ID to PMD 7 format
func convertPatternIDToPMD7(patternID string) (string, error) {
	if newID, ok := DeprecatedReferences[patternID]; ok {
		patternID = newID
	}

	patternID = strings.TrimPrefix(patternID, "PMD7_")
	parts := strings.Split(patternID, "_")

	if len(parts) < 4 || parts[0] != "category" {
		return "", fmt.Errorf("invalid PMD 7 pattern ID: %s", patternID)
	}

	lang := parts[1]
	category := parts[2]
	rule := strings.Join(parts[3:], "_")

	return fmt.Sprintf("category/%s/%s.xml/%s", lang, category, rule), nil
}

// createPmdConfigGeneric abstracts the config creation for both PMD and PMD7
func createPmdConfigGeneric(
	configuration []domain.PatternConfiguration,
	header string,
	convertPatternID func(string) (string, error),
	isPMD7 bool,
) string {
	if len(configuration) == 0 {
		if header == pmd7Header {
			return header + `</ruleset>`
		}
		return defaultPMDRuleset
	}

	var rules []Rule
	for _, pattern := range configuration {
		enabled := true
		var params []Parameter

		for _, param := range pattern.Parameters {
			if param.Name == "enabled" && param.Value == "false" {
				enabled = false
				break
			}
			if param.Name != "enabled" && param.Name != "version" {
				value := param.Value
				if value == "" {
					value = param.Default
				}
				if value != "" {
					params = append(params, Parameter{Name: param.Name, Value: value})
				}
			}
		}

		patternID := pattern.PatternDefinition.Id
		if !isPMD7 && !strings.HasPrefix(patternID, "PMD_") && !strings.Contains(patternID, "/") {
			patternID = prefixPatternID(patternID)
		}

		rules = append(rules, Rule{
			PatternID:  patternID,
			Parameters: params,
			Enabled:    enabled,
		})
	}

	var rulesetXML strings.Builder
	rulesetXML.WriteString(header)
	if !strings.HasSuffix(header, "\n") {
		rulesetXML.WriteString("\n")
	}

	seen := make(map[string]bool)
	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}

		ref, err := convertPatternID(rule.PatternID)
		if err != nil {
			continue
		}
		if seen[ref] {
			continue
		}
		seen[ref] = true

		if len(rule.Parameters) == 0 {
			rulesetXML.WriteString(fmt.Sprintf("    <rule ref=\"%s\"/>\n", ref))
			continue
		}

		rulesetXML.WriteString(fmt.Sprintf(`    <rule ref="%s">
        <properties>`, ref))
		for _, p := range rule.Parameters {
			rulesetXML.WriteString(fmt.Sprintf("\n            <property name=\"%s\" value=\"%s\"/>", p.Name, p.Value))
		}
		rulesetXML.WriteString(`
        </properties>
    </rule>
`)
	}

	rulesetXML.WriteString("</ruleset>")
	return rulesetXML.String()
}
