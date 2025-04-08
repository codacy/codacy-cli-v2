package tools

import (
	"codacy/cli-v2/config"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

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

	cliSupportedTools := config.Config.Tools()

	// Filter enabled tools
	var enabledTools []Tool
	for _, tool := range response.Data {
		if tool.Settings.Enabled {
			if _, exists := cliSupportedTools[strings.ToLower(tool.Name)]; exists {
				enabledTools = append(enabledTools, tool)
			}
		}
	}

	return enabledTools, nil
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
