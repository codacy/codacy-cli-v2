package domain

// RepositoryLanguage represents a language in the Codacy API
type RepositoryLanguage struct {
	Name           string   `json:"name"`
	CodacyDefaults []string `json:"codacyDefaults"`
	Extensions     []string `json:"extensions"`
	DefaultFiles   []string `json:"defaultFiles"`
	Enabled        bool     `json:"enabled"`
	Detected       bool     `json:"detected"`
}

// LanguagesResponse represents the structure of the languages response
type LanguagesResponse struct {
	Languages []RepositoryLanguage `json:"languages"`
}

// Language represents a processed language with combined extensions and files
type Language struct {
	Name       string
	Extensions []string
	Files      []string
}

// LanguageTool represents a language tool with its file extensions from the API
type LanguageTool struct {
	Name           string   `json:"name"`
	FileExtensions []string `json:"fileExtensions"`
	Files          []string `json:"files"`
}

// LanguageToolsResponse represents the structure of the language tools API response
type LanguageToolsResponse struct {
	Data []LanguageTool `json:"data"`
}

// ToolLanguageInfo contains language and extension information for a tool
type ToolLanguageInfo struct {
	Name       string   `yaml:"name"`
	Languages  []string `yaml:"languages,flow"`
	Extensions []string `yaml:"extensions,flow"`
	Files      []string `yaml:"files,flow"`
}

// LanguagesConfig represents the structure of the languages configuration file
type LanguagesConfig struct {
	Tools []ToolLanguageInfo `yaml:"tools"`
}
