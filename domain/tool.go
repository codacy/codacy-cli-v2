package domain

// ToolsResponse represents the structure of the tools response
type ToolsResponse struct {
	Data []Tool `json:"data"`
}

// Tool represents a tool in the Codacy API
type Tool struct {
	Uuid      string   `json:"uuid"`
	Name      string   `json:"name"`
	Version   string   `json:"version"`
	Languages []string `json:"languages"`
	Settings  struct {
		Enabled               bool `json:"isEnabled"`
		HasConfigurationFile  bool `json:"hasConfigurationFile"`
		UsesConfigurationFile bool `json:"usesConfigurationFile"`
	} `json:"settings"`
}
