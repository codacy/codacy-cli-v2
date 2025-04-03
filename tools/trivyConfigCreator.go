package tools

import (
	"fmt"
	"strings"
)

// CreateTrivyConfig generates a Trivy configuration based on the tool configuration
func CreateTrivyConfig(config ToolConfiguration) string {
	// Default settings - include all severities and scanners
	includeLow := true
	includeMedium := true
	includeHigh := true
	includeCritical := true
	includeSecret := true

	// Process patterns from Codacy API
	for _, pattern := range config.PatternsConfiguration {
		// Check if pattern is enabled
		patternEnabled := true
		for _, param := range pattern.ParameterConfigurations {
			if param.Name == "enabled" && param.Value == "false" {
				patternEnabled = false
			}
		}

		// Map patterns to configurations
		if pattern.PatternId == "Trivy_vulnerability_minor" {
			includeLow = patternEnabled
		}
		if pattern.PatternId == "Trivy_vulnerability_medium" {
			includeMedium = patternEnabled
		}
		if pattern.PatternId == "Trivy_vulnerability" {
			// This covers HIGH and CRITICAL
			includeHigh = patternEnabled
			includeCritical = patternEnabled
		}
		if pattern.PatternId == "Trivy_secret" {
			includeSecret = patternEnabled
		}
	}

	// Build the severity list based on enabled patterns
	var severities []string
	if includeLow {
		severities = append(severities, "LOW")
	}
	if includeMedium {
		severities = append(severities, "MEDIUM")
	}
	if includeHigh {
		severities = append(severities, "HIGH")
	}
	if includeCritical {
		severities = append(severities, "CRITICAL")
	}

	// Build the scanners list
	var scanners []string
	scanners = append(scanners, "vuln") // Always include vuln scanner
	if includeSecret {
		scanners = append(scanners, "secret")
	}

	// Generate trivy.yaml content
	var contentBuilder strings.Builder
	contentBuilder.WriteString("severity:\n")
	for _, sev := range severities {
		contentBuilder.WriteString(fmt.Sprintf("  - %s\n", sev))
	}

	contentBuilder.WriteString("\nscan:\n")
	contentBuilder.WriteString("  scanners:\n")
	for _, scanner := range scanners {
		contentBuilder.WriteString(fmt.Sprintf("    - %s\n", scanner))
	}

	return contentBuilder.String()
}
