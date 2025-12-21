package cmd

import (
	"bytes"
	"codacy/cli-v2/config"
	"codacy/cli-v2/domain"
	"codacy/cli-v2/plugins"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
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
		processSarifAndSendResults(sarifPath, commitUuid, projectToken, apiToken, config.Config.Tools())
	},
}

var sarifShortNameMap = map[string]string{
	// The keys here MUST match the exact string found in run.Tool.Driver.Name
	"ESLint (deprecated)": "eslint",
	"ESLint":              "eslint-8",
	"ESLint9":             "eslint-9",
	"PMD":                 "pmd",
	"PMD7":                "pmd-7",
	"Trivy":               "trivy",
	"Pylint":              "pylintpython3",
	"dartanalyzer":        "dartanalyzer",
	"Semgrep":             "semgrep",
	"Lizard":              "lizard",
	"revive":              "revive",
}

func getToolShortName(fullName string) string {
	if shortName, ok := sarifShortNameMap[fullName]; ok {
		return shortName
	}
	// Fallback: Use the original name if no mapping is found
	return fullName
}

func getRelativePath(baseDir string, fullURI string) string {
	localPath := fullURI
	if u, err := url.Parse(fullURI); err == nil && u.Scheme == "file" {
		// url.Path extracts the local path component correctly and may be URL-encoded
		if decodedPath, err := url.PathUnescape(u.Path); err == nil {
			localPath = decodedPath
		} else {
			localPath = u.Path
		}
	}

	baseDirNormalized := filepath.FromSlash(baseDir)
	localPathNormalized := filepath.FromSlash(localPath)

	relativePath, err := filepath.Rel(baseDirNormalized, localPathNormalized)
	if err != nil {
		// Fallback to the normalized absolute path if calculation fails
		fmt.Printf("Warning: Could not get relative path for '%s' relative to '%s': %v. Using absolute path.\n", localPathNormalized, baseDirNormalized, err)
		return localPathNormalized
	}

	return filepath.FromSlash(relativePath)
}

func processSarifAndSendResults(sarifPath string, commitUUID string, projectToken string, apiToken string, tools map[string]*plugins.ToolInfo) {
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
	payloads := processSarif(sarif, tools)
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

func processSarif(sarif Sarif, tools map[string]*plugins.ToolInfo) [][]map[string]interface{} {
	var codacyIssues []map[string]interface{}
	var payloads [][]map[string]interface{}

	baseDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting current working directory: %v\n", err)
		os.Exit(1)
	}

	for _, run := range sarif.Runs {
		//getToolName will take care of mapping sarif tool names to codacy tool names
		//especially for eslint and pmd that have multiple versions
		var toolName = getToolName(strings.ToLower(run.Tool.Driver.Name), run.Tool.Driver.Version)
		tool, patterns := loadsToolAndPatterns(toolName, false)

		for _, result := range run.Results {
			modifiedType := tool.Prefix + strings.Replace(result.RuleID, "/", "_", -1)
			pattern := getPatternByID(patterns, modifiedType)
			if pattern == nil {
				fmt.Printf("Rule '%s' doesn't have a direct mapping on Codacy\n", modifiedType)
				continue
			}
			for _, location := range result.Locations {

				fullURI := location.PhysicalLocation.ArtifactLocation.URI
				relativePath := getRelativePath(baseDir, fullURI)

				issue := map[string]interface{}{
					"source":   relativePath,
					"line":     location.PhysicalLocation.Region.StartLine,
					"type":     pattern.ID,
					"message":  result.Message.Text,
					"level":    pattern.Level,
					"category": pattern.Category,
				}

				// Only add sourceId for tools that need it
				if toolInfo, exists := tools[toolName]; exists && toolInfo.NeedsSourceIDUpload {
					issue["sourceId"] = result.RuleID
				}

				codacyIssues = append(codacyIssues, issue)
			}
		}
		var results []map[string]interface{}
		// Iterate through run.Artifacts and create entries in the results object
		for _, artifact := range run.Artifacts {
			if artifact.Location.URI != "" {

				fullURI := artifact.Location.URI
				relativePath := getRelativePath(baseDir, fullURI)

				results = append(results, map[string]interface{}{
					"filename": relativePath,
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

			// Only add sourceId for tools that need it
			if toolInfo, exists := tools[toolName]; exists && toolInfo.NeedsSourceIDUpload {
				issue["sourceId"] = obj["sourceId"].(string)
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
		var toolShortName = getToolShortName(toolName)
		payload := []map[string]interface{}{
			{
				"tool": toolShortName,
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

func getPatternByID(patterns []domain.PatternConfiguration, patternID string) *domain.SarifPatternConfiguration {
	var sarifPatterns []domain.SarifPatternConfiguration
	for _, p := range patterns {
		sarifPatterns = append(sarifPatterns, domain.SarifPatternConfiguration{
			UUID:        p.PatternDefinition.Id,
			ID:          p.PatternDefinition.Id,
			Category:    p.PatternDefinition.Category,
			Description: p.PatternDefinition.Description,
			Level:       p.PatternDefinition.Level,
		})
	}
	for _, p := range sarifPatterns {
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
