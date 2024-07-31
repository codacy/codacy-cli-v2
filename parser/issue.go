package parser

import (
	"strings"
)

type Issue struct {
	Source   string
	Line     int
	Type     string
	Message  string
	Level    string
	Category string
}

func ExtractIssuesFromSarif(sarif *Sarif) []Issue {
	var issues []Issue
	for _, run := range sarif.Runs {
		for _, result := range run.Results {
			for _, location := range result.Locations {
				issue := Issue{
					Source:   strings.TrimPrefix(location.PhysicalLocation.ArtifactLocation.URI, "file://"),
					Line:     location.PhysicalLocation.Region.StartLine,
					Type:     result.RuleID,
					Message:  result.Message.Text,
					Level:    result.Level,
					Category: run.Tool.Driver.Rules[result.RuleIndex].ShortDescription.Text,
				}
				issues = append(issues, issue)
			}
		}
	}
	return issues
}

func GroupIssuesBySource(issues []Issue) map[string]interface{} {
	groups := make(map[string]interface{})
	for _, issue := range issues {
		if _, ok := groups[issue.Source]; !ok {
			groups[issue.Source] = map[string]interface{}{
				"filename": issue.Source,
				"results":  []interface{}{},
			}
		}
		group := groups[issue.Source].(map[string]interface{})
		group["results"] = append(group["results"].([]interface{}), map[string]interface{}{
			"Issue": map[string]interface{}{
				"patternId": map[string]string{
					"value": issue.Type,
				},
				"filename": issue.Source,
				"message": map[string]string{
					"text": issue.Message,
				},
				"level":    issue.Level,
				"category": issue.Category,
				"location": map[string]interface{}{
					"LineLocation": map[string]int{
						"line": issue.Line,
					},
				},
			},
		})
	}
	return groups
}
