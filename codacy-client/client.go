package codacyclient

import (
	"codacy/cli-v2/domain"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
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
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("api-token", initFlags.ApiToken)

	resp, err := client.Do(req)

	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("request to %s failed with status %d", url, resp.StatusCode)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return body, nil
}

// GetPage fetches a single page of results from the API and returns the data and next cursor
func GetPage[T any](
	url string,
	initFlags domain.InitFlags,
	parser func([]byte) ([]T, string, error),
) ([]T, string, error) {
	response, err := getRequest(url, initFlags)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get page: %w", err)
	}
	return parser(response)
}

// getAllPages fetches all pages of results from a paginated API endpoint
//   - baseURL: the base URL for the paginated API
//   - initFlags: the API token and other flags for authentication
//   - parser: a function that parses the response body into the desired type and returns ([]T, nextCursor, error)
//
// Returns a slice of all results of type T and any error encountered.
func getAllPages[T any](
	baseURL string,
	initFlags domain.InitFlags,
	parser func([]byte) ([]T, string, error),
) ([]T, error) {
	var allResults []T
	cursor := ""
	firstRequest := true

	for {
		pageURL := baseURL
		if !firstRequest && cursor != "" {
			u, err := url.Parse(pageURL)
			if err != nil {
				return nil, err
			}
			q := u.Query()
			q.Set("cursor", cursor)
			u.RawQuery = q.Encode()
			pageURL = u.String()
		}
		firstRequest = false

		results, nextCursor, err := GetPage[T](pageURL, initFlags, parser)
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

// parsePatternConfigurations parses the response body into pattern configurations
func parsePatternConfigurations(response []byte) ([]domain.PatternConfiguration, string, error) {
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

func GetDefaultToolPatternsConfig(initFlags domain.InitFlags, toolUUID string) ([]domain.PatternConfiguration, error) {
	baseURL := fmt.Sprintf("%s/api/v3/tools/%s/patterns?enabled=true", CodacyApiBase, toolUUID)
	return getAllPages(baseURL, initFlags, parsePatternConfigurations)
}

func GetRepositoryToolPatterns(initFlags domain.InitFlags, toolUUID string) ([]domain.PatternConfiguration, error) {
	baseURL := fmt.Sprintf("%s/api/v3/analysis/organizations/%s/%s/repositories/%s/tools/%s/patterns?enabled=true",
		CodacyApiBase,
		initFlags.Provider,
		initFlags.Organization,
		initFlags.Repository,
		toolUUID)
	return getAllPages(baseURL, initFlags, parsePatternConfigurations)
}
