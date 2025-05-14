package codacyclient

import (
	"codacy/cli-v2/domain"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

const timeout = 10 * time.Second
const CodacyApiBase = "https://app.codacy.com"

func getRequest(url string, initFlags domain.InitFlags) ([]byte, error) {
	client := &http.Client{
		Timeout: timeout,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("api-token", initFlags.ApiToken)

	resp, err := client.Do(req)

	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Println("Error:", url)
		return nil, errors.New("Failed in request to " + url)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return body, nil
}

// handlePaginationGeneric fetches all paginated results of type T using a provided fetchPage function.
//   - baseURL: the base URL for the paginated API
//   - initialCursor: the initial cursor (empty string for first page)
//   - fetchPage: a function that fetches a page and returns ([]T, nextCursor, error)
//
// Returns a slice of all results of type T and any error encountered.
func handlePaginationGeneric[T any](
	baseURL string,
	initialCursor string,
	fetchPage func(url string) ([]T, string, error),
) ([]T, error) {
	var allResults []T
	cursor := initialCursor
	firstRequest := true

	for {
		pageURL := baseURL
		if !firstRequest && cursor != "" {
			if pageURL[len(pageURL)-1] == '?' || pageURL[len(pageURL)-1] == '&' {
				pageURL += fmt.Sprintf("cursor=%s", cursor)
			} else {
				pageURL += fmt.Sprintf("&cursor=%s", cursor)
			}
		}
		firstRequest = false

		results, nextCursor, err := fetchPage(pageURL)
		if err != nil {
			return nil, err
		}
		allResults = append(allResults, results...)

		if nextCursor == "" {
			break
		}
		cursor = nextCursor
	}

	return allResults, nil
}

func GetDefaultToolPatternsConfig(initFlags domain.InitFlags, toolUUID string) ([]domain.PatternConfiguration, error) {
	baseURL := fmt.Sprintf("%s/api/v3/tools/%s/patterns?enabled=true", CodacyApiBase, toolUUID)

	fetchPage := func(url string) ([]domain.PatternConfiguration, string, error) {
		response, err := getRequest(url, initFlags)
		if err != nil {
			return nil, "", fmt.Errorf("failed to get patterns page: %w", err)
		}

		var objmap map[string]json.RawMessage
		if err := json.Unmarshal(response, &objmap); err != nil {
			return nil, "", fmt.Errorf("failed to unmarshal response: %w", err)
		}

		var patterns []domain.PatternDefinition
		if err := json.Unmarshal(objmap["data"], &patterns); err != nil {
			return nil, "", fmt.Errorf("failed to unmarshal patterns: %w", err)
		}

		patternConfigurations := make([]domain.PatternConfiguration, len(patterns))
		for i, pattern := range patterns {
			patternConfigurations[i] = domain.PatternConfiguration{
				PatternDefinition: pattern,
				Parameters:        pattern.Parameters,
				Enabled:           pattern.Enabled,
			}
		}

		var pagination domain.Pagination
		if objmap["pagination"] != nil {
			if err := json.Unmarshal(objmap["pagination"], &pagination); err != nil {
				return nil, "", fmt.Errorf("failed to unmarshal pagination: %w", err)
			}
		}

		return patternConfigurations, pagination.Cursor, nil
	}

	return handlePaginationGeneric(baseURL, "", fetchPage)
}
