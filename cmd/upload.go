package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var apiToken string
var provider string
var owner string
var repository string

func init() {
	uploadResultsCmd.Flags().StringVarP(&sarifPath, "sarif-path", "s", "", "Path to the SARIF report")
	uploadResultsCmd.MarkPersistentFlagRequired("sarif-path")
	uploadResultsCmd.Flags().StringVarP(&commitUuid, "commit-uuid", "c", "", "Commit UUID")
	uploadResultsCmd.Flags().StringVarP(&projectToken, "project-token", "t", "", "Project token for Codacy API")
	uploadResultsCmd.Flags().StringVarP(&apiToken, "api-token", "a", "", "API token for Codacy API")
	uploadResultsCmd.Flags().StringVarP(&provider, "provider", "p", "", "Provider (gh, gl, bb)")
	uploadResultsCmd.Flags().StringVarP(&owner, "owner", "o", "", "Owner/Organization")
	uploadResultsCmd.Flags().StringVarP(&repository, "repository", "r", "", "Repository")

	rootCmd.AddCommand(uploadResultsCmd)
}

var uploadResultsCmd = &cobra.Command{
	Use:   "upload",
	Short: "Uploads a sarif file to Codacy",
	Long:  "YADA",
	Run: func(cmd *cobra.Command, args []string) {
		processSarifAndSendResults(sarifPath, commitUuid, projectToken, apiToken)
	},
}

func processSarifAndSendResults(sarifPath string, commitUUID string, projectToken string, apiToken string) {
	if projectToken == "" && apiToken == "" && provider == "" && repository == "" {
		fmt.Println("Error: api-token, provider and repository are required when project-token is not provided")
		os.Exit(1)
	}
	//Load SARIF file
	fmt.Printf("Loading SARIF file from path: %s\n", sarifPath)
	sarifFile, err := os.Open(sarifPath)
	if err != nil {
		fmt.Printf("Error opening SARIF file: %v\n", err)
		panic("panic")
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
	payloads := processSarif(sarif)
	if projectToken != "" {
		for _, payload := range payloads {
			sendResultsWithProjectToken(payload, commitUUID, projectToken)
		}

		resultsFinalWithProjectToken(commitUUID, projectToken)
	} else {
		for _, payload := range payloads {
			sendResultsWithAPIToken(payload, commitUUID, apiToken, provider, owner, repository)
		}
		resultsFinalWithAPIToken(commitUUID, apiToken, provider, owner, repository)
	}

}

func processSarif(sarif Sarif) [][]map[string]interface{} {
	var codacyIssues []map[string]interface{}
	var payloads [][]map[string]interface{}

	for _, run := range sarif.Runs {
		var toolName = getToolName(strings.ToLower(run.Tool.Driver.Name), run.Tool.Driver.Version)
		tool, patterns := loadsToolAndPatterns(toolName)
		for _, result := range run.Results {
			modifiedType := tool.Prefix + strings.Replace(result.RuleID, "/", "_", -1)
			pattern := getPatternByID(patterns, modifiedType)
			if pattern == nil {
				fmt.Printf("Rule '%s' doesn't have a direct mapping on Codacy\n", modifiedType)
				continue
			}
			for _, location := range result.Locations {
				codacyIssues = append(codacyIssues, map[string]interface{}{
					"source":   location.PhysicalLocation.ArtifactLocation.URI,
					"line":     location.PhysicalLocation.Region.StartLine,
					"type":     modifiedType,
					"message":  result.Message.Text,
					"level":    pattern.Level,
					"category": pattern.Category,
				})
			}
		}
		var results []map[string]interface{}
		// Iterate through run.Artifacts and create entries in the results object
		for _, artifact := range run.Artifacts {
			if artifact.Location.URI != "" {
				results = append(results, map[string]interface{}{
					"filename": artifact.Location.URI,
					"results":  []map[string]interface{}{},
				})
			}
		}
		for _, obj := range codacyIssues {
			source := obj["source"].(string)
			issue := map[string]interface{}{
				"patternId": map[string]string{
					"value": obj["type"].(string),
				},
				"filename": source,
				"message": map[string]string{
					"text": obj["message"].(string),
				},
				"level": obj["level"].(string),
				//"category": obj["category"].(string),
				"location": map[string]interface{}{
					"LineLocation": map[string]int{
						"line": obj["line"].(int),
					},
				},
			}

			// Check if we already have an entry for this filename
			found := false
			for i, result := range results {
				if result["filename"] == source {
					// If we do, append this issue to its results
					results[i]["results"] = append(results[i]["results"].([]map[string]interface{}), map[string]interface{}{"Issue": issue})
					found = true
					break
				}
			}

			// If we don't, create a new entry
			if !found {
				results = append(results, map[string]interface{}{
					"filename": source,
					"results":  []map[string]interface{}{{"Issue": issue}},
				})
			}

		}

		payload := []map[string]interface{}{
			{
				"tool": toolName,
				"issues": map[string]interface{}{
					"Success": map[string]interface{}{
						"results": results,
					},
				},
			},
		}
		payloads = append(payloads, payload)
	}

	return payloads
}

func resultsFinalWithProjectToken(commitUUID string, projectToken string) {
	url := fmt.Sprintf("https://api.codacy.com/2.0/commit/%s/resultsFinal", commitUUID)
	req, _ := http.NewRequest("POST", url, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("project-token", projectToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	fmt.Println("Response:", resp.Status)
	fmt.Println("Response Body:", string(body))
}

func resultsFinalWithAPIToken(commitUUID string, apiToken string, provider string, owner string, repository string) {
	url := fmt.Sprintf("https://api.codacy.com/2.0/%s/%s/%s/commit/%s/resultsFinal", provider, owner, repository, commitUUID)
	req, _ := http.NewRequest("POST", url, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-token", apiToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	fmt.Println("Response:", resp.Status)
	fmt.Println("Response Body:", string(body))
}

func getPatternByID(patterns []Pattern, patternID string) *Pattern {
	for _, p := range patterns {
		if strings.EqualFold(p.ID, patternID) {
			return &p
		}
	}
	return nil
}

func getMajorVersion(version string) int {
	parts := strings.Split(version, ".")
	if len(parts) > 0 {
		major, err := strconv.Atoi(parts[0])
		if err != nil {
			fmt.Println("Error converting major version to integer:", err)
			return -1
		}
		return major
	}
	return -1
}

func sendResultsWithProjectToken(payload []map[string]interface{}, commitUUID string, projectToken string) {

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("Error marshaling payload: %v\n", err)
		panic("panic")
	}

	url := fmt.Sprintf("https://api.codacy.com/2.0/commit/%s/issuesRemoteResults", commitUUID)
	fmt.Printf("Sending results to URL: %s\n", url)
	fmt.Println("Payload:", string(payloadBytes))
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
}

func sendResultsWithAPIToken(payload []map[string]interface{}, commitUUID string, apiToken string, provider string, owner string, repository string) {

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("Error marshaling payload: %v\n", err)
		panic("panic")
	}

	url := fmt.Sprintf("https://api.codacy.com/2.0/%s/%s/%s/commit/%s/issuesRemoteResults", provider, owner, repository, commitUUID)
	fmt.Printf("Sending results to URL: %s\n", url)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		os.Exit(1)
	}
	req.Header.Set("content-type", "application/json")
	req.Header.Set("api-token", apiToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Error sending results: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error sending results, status code: %d\n", resp.StatusCode)
		os.Exit(1)
	} else {
		fmt.Println("Results sent successfully")
	}
}
