package tools

import (
	"codacy/cli-v2/domain"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// FetchDefaultEnabledPatterns fetches default patterns from Codacy API for a given tool UUID
func FetchDefaultEnabledPatterns(toolUUID string) ([]domain.PatternDefinition, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Fetch default patterns from Codacy API
	url := fmt.Sprintf("https://app.codacy.com/api/v3/tools/%s/patterns", toolUUID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch default patterns: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("failed to get default patterns from Codacy API: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Get default patterns
	var apiResponse struct {
		Data       []domain.PatternDefinition `json:"data"`
		Pagination struct {
			Limit int `json:"limit"`
			Total int `json:"total"`
		} `json:"pagination"`
	}
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal API response: %w", err)
	}

	// Filter out disabled patterns
	var enabledPatterns []domain.PatternDefinition
	for _, pattern := range apiResponse.Data {
		if pattern.Enabled {
			enabledPatterns = append(enabledPatterns, pattern)
		}
	}

	return enabledPatterns, nil
}
