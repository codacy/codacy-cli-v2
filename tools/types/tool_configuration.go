package types

// ToolConfiguration represents the configuration for a tool
type ToolConfiguration struct {
	Uuid      string                 `json:"uuid"`
	IsEnabled bool                   `json:"isEnabled"`
	Patterns  []PatternConfiguration `json:"patterns"`
}

// PatternConfiguration represents a pattern configuration from the API
type PatternConfiguration struct {
	InternalId string                   `json:"internalId"`
	Parameters []ParameterConfiguration `json:"parameters"`
}

// ParameterConfiguration represents a parameter configuration from the API
type ParameterConfiguration struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
