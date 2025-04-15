package tools

import (
	"codacy/cli-v2/plugins"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

func enrichToolsWithVersion(tools []Tool) ([]Tool, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Create a new GET request
	req, err := http.NewRequest("GET", "https://api.codacy.com/api/v3/tools", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to get tools from Codacy API: status code %d", resp.StatusCode)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var response ToolsResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Create a map of tool UUIDs to versions
	versionMap := make(map[string]string)
	for _, tool := range response.Data {
		versionMap[tool.Uuid] = tool.Version
	}

	// Enrich the input tools with versions
	for i, tool := range tools {
		if version, exists := versionMap[tool.Uuid]; exists {
			tools[i].Version = version
		}
	}

	return tools, nil
}

func GetRepositoryTools(codacyBase string, apiToken string, provider string, organization string, repository string) ([]Tool, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	url := fmt.Sprintf("%s/api/v3/analysis/organizations/%s/%s/repositories/%s/tools",
		codacyBase,
		provider,
		organization,
		repository)

	// Create a new GET request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}

	// Set the headers
	req.Header.Set("api-token", apiToken)

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.New("failed to get repository tools from Codacy API")
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}

	var response ToolsResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		fmt.Println("Error unmarshaling response:", err)
		return nil, err
	}

	supportedTools, err := plugins.GetSupportedTools()
	if err != nil {
		return nil, err
	}

	// Filter enabled tools
	var enabledTools []Tool
	for _, tool := range response.Data {
		if tool.Settings.Enabled {
			if _, exists := supportedTools[strings.ToLower(tool.Name)]; exists {
				enabledTools = append(enabledTools, tool)
			}
		}
	}

	return enrichToolsWithVersion(enabledTools)
}

type ToolsResponse struct {
	Data []Tool `json:"data"`
}

type Tool struct {
	Uuid     string `json:"uuid"`
	Name     string `json:"name"`
	Version  string `json:"version"`
	Settings struct {
		Enabled        bool `json:"isEnabled"`
		UsesConfigFile bool `json:"hasConfigurationFile"`
	} `json:"settings"`
}

// FilterToolsByConfigUsage filters out tools that use their own configuration files
// Returns only tools that need configuration to be generated for them (UsesConfigFile = false)
func FilterToolsByConfigUsage(tools []Tool) []Tool {
	var filtered []Tool
	for _, tool := range tools {

		if !tool.Settings.UsesConfigFile {
			filtered = append(filtered, tool)
		} else {
			fmt.Printf("Skipping config generation for %s - configured to use repo's config file\n", tool.Name)
		}
	}
	return filtered
}
