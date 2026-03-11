// Package embedded provides access to embedded Opengrep rules.
package embedded

import "embed"

//go:embed rules.yaml
var rulesFS embed.FS

// GetOpengrepRules returns the embedded Opengrep rules.
// Opengrep is compatible with the semgrep rule format, so the same rules are used.
func GetOpengrepRules() []byte {
	data, err := rulesFS.ReadFile("rules.yaml")
	if err != nil {
		panic(err) // This should never happen as the file is embedded
	}
	return data
}
