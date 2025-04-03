package tools

type PatternParameterConfiguration struct {
	Name  string
	Value string
}

type PatternConfiguration struct {
	PatternId               string
	ParameterConfigurations []PatternParameterConfiguration
}

type ToolConfiguration struct {
	PatternsConfiguration []PatternConfiguration
}
