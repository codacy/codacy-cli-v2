package lizard

import (
	"codacy/cli-v2/domain"
	"codacy/cli-v2/utils"
	"path/filepath"
	"sort"
	"strings"
)

// convertIssuesToSarif converts Lizard issues to SARIF Report
func convertIssuesToSarif(issues []Issue, patterns []domain.PatternDefinition) *utils.SarifReport {
	// Create a map to track unique rules
	rules := make(map[string]utils.Rule)

	// First, add all patterns as rules
	for _, pattern := range patterns {
		rules[pattern.Id] = utils.Rule{
			ID: pattern.Id,
			ShortDescription: utils.MessageText{
				Text: pattern.Description,
			},
			Properties: utils.RuleProperties{
				Tags: []string{strings.ToLower(pattern.SeverityLevel)},
			},
		}
	}

	results := make([]utils.Result, 0)

	// Process each issue
	for _, issue := range issues {
		// Get the base filename for the file path
		filePath := filepath.Base(issue.File)

		// Create the result
		result := utils.Result{
			RuleID: issue.RuleID,
			Level:  strings.ToLower(issue.Severity),
			Message: utils.MessageText{
				Text: issue.Message,
			},
			Locations: []utils.Location{
				{
					PhysicalLocation: utils.PhysicalLocation{
						ArtifactLocation: utils.ArtifactLocation{
							URI: filePath,
						},
						Region: utils.Region{
							StartLine:   issue.StartLine,
							StartColumn: 1,
						},
					},
				},
			},
		}
		results = append(results, result)
	}

	// Sort rules by ID
	ruleIDs := make([]string, 0, len(rules))
	for id := range rules {
		ruleIDs = append(ruleIDs, id)
	}
	sort.Strings(ruleIDs)

	ruleSlice := make([]utils.Rule, 0, len(rules))
	for _, id := range ruleIDs {
		ruleSlice = append(ruleSlice, rules[id])
	}

	return &utils.SarifReport{
		Version: "2.1.0",
		Schema:  "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
		Runs: []utils.Run{
			{
				Tool: utils.Tool{
					Driver: utils.Driver{
						Name:           "Lizard",
						Version:        "1.17.10",
						InformationURI: "https://github.com/terryyin/lizard",
						Rules:          ruleSlice,
					},
				},
				Results: results,
			},
		},
	}
}
