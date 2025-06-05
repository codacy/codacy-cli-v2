package cmd

import (
	"codacy/cli-v2/config"
	"codacy/cli-v2/domain"
	"codacy/cli-v2/plugins"
	"codacy/cli-v2/tools"
	"codacy/cli-v2/tools/lizard"
	reviveTool "codacy/cli-v2/tools/revive"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"codacy/cli-v2/utils"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var outputFile string
var toolsToAnalyzeParam string
var autoFix bool
var outputFormat string
var sarifPath string
var commitUuid string
var projectToken string

// LanguagesConfig represents the structure of the languages configuration file
type LanguagesConfig struct {
	Tools []struct {
		Name       string   `yaml:"name" json:"name"`
		Languages  []string `yaml:"languages" json:"languages"`
		Extensions []string `yaml:"extensions" json:"extensions"`
	} `yaml:"tools" json:"tools"`
}

// LoadLanguageConfig loads the language configuration from the file
func LoadLanguageConfig() (*LanguagesConfig, error) {
	// First, try to load the YAML config
	yamlPath := filepath.Join(config.Config.ToolsConfigDirectory(), "languages-config.yaml")

	// Check if the YAML file exists
	if _, err := os.Stat(yamlPath); err == nil {
		data, err := os.ReadFile(yamlPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read languages configuration file: %w", err)
		}

		var config LanguagesConfig
		if err := yaml.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("failed to parse YAML languages configuration file: %w", err)
		}

		return &config, nil
	}

	// If YAML file doesn't exist, try the JSON config for backward compatibility
	jsonPath := filepath.Join(config.Config.ToolsConfigDirectory(), "languages-config.json")

	// Check if the JSON file exists
	if _, err := os.Stat(jsonPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("languages configuration file not found: neither %s nor %s exists", yamlPath, jsonPath)
	}

	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read JSON languages configuration file: %w", err)
	}

	var config LanguagesConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse JSON languages configuration file: %w", err)
	}

	return &config, nil
}

// GetFileExtension extracts the file extension from a path
func GetFileExtension(filePath string) string {
	return strings.ToLower(filepath.Ext(filePath))
}

// IsToolSupportedForFile checks if a tool supports a given file based on its extension
func IsToolSupportedForFile(toolName string, filePath string, langConfig *LanguagesConfig) bool {
	if langConfig == nil {
		// If no language config is available, assume all tools are supported
		return true
	}

	fileExt := GetFileExtension(filePath)
	if fileExt == "" {
		// If file has no extension, assume tool is supported
		return true
	}

	for _, tool := range langConfig.Tools {
		if tool.Name == toolName {
			// If tool has no extensions defined, assume it supports all files
			if len(tool.Extensions) == 0 {
				return true
			}

			// Check if file extension is supported by this tool
			for _, ext := range tool.Extensions {
				if strings.EqualFold(ext, fileExt) {
					return true
				}
			}

			// Extension not found in tool's supported extensions
			return false
		}
	}

	// If tool not found in config, assume it's supported
	return true
}

// FilterToolsByLanguageSupport filters tools by language support for the given files
func FilterToolsByLanguageSupport(tools map[string]*plugins.ToolInfo, files []string) map[string]*plugins.ToolInfo {

	if len(files) == 0 || files[0] == "." {
		// If no files specified or current directory, return all tools
		return tools
	}

	langConfig, err := LoadLanguageConfig()
	if err != nil {
		log.Printf("Warning: Failed to load language configuration: %v. Running all tools.", err)
		return tools
	}

	result := make(map[string]*plugins.ToolInfo)

	// For each tool, check if it supports at least one of the files
	for toolName, toolInfo := range tools {
		supported := false

		for _, file := range files {
			if IsToolSupportedForFile(toolName, file, langConfig) {
				supported = true
				break
			}
		}

		if supported {
			result[toolName] = toolInfo
		} else {
			log.Printf("Skipping %s as it doesn't support the specified file(s)", toolName)
		}
	}

	return result
}

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
	analyzeCmd.Flags().StringVarP(&toolsToAnalyzeParam, "tool", "t", "", "Which tool to run analysis with. If not specified, all configured tools will be run")
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

func runEslintAnalysis(workDirectory string, pathsToCheck []string, autoFix bool, outputFile string, outputFormat string) error {
	// Ensure ESLint tool is configured and installed
	eslint := config.Config.Tools()["eslint"]
	if eslint == nil || !config.Config.IsToolInstalled("eslint", eslint) {
		if eslint == nil {
			fmt.Println("Eslint tool configuration not found, adding and installing...")
		} else {
			fmt.Println("Eslint tool is not installed, installing...")
		}

		err := config.InstallTool("eslint", eslint, "")
		if err != nil {
			return fmt.Errorf("failed to install eslint: %w", err)
		}

		// Get the updated tool info after installation
		eslint = config.Config.Tools()["eslint"]
		if eslint == nil {
			return fmt.Errorf("eslint tool configuration still not found after installation")
		}
		fmt.Println("Eslint tool installed successfully")
	}

	// Ensure node runtime is available
	nodeRuntime := config.Config.Runtimes()["node"]
	if nodeRuntime == nil {
		return fmt.Errorf("node runtime not found - this should not happen after eslint installation")
	}

	nodeBinary := nodeRuntime.Binaries["node"]
	if nodeBinary == "" {
		return fmt.Errorf("node binary not found in runtime configuration")
	}

	// Run ESLint
	eslintInstallationDirectory := eslint.InstallDir
	return tools.RunEslint(workDirectory, eslintInstallationDirectory, nodeBinary, pathsToCheck, autoFix, outputFile, outputFormat)
}

func runTrivyAnalysis(workDirectory string, pathsToCheck []string, outputFile string, outputFormat string) error {
	// Ensure Trivy tool is configured and installed
	trivy := config.Config.Tools()["trivy"]
	if trivy == nil || !config.Config.IsToolInstalled("trivy", trivy) {
		if trivy == nil {
			fmt.Println("Trivy tool configuration not found, adding and installing...")
		} else {
			fmt.Println("Trivy tool is not installed, installing...")
		}

		err := config.InstallTool("trivy", trivy, "")
		if err != nil {
			return fmt.Errorf("failed to install trivy: %w", err)
		}

		// Get the updated tool info after installation
		trivy = config.Config.Tools()["trivy"]
		if trivy == nil {
			return fmt.Errorf("trivy tool configuration still not found after installation")
		}
		fmt.Println("Trivy tool installed successfully")
	}

	// Ensure trivy binary is available
	trivyBinary := trivy.Binaries["trivy"]
	if trivyBinary == "" {
		return fmt.Errorf("trivy binary not found in tool configuration")
	}

	// Run Trivy
	return tools.RunTrivy(workDirectory, trivyBinary, pathsToCheck, outputFile, outputFormat)
}

func runPmdAnalysis(workDirectory string, pathsToCheck []string, outputFile string, outputFormat string) error {
	// Ensure PMD tool is configured and installed
	pmd := config.Config.Tools()["pmd"]
	if pmd == nil || !config.Config.IsToolInstalled("pmd", pmd) {
		if pmd == nil {
			fmt.Println("PMD tool configuration not found, adding and installing...")
		} else {
			fmt.Println("PMD tool is not installed, installing...")
		}

		err := config.InstallTool("pmd", pmd, "")
		if err != nil {
			return fmt.Errorf("failed to install pmd: %w", err)
		}

		// Get the updated tool info after installation
		pmd = config.Config.Tools()["pmd"]
		if pmd == nil {
			return fmt.Errorf("pmd tool configuration still not found after installation")
		}
		fmt.Println("PMD tool installed successfully")
	}

	// Ensure Java runtime is available
	javaRuntime := config.Config.Runtimes()["java"]
	if javaRuntime == nil {
		return fmt.Errorf("java runtime not found - this should not happen after pmd installation")
	}

	// Ensure pmd binary is available
	pmdBinary := pmd.Binaries["pmd"]
	if pmdBinary == "" {
		return fmt.Errorf("pmd binary not found in tool configuration")
	}

	// Run PMD
	return tools.RunPmd(workDirectory, pmdBinary, pathsToCheck, outputFile, outputFormat, config.Config)
}

func runPylintAnalysis(workDirectory string, pathsToCheck []string, outputFile string, outputFormat string) error {
	// Ensure Pylint tool is configured and installed
	pylint := config.Config.Tools()["pylint"]
	if pylint == nil || !config.Config.IsToolInstalled("pylint", pylint) {
		if pylint == nil {
			fmt.Println("Pylint tool configuration not found, adding and installing...")
		} else {
			fmt.Println("Pylint tool is not installed, installing...")
		}

		err := config.InstallTool("pylint", pylint, "")
		if err != nil {
			return fmt.Errorf("failed to install pylint: %w", err)
		}

		// Get the updated tool info after installation
		pylint = config.Config.Tools()["pylint"]
		if pylint == nil {
			return fmt.Errorf("pylint tool configuration still not found after installation")
		}
		fmt.Println("Pylint tool installed successfully")
	}

	// Ensure Python runtime is available
	pythonRuntime := config.Config.Runtimes()["python"]
	if pythonRuntime == nil {
		return fmt.Errorf("python runtime not found - this should not happen after pylint installation")
	}

	// Ensure python binary is available
	pylintBinary := pylint.Binaries["python"]
	if pylintBinary == "" {
		return fmt.Errorf("python binary not found in pylint tool configuration")
	}

	// Run Pylint
	return tools.RunPylint(workDirectory, pylintBinary, pathsToCheck, outputFile, outputFormat)
}

func runDartAnalyzer(workDirectory string, pathsToCheck []string, outputFile string, outputFormat string) error {
	// Ensure Dart Analyzer tool is configured and installed
	dartanalyzer := config.Config.Tools()["dartanalyzer"]
	if dartanalyzer == nil || !config.Config.IsToolInstalled("dartanalyzer", dartanalyzer) {
		if dartanalyzer == nil {
			fmt.Println("Dart analyzer tool configuration not found, adding and installing...")
		} else {
			fmt.Println("Dart analyzer tool is not installed, installing...")
		}

		err := config.InstallTool("dartanalyzer", dartanalyzer, "")
		if err != nil {
			return fmt.Errorf("failed to install dartanalyzer: %w", err)
		}

		// Get the updated tool info after installation
		dartanalyzer = config.Config.Tools()["dartanalyzer"]
		if dartanalyzer == nil {
			return fmt.Errorf("dartanalyzer tool configuration still not found after installation")
		}
		fmt.Println("Dart analyzer tool installed successfully")
	}

	// Ensure Dart runtime is available (dart or flutter)
	if config.Config.Runtimes()["flutter"] == nil && config.Config.Runtimes()["dart"] == nil {
		return fmt.Errorf("dart or flutter runtime not found - this should not happen after dartanalyzer installation")
	}

	// Ensure dart binary is available
	dartBinary := dartanalyzer.Binaries["dart"]
	if dartBinary == "" {
		return fmt.Errorf("dart binary not found in dartanalyzer tool configuration")
	}

	// Run Dart Analyzer
	return tools.RunDartAnalyzer(workDirectory, dartanalyzer.InstallDir, dartBinary, pathsToCheck, outputFile, outputFormat)
}

func runSemgrepAnalysis(workDirectory string, pathsToCheck []string, outputFile string, outputFormat string) error {
	// Ensure Semgrep tool is configured and installed
	semgrep := config.Config.Tools()["semgrep"]
	if semgrep == nil || !config.Config.IsToolInstalled("semgrep", semgrep) {
		if semgrep == nil {
			fmt.Println("Semgrep tool configuration not found, adding and installing...")
		} else {
			fmt.Println("Semgrep tool is not installed, installing...")
		}

		err := config.InstallTool("semgrep", semgrep, "")
		if err != nil {
			return fmt.Errorf("failed to install semgrep: %w", err)
		}

		// Get the updated tool info after installation
		semgrep = config.Config.Tools()["semgrep"]
		if semgrep == nil {
			return fmt.Errorf("semgrep tool configuration still not found after installation")
		}
		fmt.Println("Semgrep tool installed successfully")
	}

	// Ensure Python runtime is available
	pythonRuntime := config.Config.Runtimes()["python"]
	if pythonRuntime == nil {
		return fmt.Errorf("python runtime not found - this should not happen after semgrep installation")
	}

	// Ensure semgrep binary is available
	semgrepBinary := semgrep.Binaries["semgrep"]
	if semgrepBinary == "" {
		return fmt.Errorf("semgrep binary not found in tool configuration")
	}

	// Run Semgrep
	return tools.RunSemgrep(workDirectory, semgrepBinary, pathsToCheck, outputFile, outputFormat)
}

func runLizardAnalysis(workDirectory string, pathsToCheck []string, outputFile string, outputFormat string) error {
	// Ensure Lizard tool is configured and installed
	lizardTool := config.Config.Tools()["lizard"]
	if lizardTool == nil || !config.Config.IsToolInstalled("lizard", lizardTool) {
		if lizardTool == nil {
			fmt.Println("Lizard tool configuration not found, adding and installing...")
		} else {
			fmt.Println("Lizard tool is not installed, installing...")
		}

		err := config.InstallTool("lizard", lizardTool, "")
		if err != nil {
			return fmt.Errorf("failed to install lizard: %w", err)
		}

		// Get the updated tool info after installation
		lizardTool = config.Config.Tools()["lizard"]
		if lizardTool == nil {
			return fmt.Errorf("lizard tool configuration still not found after installation")
		}
		fmt.Println("Lizard tool installed successfully")
	}

	// Ensure Python runtime is available
	pythonRuntime := config.Config.Runtimes()["python"]
	if pythonRuntime == nil {
		return fmt.Errorf("python runtime not found - this should not happen after lizard installation")
	}

	// Ensure python binary is available
	lizardBinary := lizardTool.Binaries["python"]
	if lizardBinary == "" {
		return fmt.Errorf("python binary not found in lizard tool configuration")
	}

	// Get configuration patterns
	configFile, exists := tools.ConfigFileExists(config.Config, "lizard.yaml")
	var patterns []domain.PatternDefinition
	var err error

	if exists {
		// Configuration exists, read from file
		patterns, err = lizard.ReadConfig(configFile)
		if err != nil {
			return fmt.Errorf("error reading config file: %v", err)
		}
	} else {
		fmt.Println("No configuration file found for Lizard, using default patterns, run init with repository token to get a custom configuration")
		patterns, err = tools.FetchDefaultEnabledPatterns(domain.Lizard)
		if err != nil {
			return fmt.Errorf("failed to fetch default patterns: %v", err)
		}
	}

	// Run Lizard
	return lizard.RunLizard(workDirectory, lizardBinary, pathsToCheck, outputFile, outputFormat, patterns)
}

func runEnigmaAnalysis(workDirectory string, pathsToCheck []string, outputFile string, outputFormat string) error {
	// Ensure Enigma tool is configured and installed
	enigma := config.Config.Tools()["codacy-enigma-cli"]
	if enigma == nil || !config.Config.IsToolInstalled("codacy-enigma-cli", enigma) {
		if enigma == nil {
			fmt.Println("Enigma tool configuration not found, adding and installing...")
		} else {
			fmt.Println("Enigma tool is not installed, installing...")
		}

		err := config.InstallTool("codacy-enigma-cli", enigma, "")
		if err != nil {
			return fmt.Errorf("failed to install codacy-enigma-cli: %w", err)
		}

		// Get the updated tool info after installation
		enigma = config.Config.Tools()["codacy-enigma-cli"]
		if enigma == nil {
			return fmt.Errorf("codacy-enigma-cli tool configuration still not found after installation")
		}
		fmt.Println("Enigma tool installed successfully")
	}

	// Ensure enigma binary is available
	enigmaBinary := enigma.Binaries["codacy-enigma-cli"]
	if enigmaBinary == "" {
		return fmt.Errorf("codacy-enigma-cli binary not found in tool configuration")
	}

	// Run Enigma
	return tools.RunEnigma(workDirectory, enigma.InstallDir, enigmaBinary, pathsToCheck, outputFile, outputFormat)
}

func runReviveAnalysis(workDirectory string, pathsToCheck []string, outputFile string, outputFormat string) error {
	revive := config.Config.Tools()["revive"]
	if revive == nil {
		log.Fatal("Revive tool configuration not found")
	}
	reviveBinary := revive.Binaries["revive"]

	return reviveTool.RunRevive(workDirectory, reviveBinary, pathsToCheck, outputFile, outputFormat)
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
		var toolsToRun map[string]*plugins.ToolInfo

		if toolsToAnalyzeParam != "" {
			// If a specific tool is specified, only run that tool
			toolsToRun = map[string]*plugins.ToolInfo{
				toolsToAnalyzeParam: config.Config.Tools()[toolsToAnalyzeParam],
			}
		} else {
			// Run all configured tools
			toolsToRun = config.Config.Tools()
			log.Println("Running all configured tools...")
		}

		if len(toolsToRun) == 0 {
			log.Fatal("No tools configured. Please run 'codacy-cli init' and 'codacy-cli install' first")
		}

		// Filter tools by language support
		toolsToRun = FilterToolsByLanguageSupport(toolsToRun, args)

		if len(toolsToRun) == 0 {
			log.Println("No tools support the specified file(s). Skipping analysis.")
			return
		}

		log.Println("Running tools for the specified file(s)...")

		if outputFormat == "sarif" {
			// Create temporary directory for individual tool outputs
			tmpDir, err := os.MkdirTemp("", "codacy-analysis-*")
			if err != nil {
				log.Fatalf("Failed to create temporary directory: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			var sarifOutputs []string
			for toolName := range toolsToRun {
				log.Printf("Running %s...\n", toolName)
				tmpFile := filepath.Join(tmpDir, fmt.Sprintf("%s.sarif", toolName))
				if err := runTool(workDirectory, toolName, args, tmpFile); err != nil {
					log.Printf("Tool failed to run: %s: %v\n", toolName, err)
				}
				sarifOutputs = append(sarifOutputs, tmpFile)
			}

			// create output file tmp file
			tmpOutputFile := filepath.Join(tmpDir, "merged.sarif")

			// Merge all SARIF outputs
			if err := utils.MergeSarifOutputs(sarifOutputs, tmpOutputFile); err != nil {
				log.Fatalf("Failed to merge SARIF outputs: %v", err)
			}

			// Filter rules from the merged SARIF output
			sarifData, err := os.ReadFile(tmpOutputFile)
			if err != nil {
				log.Fatalf("Failed to read merged SARIF output: %v", err)
			}

			filteredData, err := utils.FilterRulesFromSarif(sarifData)
			if err != nil {
				log.Fatalf("Failed to filter rules from SARIF: %v", err)
			}

			if outputFile != "" {
				// Write filtered SARIF to output file
				os.WriteFile(outputFile, filteredData, utils.DefaultFilePerms)
			} else {
				// Print the filtered SARIF output
				fmt.Println(string(filteredData))
			}
		} else {
			// Run tools without merging outputs
			for toolName := range toolsToRun {
				log.Printf("Running %s...\n", toolName)
				if err := runTool(workDirectory, toolName, args, outputFile); err != nil {
					log.Printf("Tool failed to run: %s: %v\n", toolName, err)
				}
			}
		}
	},
}

func runTool(workDirectory string, toolName string, args []string, outputFile string) error {
	switch toolName {
	case "eslint":
		return runEslintAnalysis(workDirectory, args, autoFix, outputFile, outputFormat)
	case "trivy":
		return runTrivyAnalysis(workDirectory, args, outputFile, outputFormat)
	case "pmd":
		return runPmdAnalysis(workDirectory, args, outputFile, outputFormat)
	case "pylint":
		return runPylintAnalysis(workDirectory, args, outputFile, outputFormat)
	case "semgrep":
		return runSemgrepAnalysis(workDirectory, args, outputFile, outputFormat)
	case "dartanalyzer":
		return runDartAnalyzer(workDirectory, args, outputFile, outputFormat)
	case "lizard":
		return runLizardAnalysis(workDirectory, args, outputFile, outputFormat)
	case "codacy-enigma-cli":
		return runEnigmaAnalysis(workDirectory, args, outputFile, outputFormat)
	case "revive":
		return runReviveAnalysis(workDirectory, args, outputFile, outputFormat)
	default:
		return fmt.Errorf("unsupported tool: %s", toolName)
	}
}
