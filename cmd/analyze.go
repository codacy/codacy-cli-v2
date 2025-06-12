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
	"log"
	"os"
	"path/filepath"
	"strings"

	"codacy/cli-v2/utils"

	codacyclient "codacy/cli-v2/codacy-client"

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

func init() {
	analyzeCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file for analysis results")
	analyzeCmd.Flags().StringVarP(&toolsToAnalyzeParam, "tool", "t", "", "Which tool to run analysis with. If not specified, all configured tools will be run")
	analyzeCmd.Flags().StringVar(&outputFormat, "format", "", "Output format (use 'sarif' for SARIF format)")
	analyzeCmd.Flags().BoolVar(&autoFix, "fix", false, "Apply auto fix to your issues when available")
	rootCmd.AddCommand(analyzeCmd)
}

func loadsToolAndPatterns(toolName string) (domain.Tool, []domain.PatternConfiguration) {
	var toolsResponse, err = codacyclient.GetToolsVersions()
	if err != nil {
		fmt.Println("Error:", err)
		return domain.Tool{}, []domain.PatternConfiguration{}
	}
	var tool domain.Tool
	for _, t := range toolsResponse {
		if t.Name == toolName {
			tool = t
			break
		}
	}
	var patterns []domain.PatternConfiguration
	patterns, err = codacyclient.GetDefaultToolPatternsConfig(domain.InitFlags{}, tool.Uuid)
	if err != nil {
		fmt.Println("Error:", err)
		return domain.Tool{}, []domain.PatternConfiguration{}
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

func validateToolName(toolName string) error {
	if toolName == "" {
		return fmt.Errorf("tool name cannot be empty")
	}

	// Get plugin manager to access the tools filesystem
	pluginManager := plugins.GetPluginManager()

	// Try to get the tool configuration - this will fail if the tool doesn't exist
	_, err := pluginManager.GetToolConfig(toolName)
	if err != nil {
		return fmt.Errorf("tool '%s' is not supported", toolName)
	}

	return nil
}

func runToolByName(toolName string, workDirectory string, pathsToCheck []string, autoFix bool, outputFile string, outputFormat string, tool *plugins.ToolInfo, runtime *plugins.RuntimeInfo) error {
	switch toolName {
	case "eslint":
		binaryPath := runtime.Binaries[tool.Runtime]
		return tools.RunEslint(workDirectory, tool.InstallDir, binaryPath, pathsToCheck, autoFix, outputFile, outputFormat)
	case "trivy":
		binaryPath := tool.Binaries[toolName]
		return tools.RunTrivy(workDirectory, binaryPath, pathsToCheck, outputFile, outputFormat)
	case "pmd":
		binaryPath := tool.Binaries[toolName]
		return tools.RunPmd(workDirectory, binaryPath, pathsToCheck, outputFile, outputFormat, config.Config)
	case "pylint":
		binaryPath := tool.Binaries[tool.Runtime]
		return tools.RunPylint(workDirectory, binaryPath, pathsToCheck, outputFile, outputFormat)
	case "dartanalyzer":
		binaryPath := tool.Binaries[tool.Runtime]
		return tools.RunDartAnalyzer(workDirectory, tool.InstallDir, binaryPath, pathsToCheck, outputFile, outputFormat)
	case "semgrep":
		binaryPath := tool.Binaries[toolName]
		return tools.RunSemgrep(workDirectory, binaryPath, pathsToCheck, outputFile, outputFormat)
	case "lizard":
		binaryPath := tool.Binaries[tool.Runtime]
		return lizard.RunLizard(workDirectory, binaryPath, pathsToCheck, outputFile, outputFormat)
	case "codacy-enigma-cli":
		return tools.RunEnigma(workDirectory, tool.InstallDir, tool.Binaries["codacy-enigma-cli"], pathsToCheck, outputFile, outputFormat)
	case "revive":
		return reviveTool.RunRevive(workDirectory, tool.Binaries["revive"], pathsToCheck, outputFile, outputFormat)
	}
	return fmt.Errorf("unsupported tool: %s", toolName)
}

func runTool(workDirectory string, toolName string, pathsToCheck []string, outputFile string, autoFix bool, outputFormat string) error {
	err := validateToolName(toolName)
	if err != nil {
		return err
	}
	log.Println("Running tools for the specified file(s)...")
	log.Printf("Running %s...\n", toolName)

	tool := config.Config.Tools()[toolName]
	var isToolInstalled bool
	if tool == nil {
		isToolInstalled = false
	} else {
		isToolInstalled = config.Config.IsToolInstalled(toolName, tool)
	}
	var isRuntimeInstalled bool

	var runtime *plugins.RuntimeInfo

	if toolName == "codacy-enigma-cli" {
		isToolInstalled = true
	}

	if tool == nil || !isToolInstalled {
		if tool == nil {
			fmt.Println("Tool configuration not found, adding and installing...")
		}
		if !isToolInstalled {
			fmt.Println("Tool is not installed, installing...")
		}
		err := config.InstallTool(toolName, tool, "")
		if err != nil {
			return fmt.Errorf("failed to install %s: %w", toolName, err)
		}
		tool = config.Config.Tools()[toolName]
		runtime = config.Config.Runtimes()[tool.Runtime]
		isRuntimeInstalled = runtime == nil || config.Config.IsRuntimeInstalled(tool.Runtime, runtime)
		if !isRuntimeInstalled {
			fmt.Printf("%s runtime is not installed, installing...\n", tool.Runtime)
			err := config.InstallRuntime(tool.Runtime, runtime)
			if err != nil {
				return fmt.Errorf("failed to install %s runtime: %w", tool.Runtime, err)
			}
			runtime = config.Config.Runtimes()[tool.Runtime]
		}

	} else {
		runtime = config.Config.Runtimes()[tool.Runtime]
		isRuntimeInstalled = runtime == nil || config.Config.IsRuntimeInstalled(tool.Runtime, runtime)
		if !isRuntimeInstalled {
			fmt.Printf("%s runtime is not installed, installing...\n", tool.Runtime)
			err := config.InstallRuntime(tool.Runtime, runtime)
			if err != nil {
				return fmt.Errorf("failed to install %s runtime: %w", tool.Runtime, err)
			}
			runtime = config.Config.Runtimes()[tool.Runtime]
		}
	}

	if tool.ConfigFileName != "" {
		configFile, exists := HasConfigFile(config.Config, tool.ConfigFileName)
		if !exists {
			fmt.Printf("Config file %s not found, creating...\n", tool.ConfigFileName)
			// TODO: Implement CreateConfigFile function in utils
			// err := utils.CreateConfigFile(config.Config, tool.ConfigFileName)
			// if err != nil {
			//	return fmt.Errorf("failed to create config file %s: %w", tool.ConfigFileName, err)
			// }
		} else {
			fmt.Printf("Using existing config file: %s\n", configFile)
		}
	}

	return runToolByName(toolName, workDirectory, pathsToCheck, autoFix, outputFile, outputFormat, tool, runtime)
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

		if outputFormat == "sarif" {
			// Create temporary directory for individual tool outputs
			tmpDir, err := os.MkdirTemp("", "codacy-analysis-*")
			if err != nil {
				log.Fatalf("Failed to create temporary directory: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			var sarifOutputs []string
			for toolName := range toolsToRun {
				tmpFile := filepath.Join(tmpDir, fmt.Sprintf("%s.sarif", toolName))
				if err := runTool(workDirectory, toolName, args, tmpFile, autoFix, outputFormat); err != nil {
					log.Printf("Tool failed to run: %v\n", err)
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
				if err := runTool(workDirectory, toolName, args, outputFile, autoFix, outputFormat); err != nil {
					log.Printf("Tool failed to run: %v\n", err)
				}
			}
		}
	},
}

// HasConfigFile checks if a tool's config file exists in the .codacy/tools-configs directory
// or in the repository root directory.
//
// Parameters:
//   - config: The configuration object containing directory paths
//   - configFileName: The name of the config file to check for
//
// Returns:
//   - string: The full path to the config file if it exists
//   - bool: True if the config file exists, false otherwise
func HasConfigFile(config interface{}, configFileName string) (string, bool) {
	fmt.Println("Checking for config file:", configFileName)
	// If no config file name is specified, return false
	if configFileName == "" {
		return "", false
	}

	// We need to get the workspace directory from the config
	// For now, we'll check in the current directory and .codacy/tools-config
	workDir, err := os.Getwd()
	if err != nil {
		return "", false
	}

	fmt.Println("Work directory:", workDir)
	fmt.Println("Checking in .codacy/tools-configs directory first")
	// Check in .codacy/tools-configs directory first
	toolsConfigDir := filepath.Join(workDir, ".codacy", "tools-configs")
	fmt.Println("Tools config directory:", toolsConfigDir)
	configPath := filepath.Join(toolsConfigDir, configFileName)
	fmt.Println("Config path:", configPath)
	if _, err := os.Stat(configPath); err == nil {
		fmt.Println("Found config file in tools-configs directory!")
		return configPath, true
	} else {
		fmt.Println("Config file NOT found in tools-configs directory. Error:", err)
	}

	// Check in repository root directory as fallback
	rootConfigPath := filepath.Join(workDir, configFileName)
	fmt.Println("Checking root config path:", rootConfigPath)
	if _, err := os.Stat(rootConfigPath); err == nil {
		fmt.Println("Found config file in repository root directory!")
		return rootConfigPath, true
	} else {
		fmt.Println("Config file NOT found in repository root. Error:", err)
	}

	// Config file not found
	fmt.Println("Config file not found anywhere")
	return "", false
}
