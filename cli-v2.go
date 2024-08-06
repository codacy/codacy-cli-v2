package main

import (
	"bytes"
	"codacy/cli-v2/cmd"
	"codacy/cli-v2/config"
	cfg "codacy/cli-v2/config-file"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type Sarif struct {
	Runs []struct {
		Tool struct {
			Driver struct {
				Name    string `json:"name"`
				Version string `json:"version"`
				Rules   []struct {
					ID               string `json:"id"`
					HelpURI          string `json:"helpUri"`
					ShortDescription struct {
						Text string `json:"text"`
					} `json:"shortDescription"`
				} `json:"rules"`
			} `json:"driver"`
		} `json:"tool"`
		Artifacts []struct {
			Location struct {
				URI string `json:"uri"`
			} `json:"location"`
		} `json:"artifacts"`
		Results []struct {
			Level   string `json:"level"`
			Message struct {
				Text string `json:"text"`
			} `json:"message"`
			Locations []struct {
				PhysicalLocation struct {
					ArtifactLocation struct {
						URI   string `json:"uri"`
						Index int    `json:"index"`
					} `json:"artifactLocation"`
					Region struct {
						StartLine   int `json:"startLine"`
						StartColumn int `json:"startColumn"`
						EndLine     int `json:"endLine"`
						EndColumn   int `json:"endColumn"`
					} `json:"region"`
				} `json:"physicalLocation"`
			} `json:"locations"`
			RuleID    string `json:"ruleId"`
			RuleIndex int    `json:"ruleIndex"`
		} `json:"results"`
	} `json:"runs"`
}

type CodacyPattern struct {
	UUID        string `json:"uuid"`
	ID          string `json:"id"`
	Name        string `json:"name"`
	Category    string `json:"category"`
	Description string `json:"description"`
}

type CodacyPatternResponse struct {
	Data []CodacyPattern `json:"data"`
}

type CodacyIssue struct {
	Source   string `json:"source"`
	Line     int    `json:"line"`
	Type     string `json:"type"`
	Message  string `json:"message"`
	Level    string `json:"level"`
	Category string `json:"category"`
}

type CodacyPayload struct {
	Tool struct {
		Name string `json:"name"`
	} `json:"tool"`
	Issues []CodacyIssue `json:"issues"`
}

var codacyPatterns []CodacyPattern

func main() {
	fmt.Println("Running original CLI functionality...")
	// Original functionality
	config.Init()

	configErr := cfg.ReadConfigFile(config.Config.ProjectConfigFile())
	// whenever there is no configuration file, the only command allowed to run is the 'init'
	if configErr != nil && len(os.Args) > 1 && os.Args[1] != "init" {
		fmt.Println("No configuration file was found, execute init command first.")
		return
	}

	cmd.Execute()
}

func run() {
	var sarifPath, commitUuid, projectToken, apiToken string
	flag.StringVar(&sarifPath, "sarif-path", "", "Path to the SARIF report")
	flag.StringVar(&commitUuid, "commit-uuid", "", "Commit UUID")
	flag.StringVar(&projectToken, "project-token", "", "Project token for Codacy API")
	flag.StringVar(&apiToken, "api-token", "", "API token for Codacy")
	flag.Parse()

	if commitUuid != "" && projectToken != "" && apiToken != "" {
		fmt.Println("Processing SARIF and sending results...")
		processSarifAndSendResults(sarifPath, commitUuid, projectToken, apiToken)
		return
	}
}

func processSarifAndSendResults(sarifPath, commitUuid, projectToken, apiToken string) {
	fmt.Printf("Loading SARIF file from path: %s\n", sarifPath)
	// Load SARIF file
	sarifFile, err := os.Open(sarifPath)
	if err != nil {
		fmt.Printf("Error opening SARIF file: %v\n", err)
		os.Exit(1)
	}
	defer sarifFile.Close()

	var sarif Sarif
	fmt.Println("Parsing SARIF file...")
	err = json.NewDecoder(sarifFile).Decode(&sarif)
	if err != nil {
		fmt.Printf("Error parsing SARIF file: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Loading Codacy patterns...")
	// Load Codacy patterns
	loadCodacyPatterns(apiToken)

	fmt.Println("Processing SARIF results...")
	// Process SARIF results
	codacyIssues := processSarif(sarif)

	fmt.Println("Filtering issues...")
	// Filter issues
	filterIn, _ := filterPatterns(codacyIssues)

	fmt.Println("Sending results to Codacy...")
	// Send results to Codacy
	sendResults(filterIn, commitUuid, projectToken)
}

func extractCursorFromResponseBody(body io.Reader) (string, error) {
	var response struct {
		Pagination struct {
			Cursor string `json:"cursor"`
		} `json:"pagination"`
	}

	err := json.NewDecoder(body).Decode(&response)
	if err != nil {
		return "", fmt.Errorf("error decoding response body: %w", err)
	}
	return response.Pagination.Cursor, nil
}

func loadCodacyPatterns(apiToken string) {
	fmt.Println("Fetching Codacy patterns...")

	var cursor string
	const baseURL = "https://app.codacy.com/api/v3/tools/f8b29663-2cb2-498d-b923-a10c6a8c05cd/patterns"

	fmt.Printf("Requesting patterns from URL: %s\n", baseURL)
	req, err := http.NewRequest("GET", baseURL, nil)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		os.Exit(1)
	}
	req.Header.Set("api-token", apiToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Error fetching patterns: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error fetching patterns, status code: %d\n", resp.StatusCode)
		os.Exit(1)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		os.Exit(1)
	}

	var patternResp struct {
		Data []CodacyPattern `json:"data"`
	}
	fmt.Println("Decoding pattern response...")
	err = json.Unmarshal(body, &patternResp)
	if err != nil {
		fmt.Printf("Error decoding pattern response: %v\n", err)
		os.Exit(1)
	}

	patternCount := len(patternResp.Data)
	fmt.Printf("Loaded %d patterns.\n", patternCount)
	codacyPatterns = append(codacyPatterns, patternResp.Data...)

	cursor, err = extractCursorFromResponseBody(bytes.NewReader(body))
	if err != nil {
		fmt.Printf("Error extracting cursor: %v\n", err)
		os.Exit(1)
	}

	for {
		url := baseURL
		if cursor != "" {
			url += "?cursor=" + cursor
		}

		fmt.Printf("Requesting patterns from URL: %s\n", url)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			fmt.Printf("Error creating request: %v\n", err)
			os.Exit(1)
		}
		req.Header.Set("api-token", apiToken)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Printf("Error fetching patterns: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			fmt.Printf("Error fetching patterns, status code: %d\n", resp.StatusCode)
			os.Exit(1)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Error reading response body: %v\n", err)
			os.Exit(1)
		}

		var patternResp struct {
			Data []CodacyPattern `json:"data"`
		}
		fmt.Println("Decoding pattern response...")
		err = json.Unmarshal(body, &patternResp)
		if err != nil {
			fmt.Printf("Error decoding pattern response: %v\n", err)
			os.Exit(1)
		}

		patternCount := len(patternResp.Data)
		fmt.Printf("Loaded %d patterns.\n", patternCount)
		codacyPatterns = append(codacyPatterns, patternResp.Data...)

		cursor, err = extractCursorFromResponseBody(bytes.NewReader(body))
		if err != nil {
			fmt.Printf("Error extracting cursor: %v\n", err)
			os.Exit(1)
		}

		if cursor == "" {
			fmt.Println("No more patterns to load.")
			break
		}
	}
}

func processSarif(sarif Sarif) []CodacyIssue {
	fmt.Println("Processing SARIF results...")
	var codacyIssues []CodacyIssue

	for _, run := range sarif.Runs {
		for _, result := range run.Results {
			for _, location := range result.Locations {
				source := strings.Replace(location.PhysicalLocation.ArtifactLocation.URI, "file://", "", 1)

				codacyIssues = append(codacyIssues, CodacyIssue{
					Source:  source,
					Line:    location.PhysicalLocation.Region.StartLine,
					Type:    result.RuleID,
					Message: result.Message.Text,
					Level:   result.Level,
				})
			}
		}
	}

	fmt.Printf("Processed %d issues from SARIF.\n", len(codacyIssues))
	fmt.Print(codacyIssues)
	return codacyIssues
}

func filterPatterns(issues []CodacyIssue) ([]CodacyIssue, []CodacyIssue) {
	fmt.Println("Filtering issues based on Codacy patterns...")
	var filterIn, filterOut []CodacyIssue

	codacyPatternMap := make(map[string]CodacyPattern)
	for _, pattern := range codacyPatterns {
		codacyPatternMap[pattern.ID] = pattern
	}

	for _, issue := range issues {
		if pattern, exists := codacyPatternMap[issue.Type]; exists {
			issue.Type = pattern.ID
			filterIn = append(filterIn, issue)
		} else {
			filterOut = append(filterOut, issue)
		}
	}

	fmt.Printf("Filtered %d issues in and %d issues out.\n", len(filterIn), len(filterOut))
	return filterIn, filterOut
}

func sendResults(issues []CodacyIssue, commitUuid, projectToken string) {
	fmt.Println("Sending issues to Codacy...")
	payload := CodacyPayload{
		Tool: struct {
			Name string `json:"name"`
		}{Name: "ESLint"},
		Issues: issues,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("Error marshaling payload: %v\n", err)
		os.Exit(1)
	}

	url := fmt.Sprintf("https://api.codacy.com/2.0/commit/%s/issuesRemoteResults", commitUuid)
	fmt.Printf("Sending results to URL: %s\n", url)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		os.Exit(1)
	}
	req.Header.Set("content-type", "application/json")
	req.Header.Set("project-token", projectToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Error sending results: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error sending results, status code: %d\n", resp.StatusCode)
		os.Exit(1)
	}

	fmt.Println("Results sent successfully.")
}
