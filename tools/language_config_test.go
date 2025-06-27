package tools

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	codacyclient "codacy/cli-v2/codacy-client"
	"codacy/cli-v2/config"
	"codacy/cli-v2/domain"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestGetRepositoryLanguages(t *testing.T) {
	// Test cases
	tests := []struct {
		name           string
		response       []map[string]interface{}
		expectedResult map[string][]string
		expectedError  bool
	}{
		{
			name: "Single enabled and detected language",
			response: []map[string]interface{}{
				{
					"name":           "JavaScript",
					"codacyDefaults": []string{".js", ".jsx", ".jsm"},
					"extensions":     []string{".js", ".vue"},
					"enabled":        true,
					"detected":       true,
				},
				{
					"name":           "Python",
					"codacyDefaults": []string{".py"},
					"extensions":     []string{},
					"enabled":        false,
					"detected":       true,
				},
			},
			expectedResult: map[string][]string{
				"JavaScript": {".js", ".jsx", ".jsm", ".vue"},
			},
			expectedError: false,
		},
		{
			name: "Multiple enabled and detected languages",
			response: []map[string]interface{}{
				{
					"name":           "JavaScript",
					"codacyDefaults": []string{".js", ".jsx"},
					"extensions":     []string{".js"},
					"enabled":        true,
					"detected":       true,
				},
				{
					"name":           "Python",
					"codacyDefaults": []string{".py"},
					"extensions":     []string{".pyi"},
					"enabled":        true,
					"detected":       true,
				},
			},
			expectedResult: map[string][]string{
				"JavaScript": {".js", ".jsx"},
				"Python":     {".py", ".pyi"},
			},
			expectedError: false,
		},
		{
			name: "No enabled languages",
			response: []map[string]interface{}{
				{
					"name":           "JavaScript",
					"codacyDefaults": []string{".js"},
					"extensions":     []string{},
					"enabled":        false,
					"detected":       true,
				},
			},
			expectedResult: map[string][]string{},
			expectedError:  false,
		},
		{
			name: "No detected languages",
			response: []map[string]interface{}{
				{
					"name":           "JavaScript",
					"codacyDefaults": []string{".js"},
					"extensions":     []string{},
					"enabled":        true,
					"detected":       false,
				},
			},
			expectedResult: map[string][]string{},
			expectedError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify the request
				assert.Equal(t, "GET", r.Method)
				assert.Equal(t, "/api/v3/organizations/gh/org/repositories/repo/settings/languages", r.URL.Path)
				assert.Equal(t, "test-token", r.Header.Get("api-token"))

				// Write the response
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"languages": tt.response,
				})
			}))
			defer server.Close()

			// Create a test function that uses the test server URL
			testGetRepositoryLanguages := func(apiToken string, provider string, organization string, repository string) (map[string][]string, error) {
				client := &http.Client{
					Timeout: 10 * time.Second,
				}

				url := fmt.Sprintf("%s/api/v3/organizations/%s/%s/repositories/%s/settings/languages",
					server.URL,
					provider,
					organization,
					repository)

				// Create a new GET request
				req, err := http.NewRequest("GET", url, nil)
				if err != nil {
					return nil, fmt.Errorf("failed to create request: %w", err)
				}

				// Set the API token header
				req.Header.Set("api-token", apiToken)

				// Send the request
				resp, err := client.Do(req)
				if err != nil {
					return nil, fmt.Errorf("failed to send request: %w", err)
				}
				defer resp.Body.Close()

				if resp.StatusCode != 200 {
					return nil, fmt.Errorf("failed to get repository languages: status code %d", resp.StatusCode)
				}

				// Read the response body
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, fmt.Errorf("failed to read response body: %w", err)
				}

				// Define the response structure
				type LanguageResponse struct {
					Name           string   `json:"name"`
					CodacyDefaults []string `json:"codacyDefaults"`
					Extensions     []string `json:"extensions"`
					Enabled        bool     `json:"enabled"`
					Detected       bool     `json:"detected"`
				}

				type LanguagesResponse struct {
					Languages []LanguageResponse `json:"languages"`
				}

				var response LanguagesResponse
				if err := json.Unmarshal(body, &response); err != nil {
					return nil, fmt.Errorf("failed to unmarshal response: %w", err)
				}

				// Create map to store language name -> combined extensions
				result := make(map[string][]string)

				// Filter and process languages
				for _, lang := range response.Languages {
					if lang.Enabled && lang.Detected {
						// Combine and deduplicate extensions
						extensions := make(map[string]struct{})
						for _, ext := range lang.CodacyDefaults {
							extensions[ext] = struct{}{}
						}
						for _, ext := range lang.Extensions {
							extensions[ext] = struct{}{}
						}

						// Convert map to slice
						extSlice := make([]string, 0, len(extensions))
						for ext := range extensions {
							extSlice = append(extSlice, ext)
						}

						// Sort extensions for consistent order
						sort.Strings(extSlice)

						// Add to result map
						result[lang.Name] = extSlice
					}
				}

				return result, nil
			}

			// Call the test function
			result, err := testGetRepositoryLanguages("test-token", "gh", "org", "repo")

			// Check the results
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Sort expected extensions for each language
				for lang := range tt.expectedResult {
					sort.Strings(tt.expectedResult[lang])
				}
				assert.Equal(t, tt.expectedResult, result)
			}
		})
	}
}

func TestCreateLanguagesConfigFile_ExtensionsFromRepository(t *testing.T) {
	tempDir := t.TempDir()

	// Create cli-config.yaml with remote mode to test remote mode behavior
	cliConfigPath := filepath.Join(tempDir, "cli-config.yaml")
	cliConfigContent := "mode: remote\n"
	err := os.WriteFile(cliConfigPath, []byte(cliConfigContent), 0644)
	assert.NoError(t, err)

	// Create a temporary config instance with the temp directory
	tempConfig := config.NewConfigType("", tempDir, filepath.Join(tempDir, "cache"))

	// Save the original global config
	originalConfig := config.Config

	// Override the global config temporarily
	config.Config = *tempConfig
	defer func() { config.Config = originalConfig }()

	// Mock API server for getRepositoryLanguages and GetLanguageTools
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if strings.Contains(r.URL.Path, "/languages/tools") {
			// Mock GetLanguageTools response
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": []map[string]interface{}{
					{
						"name":           "JavaScript",
						"fileExtensions": []string{".js", ".jsx", ".mjs"},
					},
					{
						"name":           "TypeScript",
						"fileExtensions": []string{".ts", ".tsx"},
					},
					{
						"name":           "Python",
						"fileExtensions": []string{".py"},
					},
					{
						"name":           "Java",
						"fileExtensions": []string{".java"},
					},
					{
						"name":           "Apex",
						"fileExtensions": []string{".cls", ".trigger"},
					},
					{
						"name":           "Scala",
						"fileExtensions": []string{".scala"},
					},
					{
						"name":           "Ruby",
						"fileExtensions": []string{".rb", ".gemspec"},
					},
				},
			})
		} else {
			// Mock getRepositoryLanguages response
			json.NewEncoder(w).Encode(map[string]interface{}{
				"languages": []map[string]interface{}{
					{
						"name":           "JavaScript",
						"codacyDefaults": []string{".js", ".jsx"},
						"extensions":     []string{".js", ".vue"},
						"enabled":        true,
						"detected":       true,
					},
					{
						"name":           "Python",
						"codacyDefaults": []string{".py"},
						"extensions":     []string{".testPy"},
						"enabled":        true,
						"detected":       true,
					},
					{
						"name":           "Apex",
						"codacyDefaults": []string{".cls"},
						"extensions":     []string{".app", ".trigger"},
						"enabled":        true,
						"detected":       true,
					},
					{
						"name":           "Scala",
						"codacyDefaults": []string{".scala"},
						"extensions":     []string{},
						"enabled":        true,
						"detected":       true,
					},
					{
						"name":           "Ruby",
						"codacyDefaults": []string{".rb"},
						"extensions":     []string{".gemspec"},
						"enabled":        true,
						"detected":       true,
					},
				},
			})
		}
	}))
	defer server.Close()

	// Patch CodacyApiBase to use the test server
	oldBase := codacyclient.CodacyApiBase
	codacyclient.CodacyApiBase = server.URL
	defer func() { codacyclient.CodacyApiBase = oldBase }()

	apiTools := []domain.Tool{
		{
			Uuid:      "eslint-uuid",
			Languages: []string{"JavaScript", "TypeScript"},
		},
		{
			Uuid:      "pylint-uuid",
			Languages: []string{"Python"},
		},
		{
			Uuid:      "pmd-uuid",
			Languages: []string{"Java", "JavaScript", "Apex", "Scala", "Ruby"},
		},
	}
	toolIDMap := map[string]string{
		"eslint-uuid": "eslint",
		"pylint-uuid": "pylint",
		"pmd-uuid":    "pmd",
	}

	initFlags := domain.InitFlags{
		ApiToken:     "test-token",
		Provider:     "gh",
		Organization: "org",
		Repository:   "repo",
	}
	err = CreateLanguagesConfigFile(apiTools, tempDir, toolIDMap, initFlags)
	assert.NoError(t, err)

	// Read and unmarshal the generated YAML
	data, err := os.ReadFile(tempDir + "/languages-config.yaml")
	assert.NoError(t, err)

	var config domain.LanguagesConfig
	err = yaml.Unmarshal(data, &config)
	assert.NoError(t, err)

	// Check that extensions are correct for each tool
	eslint := findTool(config.Tools, "eslint")
	assert.ElementsMatch(t, []string{".js", ".jsx", ".vue"}, eslint.Extensions)
	pylint := findTool(config.Tools, "pylint")
	assert.ElementsMatch(t, []string{".py", ".testPy"}, pylint.Extensions)
	pmd := findTool(config.Tools, "pmd")
	assert.ElementsMatch(t, []string{".cls", ".app", ".trigger", ".scala", ".rb", ".gemspec", ".js", ".jsx", ".vue"}, pmd.Extensions)
}

func findTool(tools []domain.ToolLanguageInfo, name string) domain.ToolLanguageInfo {
	for _, t := range tools {
		if t.Name == name {
			return t
		}
	}
	return domain.ToolLanguageInfo{}
}
