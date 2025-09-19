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

// CodacyApiBase is the base URL for the Codacy API
var CodacyApiBase = "https://app.codacy.com"

func getRequest(url string, apiToken string) ([]byte, error) {
	client := &http.Client{
		Timeout: timeout,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if apiToken != "" {
		req.Header.Set("api-token", apiToken)
	}

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
	response, err := getRequest(url, initFlags.ApiToken)
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

// parseDefaultPatternConfigurations parses the response body into pattern configurations
func parseDefaultPatternConfigurations(response []byte) ([]domain.PatternConfiguration, string, error) {
	var objmap map[string]json.RawMessage
	if err := json.Unmarshal(response, &objmap); err != nil {
		return nil, "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	var patternResponses []domain.PatternDefinition
	if err := json.Unmarshal(objmap["data"], &patternResponses); err != nil {
		return nil, "", fmt.Errorf("failed to unmarshal patterns: %w", err)
	}

	patternConfigurations := make([]domain.PatternConfiguration, len(patternResponses))
	for i, patternDef := range patternResponses {
		patternConfigurations[i] = domain.PatternConfiguration{
			PatternDefinition: patternDef,
			Parameters:        patternDef.Parameters,
			Enabled:           patternDef.Enabled,
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

// parsePatternConfigurations parses the response body into pattern configurations
func parsePatternConfigurations(response []byte) ([]domain.PatternConfiguration, string, error) {

	var objmap map[string]json.RawMessage
	if err := json.Unmarshal(response, &objmap); err != nil {
		return nil, "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	var patternResponses []domain.PatternResponse
	if err := json.Unmarshal(objmap["data"], &patternResponses); err != nil {
		return nil, "", fmt.Errorf("failed to unmarshal patterns: %w", err)
	}

	patternConfigurations := make([]domain.PatternConfiguration, len(patternResponses))
	for i, patternResp := range patternResponses {
		patternConfigurations[i] = domain.PatternConfiguration{
			PatternDefinition: patternResp.PatternDefinition,
			Parameters:        patternResp.Parameters,
			Enabled:           patternResp.Enabled,
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

// GetDefaultToolPatternsConfig fetches the default patterns for a tool
func GetDefaultToolPatternsConfig(initFlags domain.InitFlags, toolUUID string, onlyEnabledPatterns bool) ([]domain.PatternConfiguration, error) {
	return GetDefaultToolPatternsConfigWithCodacyAPIBase(CodacyApiBase, initFlags, toolUUID, onlyEnabledPatterns)
}

// GetDefaultToolPatternsConfigWithCodacyAPIBase fetches the default patterns for a tool, and a base api url
func GetDefaultToolPatternsConfigWithCodacyAPIBase(codacyAPIBaseURL string, initFlags domain.InitFlags, toolUUID string, onlyEnabledPatterns bool) ([]domain.PatternConfiguration, error) {
	baseURL := fmt.Sprintf("%s/api/v3/tools/%s/patterns", codacyAPIBaseURL, toolUUID)
	if onlyEnabledPatterns {
		baseURL += "?enabled=true"
	}

	allPaterns, err := getAllPages(baseURL, initFlags, parseDefaultPatternConfigurations)
	if err != nil {
		return nil, err
	}

	onlyRecommendedPatterns := make([]domain.PatternConfiguration, 0)
	for _, pattern := range allPaterns {
		if pattern.PatternDefinition.Enabled {
			onlyRecommendedPatterns = append(onlyRecommendedPatterns, pattern)
		}
	}

	return onlyRecommendedPatterns, nil
}

// GetRepositoryToolPatterns fetches the patterns for a tool in a repository
func GetRepositoryToolPatterns(initFlags domain.InitFlags, toolUUID string) ([]domain.PatternConfiguration, error) {
	baseURL := fmt.Sprintf("%s/api/v3/analysis/organizations/%s/%s/repositories/%s/tools/%s/patterns?enabled=true",
		CodacyApiBase,
		initFlags.Provider,
		initFlags.Organization,
		initFlags.Repository,
		toolUUID)

	result, err := getAllPages(baseURL, initFlags, parsePatternConfigurations)
	return result, err
}

// GetRepositoryTools fetches the tools for a repository
func GetRepositoryTools(initFlags domain.InitFlags) ([]domain.Tool, error) {
	baseURL := fmt.Sprintf("%s/api/v3/analysis/organizations/%s/%s/repositories/%s/tools",
		CodacyApiBase,
		initFlags.Provider,
		initFlags.Organization,
		initFlags.Repository)

	bodyResponse, err := getRequest(baseURL, initFlags.ApiToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository tools: %w", err)
	}

	var toolsResponse domain.ToolsResponse

	err = json.Unmarshal(bodyResponse, &toolsResponse)
	if err != nil {
		fmt.Println("Error unmarshaling response:", err)
		return nil, err
	}

	// Get global tools with languages to populate the Languages field
	globalTools, err := GetToolsVersions()
	if err != nil {
		fmt.Printf("Warning: Failed to get global tools for languages: %v\n", err)
		return toolsResponse.Data, nil // Return repository tools without languages
	}

	// Create a map of UUID to languages from global tools
	uuidToLanguages := make(map[string][]string)
	for _, globalTool := range globalTools {
		uuidToLanguages[globalTool.Uuid] = globalTool.Languages
	}

	// Populate Languages field in repository tools
	for i := range toolsResponse.Data {
		if languages, exists := uuidToLanguages[toolsResponse.Data[i].Uuid]; exists {
			toolsResponse.Data[i].Languages = languages
		}
	}

	return toolsResponse.Data, nil
}

// GetToolsVersions fetches the tools versions
func GetToolsVersions() ([]domain.Tool, error) {
	baseURL := fmt.Sprintf("%s/api/v3/tools", CodacyApiBase)

	bodyResponse, err := getRequest(baseURL, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get tool versions: %w", err)
	}

	var toolsResponse domain.ToolsResponse
	err = json.Unmarshal(bodyResponse, &toolsResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return toolsResponse.Data, nil
}

// GetRepositoryLanguages fetches the languages for a repository
func GetRepositoryLanguages(initFlags domain.InitFlags) ([]domain.RepositoryLanguage, error) {
	baseURL := fmt.Sprintf("%s/api/v3/organizations/%s/%s/repositories/%s/settings/languages",
		CodacyApiBase,
		initFlags.Provider,
		initFlags.Organization,
		initFlags.Repository)

	bodyResponse, err := getRequest(baseURL, initFlags.ApiToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository languages: %w", err)
	}

	var languagesResponse domain.LanguagesResponse
	err = json.Unmarshal(bodyResponse, &languagesResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return languagesResponse.Languages, nil
}

// GetLanguageTools fetches the default language file extensions from the API
func GetLanguageTools() ([]domain.LanguageTool, error) {
	baseURL := fmt.Sprintf("%s/api/v3/languages/tools", CodacyApiBase)

	bodyResponse, err := getRequest(baseURL, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get language tools: %w", err)
	}

	var languageToolsResponse domain.LanguageToolsResponse
	err = json.Unmarshal(bodyResponse, &languageToolsResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal language tools response: %w", err)
	}

	return languageToolsResponse.Data, nil
}
