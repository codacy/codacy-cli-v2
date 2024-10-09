package cmd

import (
	"codacy/cli-v2/config"
	"codacy/cli-v2/tools"
	"codacy/cli-v2/utils"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var outputFile string
var toolToAnalyze string
var autoFix bool
var doNewPr bool
var sarifPath string
var commitUuid string
var projectToken string

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

type CodacyIssue struct {
	Source   string `json:"source"`
	Line     int    `json:"line"`
	Type     string `json:"type"`
	Message  string `json:"message"`
	Level    string `json:"level"`
	Category string `json:"category"`
}

type Tool struct {
	UUID      string `json:"uuid"`
	ShortName string `json:"shortName"`
	Prefix    string `json:"prefix"`
}

type Pattern struct {
	UUID        string `json:"uuid"`
	ID          string `json:"id"`
	Name        string `json:"name"`
	Category    string `json:"category"`
	Description string `json:"description"`
	Level       string `json:"level"`
}

func init() {
	analyzeCmd.Flags().StringVarP(&outputFile, "output", "o", "", "output file for the results")
	analyzeCmd.Flags().StringVarP(&toolToAnalyze, "tool", "t", "", "Which tool to run analysis with")
	analyzeCmd.Flags().BoolVarP(&autoFix, "fix", "f", false, "Apply auto fix to your issues when available")
	analyzeCmd.Flags().BoolVar(&doNewPr, "new-pr", false, "Create a new PR on GitHub containing the fixed issues")
	rootCmd.AddCommand(analyzeCmd)
}

func loadsToolAndPatterns(toolName string) (Tool, []Pattern) {
	var toolsURL = "https://app.codacy.com/api/v3/tools"

	req, err := http.NewRequest("GET", toolsURL, nil)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		panic("panic")
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Error fetching patterns: %v\n", err)
		panic("panic")
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var toolsResponse struct {
		Data []Tool `json:"data"`
	}
	json.Unmarshal(body, &toolsResponse)
	var tool Tool
	for _, t := range toolsResponse.Data {
		if t.ShortName == toolName {
			tool = t
			break
		}
	}
	// TO DO - PANIC
	//if tool == nil {
	//	return nil, nil
	//}
	var patterns []Pattern
	var hasNext bool = true
	cursor := ""
	client := &http.Client{}

	for hasNext {
		baseURL := fmt.Sprintf("https://app.codacy.com/api/v3/tools/%s/patterns?limit=1000%s", tool.UUID, cursor)
		req, _ := http.NewRequest("GET", baseURL, nil)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Error:", err)
			break
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)

		var patternsResponse struct {
			Data       []Pattern `json:"data"`
			Pagination struct {
				Cursor string `json:"cursor"`
			} `json:"pagination"`
		}
		json.Unmarshal(body, &patternsResponse)
		patterns = append(patterns, patternsResponse.Data...)
		hasNext = patternsResponse.Pagination.Cursor != ""
		if hasNext {
			cursor = "&cursor=" + patternsResponse.Pagination.Cursor
		}
	}
	return tool, patterns
}

func getToolName(toolName string, version string) string {

	if toolName == "eslint" {
		majorVersion := getMajorVersion(version)
		switch majorVersion {
		case 7:
			return "eslint"
		case 8:
			return "eslint-8"
		case 9:
			return "eslint-9"
		}

	}

	return toolName
}

var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Runs all linters.",
	Long:  "Runs all tools for all runtimes.",
	Run: func(cmd *cobra.Command, args []string) {
		workDirectory, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}

		// TODO add more tools here
		switch toolToAnalyze {
		case "eslint":
			// nothing
		case "":
			log.Fatal("You need to specify a tool to run analysis with, e.g., '--tool eslint'", toolToAnalyze)
		default:
			log.Fatal("Trying to run unsupported tool: ", toolToAnalyze)
		}

		// can't create a new PR if there will be no changes/fixed issues
		if doNewPr && !autoFix {
			log.Fatal("Can't create a new PR with fixes without fixing issues. Use the '--fix' option.")
		} else if doNewPr {
			failIfThereArePendingChanges()
		}

		eslint := config.Config.Tools()["eslint"]
		eslintInstallationDirectory := eslint.Info()["installDir"]
		nodeRuntime := config.Config.Runtimes()["node"]
		nodeBinary := nodeRuntime.Info()["node"]

		log.Printf("Running %s...\n", toolToAnalyze)
		if outputFile != "" {
			log.Println("Output will be available at", outputFile)
		}

		tools.RunEslint(workDirectory, eslintInstallationDirectory, nodeBinary, args, autoFix, outputFile)

		if doNewPr {
			utils.CreatePr(false)
		}
	},
}

func failIfThereArePendingChanges() {
	cmd := exec.Command("git", "status", "--porcelain")
	out, _ := cmd.Output()

	if string(out) != "" {
		log.Fatal("There are pending changes, cannot proceed. Commit your pending changes.")
	}
}
