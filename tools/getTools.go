package tools

import (
	"encoding/json"
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
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
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
	Uuid    string `json:"uuid"`
	Name    string `json:"name"`
	Version string `json:"version"`
}
