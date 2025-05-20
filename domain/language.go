package domain

// Language represents a language in the Codacy API
type Language struct {
	Name           string   `json:"name"`
	CodacyDefaults []string `json:"codacyDefaults"`
	Extensions     []string `json:"extensions"`
	Enabled        bool     `json:"enabled"`
	Detected       bool     `json:"detected"`
}

// LanguagesResponse represents the structure of the languages response
type LanguagesResponse struct {
	Languages []Language `json:"languages"`
}

// ToolLanguageInfo contains language and extension information for a tool
type ToolLanguageInfo struct {
	Name       string   `yaml:"name"`
	Languages  []string `yaml:"languages,flow"`
	Extensions []string `yaml:"extensions,flow"`
}

// LanguagesConfig represents the structure of the languages configuration file
type LanguagesConfig struct {
	Tools []ToolLanguageInfo `yaml:"tools"`
}
