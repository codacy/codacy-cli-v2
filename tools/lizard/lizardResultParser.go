package lizard

import (
	"codacy/cli-v2/domain"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// parseLizardResults parses the output from Lizard into a structured format
func parseLizardResults(output string) (*LizardResults, error) {
	lines := strings.Split(output, "\n")
	results := &LizardResults{
		Methods: make([]LizardMethod, 0),
		Files:   make([]LizardFile, 0),
	}

	var isMethodSection, isFileSection, isWarningSection bool

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.Contains(line, "!!!! Warnings") {
			isWarningSection = true
			continue
		}

		if strings.HasPrefix(line, "===") {
			isMethodSection = false
			isFileSection = false
			if !isWarningSection {
				continue
			}
			continue
		}

		if isWarningSection {
			continue
		}

		if strings.HasPrefix(line, "===") || strings.HasPrefix(line, "---") || line == "" || strings.Contains(line, "file analyzed") {
			continue
		}

		if strings.Contains(line, "NLOC    CCN   token  PARAM  length  location") {
			isMethodSection = true
			isFileSection = false
			continue
		}

		if strings.Contains(line, "NLOC    Avg.NLOC  AvgCCN  Avg.token  function_cnt    file") {
			isMethodSection = false
			isFileSection = true
			continue
		}

		if isMethodSection {
			// Replace multiple spaces and @ with single space
			line = regexp.MustCompile(`\s+|@`).ReplaceAllString(line, " ")
			line = strings.TrimSpace(line)
			parts := strings.Split(line, " ")
			if len(parts) != 8 {
				continue
			}

			fromToLine := strings.Split(parts[6], "-")
			if len(fromToLine) != 2 {
				continue
			}

			fromLine, err := strconv.Atoi(fromToLine[0])
			if err != nil {
				continue
			}

			toLine, err := strconv.Atoi(fromToLine[1])
			if err != nil {
				continue
			}

			nloc, _ := strconv.Atoi(parts[0])
			ccn, _ := strconv.Atoi(parts[1])
			tokens, _ := strconv.Atoi(parts[2])
			params, _ := strconv.Atoi(parts[3])

			results.Methods = append(results.Methods, LizardMethod{
				Name:     parts[5],
				FromLine: fromLine,
				ToLine:   toLine,
				File:     parts[7],
				Nloc:     nloc,
				Ccn:      ccn,
				Params:   params,
				Tokens:   tokens,
			})
		}

		if isFileSection {
			// Replace multiple spaces with single space
			line = regexp.MustCompile(`\s+`).ReplaceAllString(line, " ")
			line = strings.TrimSpace(line)
			parts := strings.Split(line, " ")
			if len(parts) != 6 {
				continue
			}

			nloc, _ := strconv.Atoi(parts[0])
			avgNloc, _ := strconv.ParseFloat(parts[1], 64)
			avgCcn, _ := strconv.ParseFloat(parts[2], 64)
			avgTokens, _ := strconv.ParseFloat(parts[3], 64)
			methodsCount, _ := strconv.Atoi(parts[4])

			// Calculate maxCcn for this file
			maxCcn := 0
			for _, method := range results.Methods {
				if method.File == parts[5] && method.Ccn > maxCcn {
					maxCcn = method.Ccn
				}
			}

			results.Files = append(results.Files, LizardFile{
				File:          parts[5],
				Nloc:          nloc,
				MaxCcn:        maxCcn,
				AverageNloc:   avgNloc,
				AverageCcn:    avgCcn,
				AverageTokens: avgTokens,
				MethodsCount:  methodsCount,
			})
		}
	}

	return results, nil
}

// generateIssuesFromResults generates SARIF issues from Lizard results
func generateIssuesFromResults(results *LizardResults, patterns []domain.PatternDefinition) []Issue {
	var issues []Issue

	// Create a map of pattern IDs to their definitions for quick lookup
	patternMap := make(map[string]domain.PatternDefinition)
	for _, pattern := range patterns {
		patternMap[pattern.Id] = pattern
	}

	// Check method-level issues
	for _, method := range results.Methods {
		// Check NLOC rules
		checkMetricThreshold(method.Nloc, method.File, method.Name, method.FromLine, method.ToLine, "nloc", patternMap, &issues)

		// Check CCN rules
		checkMetricThreshold(method.Ccn, method.File, method.Name, method.FromLine, method.ToLine, "ccn", patternMap, &issues)

		// Check parameter count rules
		checkMetricThreshold(method.Params, method.File, method.Name, method.FromLine, method.ToLine, "parameter-count", patternMap, &issues)
	}

	// Check file-level issues
	for _, file := range results.Files {
		// Check file NLOC rules
		checkMetricThreshold(file.Nloc, file.File, file.File, 0, 0, "file-nloc", patternMap, &issues)
	}

	return issues
}

// checkMetricThreshold checks if a metric value exceeds any thresholds and creates issues accordingly
func checkMetricThreshold(value int, file string, methodName string, startLine int, endLine int, metricType string, patternMap map[string]domain.PatternDefinition, issues *[]Issue) {
	// Check severities in order from most severe to least severe
	severities := []string{"critical", "medium", "minor"}
	for _, severity := range severities {
		patternID := fmt.Sprintf("Lizard_%s-%s", metricType, severity)
		pattern, exists := patternMap[patternID]
		if !exists {
			continue
		}

		threshold := getThresholdFromParams(pattern.Parameters)
		if value > threshold {
			message := formatMessage(methodName, value, threshold, metricType)
			*issues = append(*issues, Issue{
				File:        file,
				Message:     message,
				Severity:    strings.ToLower(pattern.SeverityLevel),
				StartLine:   startLine,
				EndLine:     endLine,
				RuleID:      pattern.Id,
				Description: pattern.Description,
			})
			// Break after finding the first (most severe) violation
			break
		}
	}
}

// formatMessage creates a formatted message based on the metric type and whether it's a method or file issue
func formatMessage(methodName string, value int, threshold int, metricType string) string {
	if methodName == "" {
		// File-level issue
		return fmt.Sprintf("File has %d lines of code (limit is %d)", value, threshold)
	}

	// Method-level issue
	switch metricType {
	case "nloc":
		return fmt.Sprintf("Method %s has %d lines of code (limit is %d)", methodName, value, threshold)
	case "ccn":
		return fmt.Sprintf("Method %s has a cyclomatic complexity of %d (limit is %d)", methodName, value, threshold)
	case "parameter-count":
		return fmt.Sprintf("Method %s has %d parameters (limit is %d)", methodName, value, threshold)
	default:
		return fmt.Sprintf("Method %s has %d %s (limit is %d)", methodName, value, metricType, threshold)
	}
}
