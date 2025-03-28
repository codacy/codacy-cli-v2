package utils

import (
	"encoding/json"
	"log"
)

// ConvertPylintToSarif converts pylint JSON output to SARIF format
func ConvertPylintToSarif(jsonData []byte) []byte {
	var pylintResults []map[string]interface{}
	err := json.Unmarshal(jsonData, &pylintResults)
	if err != nil {
		log.Printf("Failed to parse pylint JSON output: %v", err)
		return createEmptySarif()
	}

	// Create SARIF structure
	sarif := map[string]interface{}{
		"version": "2.1.0",
		"$schema": "http://json.schemastore.org/sarif-2.1.0-rtm.5",
		"runs": []map[string]interface{}{
			{
				"tool": map[string]interface{}{
					"driver": map[string]interface{}{
						"name":           "Pylint",
						"informationUri": "https://pylint.org",
						"rules":          []map[string]interface{}{},
					},
				},
				"artifacts": []map[string]interface{}{},
				"results":   []map[string]interface{}{},
			},
		},
	}

	// Process pylint results
	run := sarif["runs"].([]map[string]interface{})[0]
	rules := map[string]bool{}
	artifacts := map[string]int{}

	// Track files for artifacts
	for _, result := range pylintResults {
		path, ok := result["path"].(string)
		if ok && path != "" {
			artifacts[path] = 0
		}
	}

	// Create artifact entries
	artifactsList := []map[string]interface{}{}
	i := 0
	for path := range artifacts {
		artifactsList = append(artifactsList, map[string]interface{}{
			"location": map[string]interface{}{
				"uri": path,
			},
		})
		artifacts[path] = i
		i++
	}
	run["artifacts"] = artifactsList

	// Process results
	results := []map[string]interface{}{}
	for _, pylintResult := range pylintResults {
		messageId, messageIdOk := pylintResult["message-id"].(string)
		symbol, symbolOk := pylintResult["symbol"].(string)
		message, messageOk := pylintResult["message"].(string)
		path, pathOk := pylintResult["path"].(string)
		line, lineOk := pylintResult["line"].(float64)
		column, columnOk := pylintResult["column"].(float64)

		if !messageIdOk || !symbolOk || !messageOk || !pathOk || !lineOk {
			continue
		}

		// Add rule if not already added
		ruleId := symbol
		if !rules[ruleId] {
			rules[ruleId] = true
			rule := map[string]interface{}{
				"id": ruleId,
				"shortDescription": map[string]interface{}{
					"text": messageId,
				},
			}
			run["tool"].(map[string]interface{})["driver"].(map[string]interface{})["rules"] =
				append(run["tool"].(map[string]interface{})["driver"].(map[string]interface{})["rules"].([]map[string]interface{}), rule)
		}

		// Create result
		result := map[string]interface{}{
			"ruleId": ruleId,
			"message": map[string]interface{}{
				"text": message,
			},
			"locations": []map[string]interface{}{
				{
					"physicalLocation": map[string]interface{}{
						"artifactLocation": map[string]interface{}{
							"uri":   path,
							"index": artifacts[path],
						},
						"region": map[string]interface{}{
							"startLine":   int(line),
							"startColumn": 1, // Default to 1 if column is not available
						},
					},
				},
			},
		}

		// Use column if available
		if columnOk {
			result["locations"].([]map[string]interface{})[0]["physicalLocation"].(map[string]interface{})["region"].(map[string]interface{})["startColumn"] = int(column)
		}

		results = append(results, result)
	}
	run["results"] = results

	// Convert to JSON
	sarifData, err := json.MarshalIndent(sarif, "", "  ")
	if err != nil {
		log.Printf("Failed to marshal SARIF data: %v", err)
		return createEmptySarif()
	}

	return sarifData
}

// createEmptySarif returns an empty SARIF structure as a byte array
func createEmptySarif() []byte {
	return []byte(`{
		"version": "2.1.0",
		"$schema": "http://json.schemastore.org/sarif-2.1.0-rtm.5",
		"runs": [
			{
				"tool": {
					"driver": {
						"name": "Pylint",
						"informationUri": "https://pylint.org"
					}
				},
				"results": []
			}
		]
	}`)
}
