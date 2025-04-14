package domain

type ParameterConfiguration struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type PatternDefinition struct {
	Id       string `json:"id"`
	Category string `json:"category"`
	Level    string `json:"level"`
}

type PatternConfiguration struct {
	PatternDefinition PatternDefinition `json:"patternDefinition"`
	Enabled           bool              `json:"enabled"`
	Parameters        []ParameterConfiguration
}
