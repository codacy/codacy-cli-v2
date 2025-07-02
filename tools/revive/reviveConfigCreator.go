package tools

import (
	"fmt"
	"strings"

	"codacy/cli-v2/domain"
)

// GenerateReviveConfig generates a TOML config for revive based on enabled patterns and their parameters
func GenerateReviveConfig(patterns []domain.PatternConfiguration) string {
	var sb strings.Builder
	sb.WriteString("[revive]\n")
	sb.WriteString("ignoreGeneratedHeader = true\n")
	sb.WriteString("severity = \"warning\"\n")
	sb.WriteString("confidence = 0.8\n")
	sb.WriteString("errorCode = 0\n")
	sb.WriteString("warningCode = 0\n\n")

	enabledRules := make([]string, 0)
	for _, pattern := range patterns {
		if pattern.Enabled {
			ruleName := strings.TrimPrefix(pattern.PatternDefinition.Id, "Revive_")
			enabledRules = append(enabledRules, fmt.Sprintf("\"%s\"", ruleName))
		}
	}
	if len(enabledRules) > 0 {
		sb.WriteString(fmt.Sprintf("rules = [%s]\n\n", strings.Join(enabledRules, ", ")))
	}

	for _, pattern := range patterns {
		if pattern.Enabled {
			ruleName := strings.TrimPrefix(pattern.PatternDefinition.Id, "Revive_")
			sb.WriteString(fmt.Sprintf("[rule.%s]\n", ruleName))
			if len(pattern.Parameters) > 0 {
				sb.WriteString("arguments = [")
				args := make([]string, 0)
				for _, param := range pattern.Parameters {
					// TOML: string values should be quoted
					args = append(args, fmt.Sprintf("\"%s\"", param.Value))
				}
				sb.WriteString(strings.Join(args, ", "))
				sb.WriteString("]\n")
			}
			sb.WriteString("\n")
		}
	}

	return sb.String()
}
