package cmd

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetRelativePath(t *testing.T) {
	const baseDir = "/home/user/project/src"

	tests := []struct {
		name     string
		baseDir  string
		fullURI  string
		expected string
	}{
		{
			name:     "1. File URI with standard path",
			baseDir:  baseDir,
			fullURI:  "file:///home/user/project/src/lib/file.go",
			expected: "lib/file.go",
		},
		{
			name:     "2. File URI with baseDir as the file path",
			baseDir:  baseDir,
			fullURI:  "file:///home/user/project/src",
			expected: ".",
		},
		{
			name:     "3. Simple path (no scheme)",
			baseDir:  baseDir,
			fullURI:  "/home/user/project/src/main.go",
			expected: "main.go",
		},
		{
			name:    "4. URI outside baseDir (should return absolute path if relative fails)",
			baseDir: baseDir,
			fullURI: "file:///etc/config/app.json",
			// This is outside of baseDir, so we expect the absolute path starting from the baseDir root
			expected: "../../../../etc/config/app.json",
		},
		{
			name:     "5. Plain URI with different scheme (should be treated as plain path)",
			baseDir:  baseDir,
			fullURI:  "http://example.com/api/v1/file.go",
			expected: "http://example.com/api/v1/file.go",
		},
		{
			name:     "6. Empty URI",
			baseDir:  baseDir,
			fullURI:  "",
			expected: "",
		},
		{
			name:     "7. Windows path on a file URI (should correctly strip the leading slash from the path component)",
			baseDir:  "C:\\Users\\dev\\repo",
			fullURI:  "file:///C:/Users/dev/repo/app/main.go",
			expected: "/C:/Users/dev/repo/app/main.go",
		},
		{
			name:     "8. URI with spaces (URL encoded)",
			baseDir:  baseDir,
			fullURI:  "file:///home/user/project/src/file%20with%20spaces.go",
			expected: "file with spaces.go",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := getRelativePath(tt.baseDir, tt.fullURI)
			expectedNormalized := filepath.FromSlash(tt.expected)
			assert.Equal(t, expectedNormalized, actual, "Relative path should match expected")
		})
	}
}
func TestGetToolShortName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "MappedTool_ESLint8",
			input:    "ESLint",
			expected: "eslint-8",
		},
		{
			name:     "MappedTool_PMD7",
			input:    "PMD7",
			expected: "pmd-7",
		},
		{
			name:     "MappedTool_Pylint",
			input:    "Pylint",
			expected: "pylintpython3",
		},
		{
			name:     "UnmappedTool_Fallback",
			input:    "NewToolName",
			expected: "NewToolName",
		},
		{
			name:     "UnmappedTool_AnotherFallback",
			input:    "SomeAnalyzer",
			expected: "SomeAnalyzer",
		},
		{
			name:     "EmptyInput_Fallback",
			input:    "",
			expected: "",
		},
		{
			name:     "MappedTool_Deprecated",
			input:    "ESLint (deprecated)",
			expected: "eslint",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := getToolShortName(tt.input)
			if actual != tt.expected {
				t.Errorf("getToolShortName(%q) = %q; want %q", tt.input, actual, tt.expected)
			}
		})
	}
}
