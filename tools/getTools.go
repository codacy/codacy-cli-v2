package tools

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

func GetTools() ([]Tool, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Create a new GET request
	req, err := http.NewRequest("GET", "https://api.codacy.com/api/v3/tools", nil)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, errors.New("failed to get tools from Codacy API")
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}

	var response ToolsResponse
	_ = json.Unmarshal(body, &response)

	return response.Data, nil
}

type ToolsResponse struct {
	Data []Tool `json:"data"`
}

type Tool struct {
	Uuid         string `json:"uuid"`
	Name         string `json:"name"`
	Version      string `json:"version"`
	IsClientSide bool   `json:"isClientSide"`
	Settings     struct {
		Name                  string `json:"name"`
		IsEnabled             bool   `json:"isEnabled"`
		FollowsStandard       bool   `json:"followsStandard"`
		IsCustom              bool   `json:"isCustom"`
		HasConfigurationFile  bool   `json:"hasConfigurationFile"`
		UsesConfigurationFile bool   `json:"usesConfigurationFile"`
	} `json:"settings"`
}
