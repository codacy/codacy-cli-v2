package pylint

import "codacy/cli-v2/tools/types"

// PatternDefaultParameters contains the default parameters for Pylint patterns
var PatternDefaultParameters = map[string][]types.ParameterConfiguration{
	"R0914": {
		{
			Name:  "max-locals",
			Value: "15",
		},
	},
	"C0301": {
		{
			Name:  "max-line-length",
			Value: "120",
		},
	},
	"C0102": {
		{
			Name:  "bad-names",
			Value: "foo,bar,baz,toto,tutu,tata",
		},
	},
	"C0103": {
		{
			Name:  "argument-rgx",
			Value: "[a-z_][a-z0-9_]{2,30}$",
		},
		{
			Name:  "attr-rgx",
			Value: "[a-z_][a-z0-9_]{2,30}$",
		},
		{
			Name:  "class-rgx",
			Value: "[A-Z_][a-zA-Z0-9]+$",
		},
		{
			Name:  "const-rgx",
			Value: "(([A-Z_][A-Z0-9_]*)|(__.*__))$",
		},
		{
			Name:  "function-rgx",
			Value: "[a-z_][a-z0-9_]{2,30}$",
		},
		{
			Name:  "method-rgx",
			Value: "[a-z_][a-z0-9_]{2,30}$",
		},
		{
			Name:  "module-rgx",
			Value: "(([a-z_][a-z0-9_]*)|([A-Z][a-zA-Z0-9]+))$",
		},
		{
			Name:  "variable-rgx",
			Value: "[a-z_][a-z0-9_]{2,30}$",
		},
		{
			Name:  "inlinevar-rgx",
			Value: "[A-Za-z_][A-Za-z0-9_]*$",
		},
		{
			Name:  "class-attribute-rgx",
			Value: "([A-Za-z_][A-Za-z0-9_]{2,30}|(__.*__))$",
		},
	},
}
