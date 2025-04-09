package pylint

import (
	"codacy/cli-v2/domain"
	"fmt"
	"log"
	"strings"
)

// getDefaultParametersForPatterns returns a map of pattern IDs to their default parameters
func getDefaultParametersForPatterns(patternIDs []string) map[string][]domain.ParameterConfiguration {
	defaultParams := make(map[string][]domain.ParameterConfiguration)

	for _, patternID := range patternIDs {
		if params, exists := PatternDefaultParameters[patternID]; exists {
			defaultParams[patternID] = params
		}
	}

	return defaultParams
}

// writePylintRCHeader writes the common header sections to the RC content
func writePylintRCHeader(rcContent *strings.Builder) {
	rcContent.WriteString("[MASTER]\n")
	rcContent.WriteString("ignore=CVS\n")
	rcContent.WriteString("persistent=yes\n")
	rcContent.WriteString("load-plugins=\n\n")
	rcContent.WriteString("[MESSAGES CONTROL]\n")
	rcContent.WriteString("disable=all\n")
}

// writeEnabledPatterns writes the enabled patterns section to the RC content
func writeEnabledPatterns(rcContent *strings.Builder, patternIDs []string) {
	if len(patternIDs) > 0 {
		rcContent.WriteString(fmt.Sprintf("enable=%s\n", strings.Join(patternIDs, ",")))
	}
	rcContent.WriteString("\n")
}

// writeParametersBySection writes the parameters grouped by section to the RC content
func writeParametersBySection(rcContent *strings.Builder, groupedParams map[string][]domain.ParameterConfiguration) {
	for sectionName, params := range groupedParams {
		rcContent.WriteString(fmt.Sprintf("[%s]\n", sectionName))
		for _, param := range params {
			rcContent.WriteString(fmt.Sprintf("%s=%s\n", param.Name, param.Value))
		}
		rcContent.WriteString("\n")
	}
}

// groupParametersByPatterns groups parameters from patterns into sections
func groupParametersByPatterns(patterns []domain.PatternConfiguration) map[string][]domain.ParameterConfiguration {
	groupedParams := make(map[string][]domain.ParameterConfiguration)

	for _, pattern := range patterns {
		patternID := extractPatternId(pattern.PatternDefinition.Id)
		params := pattern.Parameters

		// If no parameters, check defaults
		if len(params) == 0 {
			if defaultParams, exists := PatternDefaultParameters[patternID]; exists {
				params = defaultParams
			}
		}

		// Add parameters to their respective sections
		for _, param := range params {
			sectionName := GetParameterSection(param.Name)
			if sectionName == nil {
				log.Printf("Parameter %s has no section name", param.Name)
				continue
			}

			// Check if parameter already exists in section
			exists := false
			for _, existingParam := range groupedParams[*sectionName] {
				if existingParam.Name == param.Name {
					exists = true
					break
				}
			}

			// Only add if not already present
			if !exists {
				groupedParams[*sectionName] = append(groupedParams[*sectionName], param)
			}
		}
	}

	return groupedParams
}

func GeneratePylintRCDefault() string {
	var rcContent strings.Builder

	writePylintRCHeader(&rcContent)
	writeEnabledPatterns(&rcContent, DefaultPatterns)

	// Get default parameters for enabled patterns
	defaultParams := getDefaultParametersForPatterns(DefaultPatterns)

	// Convert default parameters to pattern configurations
	var patterns []domain.PatternConfiguration
	for patternID, params := range defaultParams {
		patterns = append(patterns, domain.PatternConfiguration{
			PatternDefinition: domain.PatternDefinition{
				Id: "PyLintPython3_" + patternID,
			},
			Parameters: params,
		})
	}

	// Group and write parameters
	groupedParams := groupParametersByPatterns(patterns)
	writeParametersBySection(&rcContent, groupedParams)

	return rcContent.String()
}

// GeneratePylintRC generates a pylintrc file content with the specified patterns enabled
func GeneratePylintRC(config []domain.PatternConfiguration) string {
	var rcContent strings.Builder

	writePylintRCHeader(&rcContent)

	// Collect enabled pattern IDs
	var enabledPatternsIds []string

	for _, pattern := range config {
		patternID := extractPatternId(pattern.PatternDefinition.Id)
		enabledPatternsIds = append(enabledPatternsIds, patternID)
	}

	writeEnabledPatterns(&rcContent, enabledPatternsIds)

	// Group and write parameters
	groupedParams := groupParametersByPatterns(config)
	writeParametersBySection(&rcContent, groupedParams)

	return rcContent.String()
}

// extractPatternId returns the part of the pattern ID after the underscore
// For example: "PyLintPython3_C0301" -> "C0301"
func extractPatternId(fullID string) string {
	parts := strings.Split(fullID, "_")
	if len(parts) > 1 {
		return parts[1]
	}
	return fullID
}
