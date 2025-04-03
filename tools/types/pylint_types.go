package types

// PylintPatternConfiguration represents a Pylint pattern configuration
type PylintPatternConfiguration struct {
	Id         string
	Parameters []PylintPatternParameterConfiguration
}

// PatternParameterConfiguration represents a parameter configuration for a Pylint pattern
type PylintPatternParameterConfiguration struct {
	Name        string
	Value       string
	SectionName *string
}
