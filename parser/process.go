package parser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

const (
	reportFileName = "eslint.sarif"
	baseUrl        = "https://api.codacy.com"
)

func postResults(commitUuid string, payload []map[string]interface{}, projectToken string) {
	url := fmt.Sprintf("%s/2.0/commit/%s/issuesRemoteResults", baseUrl, commitUuid)
	headers := map[string]string{
		"Content-Type":  "application/json",
		"Project-Token": projectToken,
	}
	post(url, headers, payload)
}

func resultsFinal(commitUuid string, projectToken string) {
	url := fmt.Sprintf("%s/2.0/commit/%s/resultsFinal", baseUrl, commitUuid)
	headers := map[string]string{
		"Content-Type":  "application/json",
		"Project-Token": projectToken,
	}
	post(url, headers, nil)
}

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

func Process(commitUuid, projectToken, repositoryBase string) {
	sarif, err := ParseSarifFile(reportFileName)
	if err != nil {
		fmt.Println("Error parsing SARIF file:", err)
		return
	}

	issues := ExtractIssuesFromSarif(sarif)
	adjustedIssues := make([]Issue, len(issues))
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
