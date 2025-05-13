package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMaskSensitiveArgs(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected []string
	}{
		{
			name:     "no sensitive flags",
			args:     []string{"codacy-cli", "analyze", "--tool", "eslint"},
			expected: []string{"codacy-cli", "analyze", "--tool", "eslint"},
		},
		{
			name:     "api token with space",
			args:     []string{"codacy-cli", "init", "--api-token", "secret123", "--tool", "eslint"},
			expected: []string{"codacy-cli", "init", "--api-token", "***", "--tool", "eslint"},
		},
		{
			name:     "api token with equals",
			args:     []string{"codacy-cli", "init", "--api-token=secret123", "--tool", "eslint"},
			expected: []string{"codacy-cli", "init", "--api-token=***", "--tool", "eslint"},
		},
		{
			name:     "repository token at end with space",
			args:     []string{"codacy-cli", "init", "--repository-token", "secret123"},
			expected: []string{"codacy-cli", "init", "--repository-token", "***"},
		},
		{
			name:     "repository token at end with equals",
			args:     []string{"codacy-cli", "init", "--repository-token=secret123"},
			expected: []string{"codacy-cli", "init", "--repository-token=***"},
		},
		{
			name:     "project token at start with space",
			args:     []string{"codacy-cli", "--project-token", "secret123", "analyze"},
			expected: []string{"codacy-cli", "--project-token", "***", "analyze"},
		},
		{
			name:     "project token at start with equals",
			args:     []string{"codacy-cli", "--project-token=secret123", "analyze"},
			expected: []string{"codacy-cli", "--project-token=***", "analyze"},
		},
		{
			name:     "multiple tokens mixed format",
			args:     []string{"codacy-cli", "--api-token=secret1", "--project-token", "secret2"},
			expected: []string{"codacy-cli", "--api-token=***", "--project-token", "***"},
		},
		{
			name:     "token flag at end without value",
			args:     []string{"codacy-cli", "analyze", "--api-token"},
			expected: []string{"codacy-cli", "analyze", "--api-token"},
		},
		{
			name:     "empty value after equals",
			args:     []string{"codacy-cli", "--api-token="},
			expected: []string{"codacy-cli", "--api-token=***"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			masked := maskSensitiveArgs(tt.args)
			assert.Equal(t, tt.expected, masked, "masked arguments should match expected")
		})
	}
}
