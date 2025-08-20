package tools

import (
	"codacy/cli-v2/domain"
	"fmt"
	"strings"
)

// CreateTrivyConfig generates a Trivy configuration based on the tool configuration
func CreateTrivyConfig(config []domain.PatternConfiguration) string {

	// Default settings - include all severities and scanners
	includeLow := false
	includeMedium := false
	includeHigh := false
	includeCritical := false
	includeSecret := false

	// Process patterns from Codacy API
	for _, pattern := range config {
		// Check if pattern is enabled
		patternEnabled := true
		for _, param := range pattern.Parameters {
			if param.Name == "enabled" && param.Value == "false" {
				patternEnabled = false
			}
		}

		// Map patterns to configurations
		if pattern.PatternDefinition.Id == "Trivy_vulnerability_minor" {
			includeLow = patternEnabled
		}
		if pattern.PatternDefinition.Id == "Trivy_vulnerability_medium" {
			includeMedium = patternEnabled
		}
		if pattern.PatternDefinition.Id == "Trivy_vulnerability_high" {
			includeHigh = patternEnabled
		}
		if pattern.PatternDefinition.Id == "Trivy_vulnerability_critical" {
			includeCritical = patternEnabled
		}
		if pattern.PatternDefinition.Id == "Trivy_vulnerability" {
			// This covers HIGH and CRITICAL
			// Now there are other patterns that turn these severities on
			includeHigh = patternEnabled || includeHigh
			includeCritical = patternEnabled || includeCritical
		}
		if pattern.PatternDefinition.Id == "Trivy_secret" {
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
