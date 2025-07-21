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
	ShortName string   `json:"shortName"`
	Prefix    string   `json:"prefix"`
	Languages []string `json:"languages"`
	Settings  struct {
		Enabled               bool `json:"isEnabled"`
		HasConfigurationFile  bool `json:"hasConfigurationFile"`
		UsesConfigurationFile bool `json:"usesConfigurationFile"`
	} `json:"settings"`
}

const (
	ESLint       string = "f8b29663-2cb2-498d-b923-a10c6a8c05cd"
	ESLint9      string = "2a30ab97-477f-4769-8b88-af596ce7a94c"
	Trivy        string = "2fd7fbe0-33f9-4ab3-ab73-e9b62404e2cb"
	PMD          string = "9ed24812-b6ee-4a58-9004-0ed183c45b8f"
	PMD7         string = "ed7e8287-707d-485a-a0cb-e211004432c2"
	PyLint       string = "31677b6d-4ae0-4f56-8041-606a8d7a8e61"
	DartAnalyzer string = "d203d615-6cf1-41f9-be5f-e2f660f7850f"
	Semgrep      string = "6792c561-236d-41b7-ba5e-9d6bee0d548b"
	Lizard       string = "76348462-84b3-409a-90d3-955e90abfb87"
	Revive       string = "bd81d1f4-1406-402d-9181-1274ee09f1aa"
	LicenseSim   string = "b7e1c2a4-5f3d-4e2a-9c8b-1a2b3c4d5e6f" // generated UUID
)

type ToolInfo struct {
	Name     string
	Priority int // lower means newer
}

// SupportedToolsMetadata group same family tools by name
var SupportedToolsMetadata = map[string]ToolInfo{
	ESLint9:      {Name: "eslint", Priority: 0},
	ESLint:       {Name: "eslint", Priority: 1},
	PMD7:         {Name: "pmd", Priority: 0},
	PMD:          {Name: "pmd", Priority: 1},
	PyLint:       {Name: "pylint", Priority: 0},
	Trivy:        {Name: "trivy", Priority: 0},
	DartAnalyzer: {Name: "dartanalyzer", Priority: 0},
	Lizard:       {Name: "lizard", Priority: 0},
	Semgrep:      {Name: "semgrep", Priority: 0},
	Revive:       {Name: "revive", Priority: 0},
	LicenseSim:   {Name: "license-sim", Priority: 0},
}
