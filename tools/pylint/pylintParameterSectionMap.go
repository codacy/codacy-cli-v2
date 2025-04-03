package pylint

// GetParameterSection returns the section name for a given parameter
func GetParameterSection(paramName string) *string {
	section := parameterSectionMap[paramName]
	if section == "" {
		return nil
	}
	return &section
}

// parameterSectionMap maps parameter names to their section names
var parameterSectionMap = map[string]string{
	// BASIC section - naming conventions
	"required-attributes": "BASIC",
	"bad-functions":       "BASIC",
	"good-names":          "BASIC",
	"bad-names":           "BASIC",
	"name-group":          "BASIC",
	"include-naming-hint": "BASIC",

	// BASIC section - regex patterns
	"function-rgx":              "BASIC",
	"function-name-hint":        "BASIC",
	"variable-rgx":              "BASIC",
	"variable-name-hint":        "BASIC",
	"const-rgx":                 "BASIC",
	"const-name-hint":           "BASIC",
	"attr-rgx":                  "BASIC",
	"attr-name-hint":            "BASIC",
	"argument-rgx":              "BASIC",
	"argument-name-hint":        "BASIC",
	"class-attribute-rgx":       "BASIC",
	"class-attribute-name-hint": "BASIC",
	"inlinevar-rgx":             "BASIC",
	"inlinevar-name-hint":       "BASIC",
	"class-rgx":                 "BASIC",
	"class-name-hint":           "BASIC",
	"module-rgx":                "BASIC",
	"module-name-hint":          "BASIC",
	"method-rgx":                "BASIC",
	"method-name-hint":          "BASIC",
	"no-docstring-rgx":          "BASIC",
	"docstring-min-length":      "BASIC",

	// SPELLING section
	"spelling-dict":                "SPELLING",
	"spelling-ignore-words":        "SPELLING",
	"spelling-private-dict-file":   "SPELLING",
	"spelling-store-unknown-words": "SPELLING",

	// SIMILARITIES section
	"min-similarity-lines": "SIMILARITIES",
	"ignore-comments":      "SIMILARITIES",
	"ignore-docstrings":    "SIMILARITIES",
	"ignore-imports":       "SIMILARITIES",

	// LOGGING section
	"logging-modules": "LOGGING",

	// FORMAT section
	"max-line-length":             "FORMAT",
	"ignore-long-lines":           "FORMAT",
	"single-line-if-stmt":         "FORMAT",
	"no-space-check":              "FORMAT",
	"max-module-lines":            "FORMAT",
	"indent-string":               "FORMAT",
	"indent-after-paren":          "FORMAT",
	"expected-line-ending-format": "FORMAT",

	// MISCELLANEOUS section
	"notes": "MISCELLANEOUS",

	// TYPECHECK section
	"ignore-mixin-members": "TYPECHECK",
	"ignored-modules":      "TYPECHECK",
	"ignored-classes":      "TYPECHECK",
	"zope":                 "TYPECHECK",
	"generated-members":    "TYPECHECK",

	// CLASSES section
	"ignore-iface-methods":                  "CLASSES",
	"defining-attr-methods":                 "CLASSES",
	"valid-classmethod-first-arg":           "CLASSES",
	"valid-metaclass-classmethod-first-arg": "CLASSES",
	"exclude-protected":                     "CLASSES",

	// DESIGN section
	"max-args":               "DESIGN",
	"ignored-argument-names": "DESIGN",
	"max-locals":             "DESIGN",
	"max-returns":            "DESIGN",
	"max-branches":           "DESIGN",
	"max-statements":         "DESIGN",
	"max-parents":            "DESIGN",
	"max-attributes":         "DESIGN",
	"min-public-methods":     "DESIGN",
	"max-public-methods":     "DESIGN",

	// IMPORTS section
	"deprecated-modules": "IMPORTS",
	"import-graph":       "IMPORTS",
	"ext-import-graph":   "IMPORTS",
	"int-import-graph":   "IMPORTS",

	// EXCEPTIONS section
	"overgeneral-exceptions": "EXCEPTIONS",
}
