package domain

type ParameterConfiguration struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type PatternDefinition struct {
	Id         string                   `json:"id"`
	Parameters []ParameterConfiguration `json:"parameters"`
}

type PatternConfiguration struct {
	PatternDefinition PatternDefinition `json:"patternDefinition"`
	Parameters        []ParameterConfiguration
}
