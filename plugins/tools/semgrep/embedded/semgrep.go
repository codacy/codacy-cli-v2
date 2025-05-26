// Package embedded contains embedded files used by the tools package
package embedded

import "embed"

//go:embed rules.yaml
var rulesFS embed.FS

// GetSemgrepRules returns the embedded Semgrep rules
func GetSemgrepRules() []byte {
	data, err := rulesFS.ReadFile("rules.yaml")
	if err != nil {
		panic(err) // This should never happen as the file is embedded
	}
	return data
}
