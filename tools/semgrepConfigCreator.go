package tools

import (
	"codacy/cli-v2/domain"
	"fmt"
	"strings"
)

// supportedLanguages is a list of all languages supported by Semgrep
var supportedLanguages = []string{
	"apex", "bash", "c", "c#", "c++", "cairo",
	"clojure", "cpp", "csharp", "dart", "docker",
	"dockerfile", "elixir", "ex", "generic", "go",
	"golang", "hack", "hcl", "html", "java",
	"javascript", "js", "json", "jsonnet", "julia",
	"kotlin", "kt", "lisp", "lua", "move_on_aptos",
	"none", "ocaml", "php", "promql", "proto",
	"proto3", "protobuf", "py", "python", "python2",
	"python3", "ql", "r", "regex", "ruby",
	"rust", "scala", "scheme", "sh", "sol",
	"solidity", "swift", "terraform", "tf", "ts",
	"typescript", "vue", "xml", "yaml",
}

// CreateSemgrepConfig generates a Semgrep configuration based on the tool configuration
func CreateSemgrepConfig(config []domain.PatternConfiguration) string {
	// Build the rules list
	var rules []string

	// Process each pattern from the API
	for _, pattern := range config {

		// Skip if pattern is not enabled
		if !pattern.Enabled || !pattern.PatternDefinition.Enabled {
			continue
		}

		// Skip if no languages defined
		if len(pattern.PatternDefinition.Languages) == 0 {
			continue
		}

		// Get the first language (Semgrep only supports one language per rule)
		language := strings.ToLower(pattern.PatternDefinition.Languages[0])

		// Map language names to Semgrep supported languages
		switch language {
		case "csharp":
			language = "c#"
		case "cpp":
			language = "c++"
		case "golang":
			language = "go"
		case "js":
			language = "javascript"
		case "kt":
			language = "kotlin"
		case "py":
			language = "python"
		case "sh":
			language = "bash"
		case "tf":
			language = "terraform"
		case "ts":
			language = "typescript"
		}

		// Skip if language is not supported by Semgrep
		isSupported := false
		for _, supportedLang := range supportedLanguages {
			if language == supportedLang {
				isSupported = true
				break
			}
		}
		if !isSupported {
			continue
		}

		// Extract rule ID from pattern ID
		parts := strings.SplitN(pattern.PatternDefinition.Id, "_", 2)
		if len(parts) != 2 {
			continue // Skip invalid pattern IDs
		}

		// The rest is the rule path
		rulePath := parts[1]

		// Ensure we have a valid ID
		if rulePath == "" {
			rulePath = "default_rule"
		}

		// Create rule entry
		rule := fmt.Sprintf(`  - id: %s
    pattern: |
      $X
    message: "Semgrep rule: %s"
    languages: [%s]
    severity: %s`,
			strings.ReplaceAll(rulePath, ".", "_"), // Replace dots with underscores in ID
			rulePath,
			language,
			strings.ToUpper(pattern.PatternDefinition.SeverityLevel))

		rules = append(rules, rule)
	}

	// If no rules were added, use a default configuration
	if len(rules) == 0 {
		return `rules:
  - id: default_rule
    pattern: |
      $X
    message: "Semgrep analysis"
    languages: [generic]
    severity: INFO
`
	}

	// Generate semgrep.yaml content
	var contentBuilder strings.Builder
	contentBuilder.WriteString("rules:\n")
	for _, rule := range rules {
		contentBuilder.WriteString(rule + "\n")
	}

	return contentBuilder.String()
}
