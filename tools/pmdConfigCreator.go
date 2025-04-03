package tools

import (
	_ "embed"
	"fmt"
	"strings"
)

//go:embed pmd/default-ruleset.xml
var defaultPMDRuleset string

func CreatePmdConfig(configuration ToolConfiguration) string {
	// If no patterns provided, return the default ruleset
	if len(configuration.PatternsConfiguration) == 0 {
		return defaultPMDRuleset
	}

	var contentBuilder strings.Builder

	// Write XML header and ruleset opening with correct name
	contentBuilder.WriteString("<?xml version=\"1.0\"?>\n")
	contentBuilder.WriteString("<ruleset name=\"Codacy PMD Ruleset\"\n")
	contentBuilder.WriteString("    xmlns=\"http://pmd.sourceforge.net/ruleset/2.0.0\"\n")
	contentBuilder.WriteString("    xmlns:xsi=\"http://www.w3.org/2001/XMLSchema-instance\"\n")
	contentBuilder.WriteString("    xsi:schemaLocation=\"http://pmd.sourceforge.net/ruleset/2.0.0 https://pmd.sourceforge.io/ruleset_2_0_0.xsd\">\n\n")

	// Process patterns from configuration
	for _, pattern := range configuration.PatternsConfiguration {
		// Check if pattern is enabled
		patternEnabled := true
		var properties []string

		for _, param := range pattern.ParameterConfigurations {
			if param.Name == "enabled" && param.Value == "false" {
				patternEnabled = false
				break
			} else if param.Name != "enabled" {
				// Store non-enabled parameters for properties
				properties = append(properties, fmt.Sprintf("            <property name=\"%s\" value=\"%s\"/>", param.Name, param.Value))
			}
		}

		if patternEnabled {
			// Convert pattern ID to correct PMD path format
			// e.g., "java/codestyle/AtLeastOneConstructor" -> "category/java/codestyle.xml/AtLeastOneConstructor"
			parts := strings.Split(pattern.PatternId, "/")
			if len(parts) >= 2 {
				category := parts[len(parts)-2]
				rule := parts[len(parts)-1]
				contentBuilder.WriteString(fmt.Sprintf("    <rule ref=\"category/java/%s.xml/%s\"", category, rule))

				// Add properties if any exist
				if len(properties) > 0 {
					contentBuilder.WriteString(">\n        <properties>\n")
					for _, prop := range properties {
						contentBuilder.WriteString(prop + "\n")
					}
					contentBuilder.WriteString("        </properties>\n    </rule>\n")
				} else {
					contentBuilder.WriteString("/>\n")
				}
			}
		}
	}

	contentBuilder.WriteString("</ruleset>")
	return contentBuilder.String()
}
