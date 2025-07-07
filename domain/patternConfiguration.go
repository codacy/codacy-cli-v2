package domain

// ParameterConfiguration represents the structure of a parameter in the Codacy API
type ParameterConfiguration struct {
	Name    string `json:"name"`
	Value   string `json:"value,omitempty"`
	Default string `json:"default,omitempty"`
}

// PatternDefinition represents the structure of a pattern in the Codacy API
type PatternDefinition struct {
	Id            string                   `json:"id"`
	Category      string                   `json:"category"`
	Level         string                   `json:"level"`
	SeverityLevel string                   `json:"severityLevel"`
	Enabled       bool                     `json:"enabled"`
	Parameters    []ParameterConfiguration `json:"parameters"`
	Title         string                   `json:"title"`
	Description   string                   `json:"description"`
	Explanation   string                   `json:"explanation"`
	Languages     []string                 `json:"languages"`
	TimeToFix     int                      `json:"timeToFix"`
}

// PatternConfiguration represents the structure of a pattern in the Codacy API
type PatternConfiguration struct {
	PatternDefinition PatternDefinition `json:"patternDefinition"`
	Parameters        []ParameterConfiguration
	Enabled           bool `json:"enabled"`
}

// PatternResponse represents the structure of a pattern in the API response
type PatternResponse struct {
	PatternDefinition PatternDefinition        `json:"patternDefinition"`
	Enabled           bool                     `json:"enabled"`
	IsCustom          bool                     `json:"isCustom"`
	Parameters        []ParameterConfiguration `json:"parameters"`
}

type SarifPatternConfiguration struct {
	UUID        string `json:"uuid"`
	ID          string `json:"id"`
	Name        string `json:"name"`
	Category    string `json:"category"`
	Description string `json:"description"`
	Level       string `json:"level"`
}
