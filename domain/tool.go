package domain

type ToolsResponse struct {
	Data []Tool `json:"data"`
}

type Tool struct {
	Uuid     string `json:"uuid"`
	Name     string `json:"name"`
	Version  string `json:"version"`
	Settings struct {
		Enabled               bool `json:"isEnabled"`
		HasConfigurationFile  bool `json:"hasConfigurationFile"`
		UsesConfigurationFile bool `json:"usesConfigurationFile"`
	} `json:"settings"`
}
