package main

import (
	"bytes"
	"codacy/cli-v2/config"
	cfg "codacy/cli-v2/config-file"
	"codacy/cli-v2/parser"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

// Entry point of the application
func main() {
	// Initialize configuration
	config.Init()

	configErr := cfg.ReadConfigFile(config.Config.ProjectConfigFile())
	// whenever there is no configuration file, the only command allowed to run is the 'init'
	if configErr != nil && os.Args[1] != "init" {
		fmt.Println("No configuration file was found, execute init command first.")
		return
	}

	// Execute CLI V2
	Execute()
}

// Execute function for CLI V2
func Execute() {
	// Parse command-line arguments
	var (
		projectToken   = flag.String("project-token", "", "the project-token to be used on the REST API")
		commitUuid     = flag.String("commit-uuid", "", "the commit uuid")
		repositoryBase = flag.String("repository-base", "", "base directory of the cloned repository")
	)
	flag.Parse()

	if *projectToken == "" || *commitUuid == "" || *repositoryBase == "" {
		fmt.Println("All parameters are required")
		flag.Usage()
		return
	}

	fmt.Println("Welcome to Codacy CLI V2")
	Process(*commitUuid, *projectToken, *repositoryBase)
}

// ParseSarifFile reads and parses the SARIF file
func ParseSarifFile(reportPath string) (*parser.Sarif, error) {
	data, err := ioutil.ReadFile(reportPath)
	if err != nil {
		return nil, fmt.Errorf("error reading SARIF file: %w", err)
	}
	var sarif parser.Sarif
	err = json.Unmarshal(data, &sarif)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling SARIF file: %w", err)
	}
	return &sarif, nil
}

// ExtractIssuesFromSarif extracts issues from SARIF data
func ExtractIssuesFromSarif(sarif *parser.Sarif) []parser.Issue {
	var issues []parser.Issue
	for _, run := range sarif.Runs {
		for _, result := range run.Results {
			for _, location := range result.Locations {
				issue := parser.Issue{
					Source:   strings.TrimPrefix(location.PhysicalLocation.ArtifactLocation.URI, "file://"),
					Line:     location.PhysicalLocation.Region.StartLine,
					Type:     result.RuleID,
					Message:  result.Message.Text,
					Level:    result.Level,
					Category: run.Tool.Driver.Rules[result.RuleIndex].ShortDescription.Text,
				}
				issues = append(issues, issue)
			}
		}
	}
	return issues
}

// GroupIssuesBySource groups issues by their source files
func GroupIssuesBySource(issues []parser.Issue) map[string]interface{} {
	groups := make(map[string]interface{})
	for _, issue := range issues {
		if _, ok := groups[issue.Source]; !ok {
			groups[issue.Source] = map[string]interface{}{
				"filename": issue.Source,
				"results":  []interface{}{},
			}
		}
		group := groups[issue.Source].(map[string]interface{})
		group["results"] = append(group["results"].([]interface{}), map[string]interface{}{
			"Issue": map[string]interface{}{
				"patternId": map[string]string{
					"value": issue.Type,
				},
				"filename": issue.Source,
				"message": map[string]string{
					"text": issue.Message,
				},
				"level":    issue.Level,
				"category": issue.Category,
				"location": map[string]interface{}{
					"LineLocation": map[string]int{
						"line": issue.Line,
					},
				},
			},
		})
	}
	return groups
}

// post sends a POST request to the given URL with the given headers and payload
func post(url string, headers map[string]string, payload interface{}) {
	client := &http.Client{}
	jsonPayload, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}
	for key, value := range headers {
		req.Header.Add(key, value)
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("Response:", resp.Status)
	fmt.Println("Response body:", string(body))
}

// postResults posts the issues to the remote server
func postResults(commitUuid string, payload []map[string]interface{}, projectToken string) {
	url := fmt.Sprintf("https://api.codacy.com/2.0/commit/%s/issuesRemoteResults", commitUuid)
	headers := map[string]string{
		"Content-Type":  "application/json",
		"Project-Token": projectToken,
	}
	post(url, headers, payload)
}

// resultsFinal posts final results to the remote server
func resultsFinal(commitUuid string, projectToken string) {
	url := fmt.Sprintf("https://api.codacy.com/2.0/commit/%s/resultsFinal", commitUuid)
	headers := map[string]string{
		"Content-Type":  "application/json",
		"Project-Token": projectToken,
	}
	post(url, headers, nil)
}

// Process handles the entire process of extracting issues and posting results
func Process(commitUuid, projectToken, repositoryBase string) {
	sarif, err := ParseSarifFile("eslint.sarif")
	if err != nil {
		fmt.Println("Error parsing SARIF file:", err)
		return
	}

	issues := ExtractIssuesFromSarif(sarif)
	adjustedIssues := make([]parser.Issue, len(issues))
	for i, issue := range issues {
		issue.Source = strings.TrimPrefix(issue.Source, "file://")
		if strings.HasPrefix(issue.Source, repositoryBase) {
			issue.Source = strings.TrimPrefix(issue.Source, repositoryBase)
		}
		adjustedIssues[i] = issue
	}
	groups := GroupIssuesBySource(adjustedIssues)

	payload := []map[string]interface{}{
		{
			"tool": "eslint",
			"issues": map[string]interface{}{
				"Success": map[string]interface{}{
					"results": []interface{}{},
				},
			},
		},
	}

	for _, group := range groups {
		payload[0]["issues"].(map[string]interface{})["Success"].(map[string]interface{})["results"] = append(payload[0]["issues"].(map[string]interface{})["Success"].(map[string]interface{})["results"].([]interface{}), group)
	}

	postResults(commitUuid, payload, projectToken)
	time.Sleep(5 * time.Second)
	resultsFinal(commitUuid, projectToken)
}
