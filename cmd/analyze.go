package cmd

import (
	"codacy/cli-v2/config"
	"codacy/cli-v2/tools"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var outputFile string
var toolToAnalyze string
var autoFix bool
var outputFormat string
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
	analyzeCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file for analysis results")
	analyzeCmd.Flags().StringVarP(&toolToAnalyze, "tool", "t", "", "Optional: Specific tool to run analysis with. If not specified, all configured tools will be run")
	analyzeCmd.Flags().StringVar(&outputFormat, "format", "", "Output format (use 'sarif' for SARIF format)")
	analyzeCmd.Flags().BoolVar(&autoFix, "fix", false, "Apply auto fix to your issues when available")
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

func runEslintAnalysis(workDirectory string, pathsToCheck []string, autoFix bool, outputFile string, outputFormat string) {
	eslint := config.Config.Tools()["eslint"]
	eslintInstallationDirectory := eslint.InstallDir
	nodeRuntime := config.Config.Runtimes()["node"]
	nodeBinary := nodeRuntime.Binaries["node"]

	tools.RunEslint(workDirectory, eslintInstallationDirectory, nodeBinary, pathsToCheck, autoFix, outputFile, outputFormat)
}

func runTrivyAnalysis(workDirectory string, pathsToCheck []string, outputFile string, outputFormat string) {
	trivy := config.Config.Tools()["trivy"]
	trivyBinary := trivy.Binaries["trivy"]

	err := tools.RunTrivy(workDirectory, trivyBinary, pathsToCheck, outputFile, outputFormat)
	if err != nil {
		log.Fatalf("Error running Trivy: %v", err)
	}
}

func runPmdAnalysis(workDirectory string, pathsToCheck []string, outputFile string, outputFormat string) {
	pmd := config.Config.Tools()["pmd"]
	pmdBinary := pmd.Binaries["pmd"]

	err := tools.RunPmd(workDirectory, pmdBinary, pathsToCheck, outputFile, outputFormat, "")
	if err != nil {
		log.Fatalf("Error running PMD: %v", err)
	}
}

func runPylintAnalysis(workDirectory string, pathsToCheck []string, outputFile string, outputFormat string) {
	pylint := config.Config.Tools()["pylint"]

	err := tools.RunPylint(workDirectory, pylint, pathsToCheck, outputFile, outputFormat)
	if err != nil {
		log.Fatalf("Error running Pylint: %v", err)
	}
}

var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Runs all configured linters.",
	Long:  "Runs all configured tools for code analysis. Use --tool flag to run a specific tool.",
	Run: func(cmd *cobra.Command, args []string) {
		workDirectory, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}

		// If a specific tool is specified, only run that tool
		if toolToAnalyze != "" {
			log.Printf("Running %s...\n", toolToAnalyze)
			runTool(workDirectory, toolToAnalyze, args, outputFile)
			return
		}

		// Run all configured tools
		tools := config.Config.Tools()
		if len(tools) == 0 {
			log.Fatal("No tools configured. Please run 'codacy-cli init' and 'codacy-cli install' first")
		}

		log.Println("Running all configured tools...")

		if outputFormat == "sarif" {
			// Create temporary directory for individual tool outputs
			tmpDir, err := os.MkdirTemp("", "codacy-analysis-*")
			if err != nil {
				log.Fatalf("Failed to create temporary directory: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			var sarifOutputs []string
			for toolName := range tools {
				log.Printf("Running %s...\n", toolName)
				tmpFile := filepath.Join(tmpDir, fmt.Sprintf("%s.sarif", toolName))
				runTool(workDirectory, toolName, args, tmpFile)
				sarifOutputs = append(sarifOutputs, tmpFile)
			}

			// Merge all SARIF outputs
			if err := mergeSarifOutputs(sarifOutputs, outputFile); err != nil {
				log.Fatalf("Failed to merge SARIF outputs: %v", err)
			}
		} else {
			// Run tools without merging outputs
			for toolName := range tools {
				log.Printf("Running %s...\n", toolName)
				runTool(workDirectory, toolName, args, outputFile)
			}
		}
	},
}

func runTool(workDirectory string, toolName string, args []string, outputFile string) {
	switch toolName {
	case "eslint":
		runEslintAnalysis(workDirectory, args, autoFix, outputFile, outputFormat)
	case "trivy":
		runTrivyAnalysis(workDirectory, args, outputFile, outputFormat)
	case "pmd":
		runPmdAnalysis(workDirectory, args, outputFile, outputFormat)
	case "pylint":
		runPylintAnalysis(workDirectory, args, outputFile, outputFormat)
	default:
		log.Printf("Warning: Unsupported tool: %s\n", toolName)
	}
}

func mergeSarifOutputs(inputFiles []string, outputFile string) error {
	var mergedSarif Sarif
	mergedSarif.Runs = make([]struct {
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
	}, 0)

	for _, file := range inputFiles {
		data, err := os.ReadFile(file)
		if err != nil {
			if os.IsNotExist(err) {
				// Skip if file doesn't exist (tool might have failed)
				continue
			}
			return fmt.Errorf("failed to read SARIF file %s: %w", file, err)
		}

		var sarif Sarif
		if err := json.Unmarshal(data, &sarif); err != nil {
			return fmt.Errorf("failed to parse SARIF file %s: %w", file, err)
		}

		mergedSarif.Runs = append(mergedSarif.Runs, sarif.Runs...)
	}

	// Create output file
	out, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer out.Close()

	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(mergedSarif); err != nil {
		return fmt.Errorf("failed to write merged SARIF: %w", err)
	}

	return nil
}
