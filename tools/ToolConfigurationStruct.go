package tools

type PatternParameterConfiguration struct {
	Name  string
	Value string
}

type PatternConfiguration struct {
	PatternId                string
	ParamenterConfigurations []PatternParameterConfiguration
}

type ToolConfiguration struct {
	PatternsConfiguration []PatternConfiguration
}
