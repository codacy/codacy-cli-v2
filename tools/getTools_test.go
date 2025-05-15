package tools

import (
	"testing"

	"codacy/cli-v2/domain"

	"github.com/stretchr/testify/assert"
)

func TestFilterToolsByConfigUsage(t *testing.T) {
	tests := []struct {
		name          string
		inputTools    []domain.Tool
		expectedCount int
		expectedTools []string
	}{
		{
			name: "tools with UsesConfigurationFile=true should be filtered out",
			inputTools: []domain.Tool{
				{
					Uuid: "eslint-uuid",
					Name: "eslint",
					Settings: struct {
						Enabled               bool `json:"isEnabled"`
						HasConfigurationFile  bool `json:"hasConfigurationFile"`
						UsesConfigurationFile bool `json:"usesConfigurationFile"`
					}{
						Enabled:               true,
						HasConfigurationFile:  true,
						UsesConfigurationFile: true,
					},
				},
				{
					Uuid: "trivy-uuid",
					Name: "trivy",
					Settings: struct {
						Enabled               bool `json:"isEnabled"`
						HasConfigurationFile  bool `json:"hasConfigurationFile"`
						UsesConfigurationFile bool `json:"usesConfigurationFile"`
					}{
						Enabled:               true,
						HasConfigurationFile:  false,
						UsesConfigurationFile: false,
					},
				},
				{
					Uuid: "pylint-uuid",
					Name: "pylint",
					Settings: struct {
						Enabled               bool `json:"isEnabled"`
						HasConfigurationFile  bool `json:"hasConfigurationFile"`
						UsesConfigurationFile bool `json:"usesConfigurationFile"`
					}{
						Enabled:               true,
						HasConfigurationFile:  false,
						UsesConfigurationFile: false,
					},
				},
			},
			expectedCount: 2,
			expectedTools: []string{"trivy", "pylint"},
		},
		{
			name: "all tools using config should be filtered out",
			inputTools: []domain.Tool{
				{
					Uuid: "eslint-uuid",
					Name: "eslint",
					Settings: struct {
						Enabled               bool `json:"isEnabled"`
						HasConfigurationFile  bool `json:"hasConfigurationFile"`
						UsesConfigurationFile bool `json:"usesConfigurationFile"`
					}{
						Enabled:               true,
						HasConfigurationFile:  true,
						UsesConfigurationFile: true,
					},
				},
				{
					Uuid: "trivy-uuid",
					Name: "trivy",
					Settings: struct {
						Enabled               bool `json:"isEnabled"`
						HasConfigurationFile  bool `json:"hasConfigurationFile"`
						UsesConfigurationFile bool `json:"usesConfigurationFile"`
					}{
						Enabled:               true,
						HasConfigurationFile:  true,
						UsesConfigurationFile: true,
					},
				},
			},
			expectedCount: 0,
			expectedTools: []string{},
		},
		{
			name: "no tools using config should all pass through",
			inputTools: []domain.Tool{
				{
					Uuid: "eslint-uuid",
					Name: "eslint",
					Settings: struct {
						Enabled               bool `json:"isEnabled"`
						HasConfigurationFile  bool `json:"hasConfigurationFile"`
						UsesConfigurationFile bool `json:"usesConfigurationFile"`
					}{
						Enabled:               true,
						HasConfigurationFile:  true,
						UsesConfigurationFile: false,
					},
				},
				{
					Uuid: "pylint-uuid",
					Name: "pylint",
					Settings: struct {
						Enabled               bool `json:"isEnabled"`
						HasConfigurationFile  bool `json:"hasConfigurationFile"`
						UsesConfigurationFile bool `json:"usesConfigurationFile"`
					}{
						Enabled:               true,
						HasConfigurationFile:  false,
						UsesConfigurationFile: false,
					},
				},
			},
			expectedCount: 2,
			expectedTools: []string{"eslint", "pylint"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use the actual function being tested
			result := FilterToolsByConfigUsage(tt.inputTools)

			// Verify the count matches
			assert.Equal(t, tt.expectedCount, len(result),
				"Expected %d tools after filtering, got %d", tt.expectedCount, len(result))

			// Verify each expected tool is in the result
			for _, expectedTool := range tt.expectedTools {
				found := false
				for _, tool := range result {
					if tool.Name == expectedTool {
						found = true
						break
					}
				}
				assert.True(t, found, "Expected tool %s not found in filtered results", expectedTool)
			}

			// Verify no tools with UsesConfigurationFile=true are in the result
			for _, tool := range result {
				assert.False(t, tool.Settings.UsesConfigurationFile,
					"Tool %s with UsesConfigurationFile=true should not be in filtered results", tool.Name)
			}
		})
	}
}
