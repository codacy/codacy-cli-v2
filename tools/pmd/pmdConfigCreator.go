package pmd

import (
	"codacy/cli-v2/tools"
	_ "embed"
	"encoding/xml"
	"fmt"
	"strings"
)

//go:embed default-ruleset.xml
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
var DeprecatedReferences = map[string]string{
	// Add deprecated pattern mappings here
	// Example: "rulesets_java_design_ExcessiveClassLength": "category_java_design_ExcessiveClassLength",
}

// prefixPatternID adds the appropriate prefix to a pattern ID
func prefixPatternID(patternID string) string {
	parts := strings.Split(patternID, "_")

	// Handle different pattern ID formats
	switch len(parts) {
	case 2:
		// Format: "patternCategory_patternName"
		return fmt.Sprintf("category_java_%s_%s", parts[0], parts[1])
	case 3:
		// Format: "langAlias_patternCategory_patternName"
		return fmt.Sprintf("category_%s_%s_%s", parts[0], parts[1], parts[2])
	case 4:
		// Format: "root_langAlias_patternCategory_patternName"
		return fmt.Sprintf("%s_%s_%s_%s", parts[0], parts[1], parts[2], parts[3])
	default:
		// Return as is if format is unknown
		return patternID
	}
}

// convertPatternIDToPMD converts a Codacy pattern ID to PMD format
func convertPatternIDToPMD(patternID string) (string, error) {
	// Check if this is a deprecated pattern
	if newID, ok := DeprecatedReferences[patternID]; ok {
		patternID = newID
	}

	// Handle both formats:
	// 1. "java/design/NPathComplexity"
	// 2. "PMD_category_java_design_NPathComplexity"
	// 3. "PMD_category_apex_security_ApexSharingViolations"
	// 4. "PMD_category_plsql_errorprone_TO_TIMESTAMPWithoutDateFormat"

	var parts []string
	if strings.Contains(patternID, "/") {
		parts = strings.Split(patternID, "/")
	} else {
		// Remove PMD_ prefix if present
		id := strings.TrimPrefix(patternID, "PMD_")
		// Split by underscore and remove "category" if present
		parts = strings.Split(id, "_")
		if parts[0] == "category" {
			parts = parts[1:]
		}
	}

	if len(parts) < 3 {
		return "", fmt.Errorf("invalid pattern ID format: %s", patternID)
	}

	// Extract language, category, and rule
	language := parts[0] // java, apex, etc.
	category := parts[1] // design, security, etc.
	rule := parts[2]     // rule name

	// If there are more parts, combine them with the rule name
	if len(parts) > 3 {
		rule = strings.Join(parts[2:], "_")
	}

	return fmt.Sprintf("category/%s/%s.xml/%s", language, category, rule), nil
}

// generateRuleXML generates XML for a single rule
func generateRuleXML(rule Rule) (string, error) {
	pmdRef, err := convertPatternIDToPMD(rule.PatternID)
	if err != nil {
		return "", err
	}

	if len(rule.Parameters) == 0 {
		return fmt.Sprintf(`    <rule ref="%s"/>`, pmdRef), nil
	}

	// Generate rule with parameters
	var params strings.Builder
	for _, param := range rule.Parameters {
		if param.Name != "enabled" && param.Name != "version" { // Skip enabled and version parameters
			params.WriteString(fmt.Sprintf(`
            <property name="%s" value="%s"/>`, param.Name, param.Value))
		}
	}

	return fmt.Sprintf(`    <rule ref="%s">
        <properties>%s
        </properties>
    </rule>`, pmdRef, params.String()), nil
}

// ConvertToPMDRuleset converts Codacy rules to PMD ruleset format
func ConvertToPMDRuleset(rules []Rule) (string, error) {
	var rulesetXML strings.Builder
	rulesetXML.WriteString(`<?xml version="1.0"?>
<ruleset name="Codacy PMD Ruleset"
    xmlns="http://pmd.sourceforge.net/ruleset/2.0.0"
    xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
    xsi:schemaLocation="http://pmd.sourceforge.net/ruleset/2.0.0 https://pmd.sourceforge.io/ruleset_2_0_0.xsd">
    <description>Codacy PMD Ruleset</description>

`)

	// Track processed rules to avoid duplicates
	processedRules := make(map[string]bool)

	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}

		pmdRef, err := convertPatternIDToPMD(rule.PatternID)
		if err != nil {
			return "", fmt.Errorf("error converting pattern ID: %w", err)
		}

		// Skip if we've already processed this rule
		if processedRules[pmdRef] {
			continue
		}

		processedRules[pmdRef] = true

		ruleXML, err := generateRuleXML(rule)
		if err != nil {
			return "", fmt.Errorf("error generating rule XML: %w", err)
		}
		rulesetXML.WriteString(ruleXML + "\n")
	}

	rulesetXML.WriteString("</ruleset>")
	return rulesetXML.String(), nil
}

// CreatePmdConfig creates a PMD configuration from the provided tool configuration
func CreatePmdConfig(configuration tools.ToolConfiguration) string {
	// If no patterns provided, return the default ruleset
	if len(configuration.PatternsConfiguration) == 0 {
		return defaultPMDRuleset
	}

	// Convert ToolConfiguration to our Rule format
	var rules []Rule
	for _, pattern := range configuration.PatternsConfiguration {
		// Check if pattern is enabled
		patternEnabled := true
		var parameters []Parameter

		for _, param := range pattern.ParameterConfigurations {
			if param.Name == "enabled" && param.Value == "false" {
				patternEnabled = false
				break
			} else if param.Name != "enabled" {
				// Store non-enabled parameters
				parameters = append(parameters, Parameter{
					Name:  param.Name,
					Value: param.Value,
				})
			}
		}

		// Apply prefix to pattern ID if needed
		patternID := pattern.PatternId
		if !strings.HasPrefix(patternID, "PMD_") && !strings.Contains(patternID, "/") {
			patternID = prefixPatternID(patternID)
		}

		rules = append(rules, Rule{
			PatternID:  patternID,
			Parameters: parameters,
			Enabled:    patternEnabled,
		})
	}

	// Convert rules to PMD ruleset
	rulesetXML, err := ConvertToPMDRuleset(rules)
	if err != nil {
		return defaultPMDRuleset
	}

	return rulesetXML
}
