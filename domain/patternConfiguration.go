package domain

type ParameterConfiguration struct {
	Name    string `json:"name"`
	Value   string `json:"value,omitempty"`
	Default string `json:"default,omitempty"`
}

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

type PatternConfiguration struct {
	PatternDefinition PatternDefinition `json:"patternDefinition"`
	Parameters        []ParameterConfiguration
	Enabled           bool `json:"enabled"`
	IsCustom          bool `json:"isCustom"`
}
