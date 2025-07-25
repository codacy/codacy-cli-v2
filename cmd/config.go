package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"codacy/cli-v2/cmd/cmdutils"
	"codacy/cli-v2/cmd/configsetup"
	codacyclient "codacy/cli-v2/codacy-client"
	"codacy/cli-v2/config"
	"codacy/cli-v2/constants"
	"codacy/cli-v2/domain"
	"codacy/cli-v2/plugins"
	"codacy/cli-v2/tools"
	"codacy/cli-v2/utils/logger"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// configResetInitFlags holds the flags for the config reset command.
var configResetInitFlags domain.InitFlags

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage Codacy configuration",
}

var configResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset Codacy configuration to default or repository-specific settings",
	Long: "Resets the Codacy configuration files and tool-specific configurations. " +
		"This command will overwrite an existing configuration with local default configurations " +
		"if no API token is provided (and current mode is not 'remote'). If an API token is provided, it will fetch and apply " +
		"repository-specific configurations from the Codacy API, effectively resetting to those.",
	Run: func(cmd *cobra.Command, args []string) {
		// Get current CLI mode from config
		currentCliMode, err := config.Config.GetCliMode()
		if err != nil {
			// Log the error for debugging purposes
			log.Printf("Warning: Could not determine CLI mode from cli-config.yaml: %v. Defaulting to 'local' mode.", err)
			// Show a user-facing warning on stdout
			fmt.Println("⚠️  Warning: Could not read or parse .codacy/cli-config.yaml. Defaulting to 'local' CLI mode.")
			fmt.Println("   You might want to run 'codacy-cli init' or 'codacy-cli config reset --api-token ...' to correctly set up your configuration.")
			fmt.Println()
			currentCliMode = "local" // Default to local as per existing logic
		}

		apiTokenFlagProvided := len(configResetInitFlags.ApiToken) > 0

		// If current mode is 'remote', prevent resetting to local without explicit API token for a remote reset.
		if currentCliMode == "remote" && !apiTokenFlagProvided {
			fmt.Println("Error: Your Codacy CLI is currently configured in 'remote' (cloud) mode.")
			fmt.Println("To reset your configuration using remote settings, you must provide the --api-token, --provider, --organization, and --repository flags.")
			fmt.Println("Running 'config reset' without these flags is not permitted while configured for 'remote' mode.")
			fmt.Println("This prevents an accidental switch to a local default configuration.")
			fmt.Println()
			if errHelp := cmd.Help(); errHelp != nil {
				log.Printf("Warning: Failed to display command help: %v\n", errHelp)
			}
			os.Exit(1)
		}

		// Validate flags: if API token is provided, other related flags must also be provided.
		if apiTokenFlagProvided {
			if configResetInitFlags.Provider == "" || configResetInitFlags.Organization == "" || configResetInitFlags.Repository == "" {
				fmt.Println("Error: When using --api-token, you must also provide --provider, --organization, and --repository flags.")
				fmt.Println("Please provide all required flags and try again.")
				fmt.Println()
				if errHelp := cmd.Help(); errHelp != nil {
					log.Fatalf("Failed to display command help: %v", errHelp)
				}
				os.Exit(1)
			}
		}

		codacyConfigFile := config.Config.ProjectConfigFile()
		// Check if the main configuration file exists
		if _, err := os.Stat(codacyConfigFile); os.IsNotExist(err) {
			fmt.Println("Configuration file (.codacy/codacy.yaml) not found, running initialization logic...")
			runConfigResetLogic(cmd, args, configResetInitFlags)
		} else {
			fmt.Println("Resetting existing Codacy configuration...")
			runConfigResetLogic(cmd, args, configResetInitFlags)
		}
	},
}

// runConfigResetLogic contains the core logic for resetting or initializing the configuration.
// It mirrors the behavior of the original init command but uses shared functions from the configsetup package.
func runConfigResetLogic(cmd *cobra.Command, args []string, flags domain.InitFlags) {
	// Create local .codacy directory first
	if err := config.Config.CreateLocalCodacyDir(); err != nil {
		log.Fatalf("Failed to create local codacy directory: %v", err)
	}

	// Create .codacy/tools-configs directory
	toolsConfigDir := config.Config.ToolsConfigDirectory()
	if err := os.MkdirAll(toolsConfigDir, constants.DefaultDirPerms); err != nil {
		log.Fatalf("Failed to create tools-configs directory: %v", err)
	}

	// Determine if running in local mode (no API token)
	cliLocalMode := len(flags.ApiToken) == 0

	if cliLocalMode {
		fmt.Println()
		fmt.Println("ℹ️  Resetting to local default configurations.")
		noTools := []domain.Tool{} // Empty slice for tools as we are in local mode without specific toolset from API initially
		if err := configsetup.CreateConfigurationFiles(noTools, cliLocalMode); err != nil {
			log.Fatalf("Failed to create base configuration files: %v", err)
		}
		// Create default configuration files for tools
		if err := configsetup.BuildDefaultConfigurationFiles(toolsConfigDir, flags); err != nil {
			log.Fatalf("Failed to build default tool configuration files: %v", err)
		}
		// Create the languages configuration file for local mode
		if err := configsetup.CreateLanguagesConfigFileLocal(toolsConfigDir); err != nil {
			log.Fatalf("Failed to create local languages configuration file: %v", err)
		}
	} else {
		// API token provided, fetch configuration from Codacy
		fmt.Println("API token specified. Fetching and applying repository-specific configurations from Codacy...")
		if err := configsetup.BuildRepositoryConfigurationFiles(flags); err != nil {
			log.Fatalf("Failed to build repository-specific configuration files: %v", err)
		}
	}

	// Create or update .gitignore file in .codacy directory
	if err := configsetup.CreateGitIgnoreFile(); err != nil {
		log.Printf("Warning: Failed to create or update .codacy/.gitignore: %v", err) // Log as warning, not fatal
	}

	fmt.Println()
	fmt.Println("✅ Successfully reset Codacy configuration!")
	fmt.Println()
	fmt.Println("🔧 Next steps:")
	fmt.Println("  1. Run 'codacy-cli install' to install all dependencies based on the new/updated configuration.")
	fmt.Println("  2. Run 'codacy-cli analyze' to start analyzing your code.")
	fmt.Println()
}

// Placeholder for the path argument for discover command
var discoverPath string

var configDiscoverCmd = &cobra.Command{
	Use:   "discover <path>",
	Short: "Discover project languages and update tool configurations",
	Long: "Scans the specified path to detect programming languages. " +
		"Updates .codacy/tools-configs/languages-config.yaml with discovered languages " +
		"and enables relevant tools in codacy.yaml. " +
		"This command does not change existing tool versions or run installations. " +
		"In Cloud mode, tools are only added if enabled in the cloud for the repository.",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		discoverPath = args[0]

		// Check if path exists
		if _, err := os.Stat(discoverPath); os.IsNotExist(err) {
			log.Fatalf("Error: Path %s does not exist", discoverPath)
		}

		fmt.Printf("Discovering languages and tools for path: %s\n", discoverPath)

		// Detect file extensions first
		extCount, err := config.DetectFileExtensions(discoverPath)
		if err != nil {
			log.Fatalf("Error detecting file extensions: %v", err)
		}

		defaultToolLangMap := tools.GetDefaultToolLanguageMapping()

		if len(extCount) > 0 {
			recognizedExts := config.GetRecognizableExtensions(extCount, defaultToolLangMap)
			if len(recognizedExts) > 0 {
				logger.Debug("Detected recognizable file extensions", logrus.Fields{
					"extensions": recognizedExts,
					"path":       discoverPath,
				})
			}
		}

		detectedTools, err := config.DetectRelevantTools(discoverPath, defaultToolLangMap)
		if err != nil {
			log.Fatalf("Error detecting relevant tools: %v", err)
		}
		if len(detectedTools) == 0 {
			fmt.Println("No relevant tools found for the file extensions in the provided path.")
			return
		}

		toolsConfigDir := config.Config.ToolsConfigsDirectory()
		if err := updateLanguagesConfigForTools(detectedTools, toolsConfigDir, defaultToolLangMap); err != nil {
			log.Fatalf("Error updating .codacy/tools-configs/languages-config.yaml: %v", err)
		}
		fmt.Println("Updated .codacy/tools-configs/languages-config.yaml")

		// For updating codacy.yaml, we need to know the current CLI mode and potentially API creds
		currentCliMode, err := config.Config.GetCliMode()
		if err != nil {
			log.Printf("Warning: Could not determine CLI mode: %v. Assuming local mode for tool enablement.", err)
			currentCliMode = "local" // Default to local
		}

		codacyYAMLPath := config.Config.ProjectConfigFile()

		if err := updateCodacyYAMLForTools(detectedTools, codacyYAMLPath, configResetInitFlags, currentCliMode); err != nil {
			if strings.Contains(err.Error(), "❌ Fatal:") {
				fmt.Println(err)
				os.Exit(1)
			}
			log.Fatalf("Error updating %s: %v", codacyYAMLPath, err)
		}
		fmt.Printf("Updated %s with relevant tools.\n", filepath.Base(codacyYAMLPath))

		// Create tool configuration files for discovered tools
		if len(detectedTools) > 0 {
			fmt.Printf("\nCreating tool configurations for discovered tools...\n")
			if err := configsetup.CreateConfigurationFilesForDiscoveredTools(detectedTools, toolsConfigDir, configResetInitFlags); err != nil {
				log.Printf("Warning: Failed to create some tool configurations: %v", err)
			}
		}

		fmt.Println("\n✅ Successfully discovered languages and updated configurations.")
		fmt.Println("   Please review the changes in '.codacy/codacy.yaml' and '.codacy/tools-configs/' directory.")
	},
}

// updateLanguagesConfigForTools updates the .codacy/tools-configs/languages-config.yaml file based on detected tools.
func updateLanguagesConfigForTools(detectedTools map[string]struct{}, toolsConfigDir string, defaultToolLangMap map[string]domain.ToolLanguageInfo) error {
	langConfigPath := filepath.Join(toolsConfigDir, "languages-config.yaml")

	// Read existing languages config if it exists
	var langConf domain.LanguagesConfig
	if content, err := os.ReadFile(langConfigPath); err == nil {
		if err := yaml.Unmarshal(content, &langConf); err != nil {
			return fmt.Errorf("failed to parse existing languages-config.yaml: %w", err)
		}
	}

	// Create a map of existing tools for quick lookup
	existingTools := make(map[string]domain.ToolLanguageInfo)
	for _, tool := range langConf.Tools {
		existingTools[tool.Name] = tool
	}

	// Add detected tools that are not already present
	for toolName := range detectedTools {
		if _, alreadyExists := existingTools[toolName]; !alreadyExists {
			if toolInfo, exists := defaultToolLangMap[toolName]; exists {
				langConf.Tools = append(langConf.Tools, toolInfo)
			}
		}
	}

	// Sort tools by name for consistent output
	sort.SliceStable(langConf.Tools, func(i, j int) bool {
		return langConf.Tools[i].Name < langConf.Tools[j].Name
	})

	data, err := yaml.Marshal(langConf)
	if err != nil {
		return fmt.Errorf("failed to marshal languages-config.yaml: %w", err)
	}
	if err := os.MkdirAll(toolsConfigDir, constants.DefaultDirPerms); err != nil {
		return fmt.Errorf("failed to create tools-configs directory: %w", err)
	}
	return os.WriteFile(langConfigPath, data, constants.DefaultFilePerms)
}

// updateCodacyYAMLForTools updates the codacy.yaml file with detected tools.
func updateCodacyYAMLForTools(detectedTools map[string]struct{}, codacyYAMLPath string, initFlags domain.InitFlags, cliMode string) error {

	var configData map[string]interface{}

	// Read and parse codacy.yaml (validation is done globally in PersistentPreRun)
	content, err := os.ReadFile(codacyYAMLPath)
	if err != nil {
		return fmt.Errorf("error reading %s: %w", codacyYAMLPath, err)
	}

	if err := yaml.Unmarshal(content, &configData); err != nil {
		return fmt.Errorf("error parsing %s: %w", codacyYAMLPath, err)
	}

	toolsRaw, _ := configData["tools"].([]interface{})
	currentToolsList := []string{}
	currentToolSetByName := make(map[string]string)        // "eslint" -> "eslint@version"
	currentToolSetWithVersion := make(map[string]struct{}) // "eslint@version" -> {}

	for _, t := range toolsRaw {
		if toolStr, ok := t.(string); ok {
			currentToolsList = append(currentToolsList, toolStr)
			currentToolSetWithVersion[toolStr] = struct{}{}
			parts := strings.Split(toolStr, "@")
			if len(parts) > 0 {
				currentToolSetByName[parts[0]] = toolStr
			}
		}
	}

	candidateToolsToAdd := detectedTools

	if cliMode == "remote" && initFlags.ApiToken != "" {
		fmt.Println("Cloud mode: Verifying tools against repository settings...")
		cloudTools, err := codacyclient.GetRepositoryTools(initFlags)
		if err != nil {
			return fmt.Errorf("failed to get repository tools from cloud: %w", err)
		}
		cloudEnabledToolNames := make(map[string]bool)
		for _, ct := range cloudTools {
			var toolShortName string
			for uuid, meta := range domain.SupportedToolsMetadata {
				if uuid == ct.Uuid {
					toolShortName = meta.Name
					break
				}
			}

			if toolShortName != "" && ct.Settings.Enabled {
				cloudEnabledToolNames[toolShortName] = true
			}
		}

		filteredCandidates := make(map[string]struct{})
		for toolName := range candidateToolsToAdd {
			if _, isEnabledInCloud := cloudEnabledToolNames[toolName]; isEnabledInCloud {
				filteredCandidates[toolName] = struct{}{}
			} else {
				fmt.Printf("Tool %s detected locally but not enabled in cloud, skipping addition to codacy.yaml.\n", toolName)
			}
		}
		candidateToolsToAdd = filteredCandidates
	}

	defaultToolVersions := plugins.GetToolVersions()

	// Discover mode: preserve existing tools and add only new discovered ones
	finalToolsList := currentToolsList // Start with existing tools
	addedNewTool := false

	// Add only newly discovered tools that aren't already configured
	for toolNameToAdd := range candidateToolsToAdd {
		if _, alreadyConfigured := currentToolSetByName[toolNameToAdd]; !alreadyConfigured {
			version, ok := defaultToolVersions[toolNameToAdd]
			if !ok {
				log.Printf("Warning: No default version found for tool %s. Skipping.", toolNameToAdd)
				continue
			}
			newToolEntry := toolNameToAdd + "@" + version
			if _, entryExists := currentToolSetWithVersion[newToolEntry]; !entryExists {
				finalToolsList = append(finalToolsList, newToolEntry)
				fmt.Printf("Adding tool to codacy.yaml: %s\n", newToolEntry)
				addedNewTool = true
			}
		}
	}

	// Sort the final list for consistency
	sort.Strings(finalToolsList)
	configData["tools"] = finalToolsList

	// Update runtimes if new tools were added
	if addedNewTool || len(currentToolsList) == 0 { // Also run if it's a new codacy.yaml
		neededRuntimes := make(map[string]string) // runtimeName -> runtimeVersion
		runtimeDependencies := plugins.GetToolRuntimeDependencies()
		defaultRuntimeVersions := plugins.GetRuntimeVersions()

		for _, toolEntry := range finalToolsList {
			toolName := strings.Split(toolEntry, "@")[0]
			if runtimeName, depends := runtimeDependencies[toolName]; depends {
				if _, alreadyNeeded := neededRuntimes[runtimeName]; !alreadyNeeded {
					if version, ok := defaultRuntimeVersions[runtimeName]; ok {
						neededRuntimes[runtimeName] = version
					} else {
						log.Printf("Warning: No default version for runtime %s needed by %s", runtimeName, toolName)
					}
				}
			}
		}
		// Add dart for dartanalyzer if not already covered by another dart-needing tool
		hasDartAnalyzer := false
		for _, toolEntry := range finalToolsList {
			if strings.HasPrefix(toolEntry, "dartanalyzer@") {
				hasDartAnalyzer = true
				break
			}
		}
		if hasDartAnalyzer {
			if _, dartNeeded := neededRuntimes["dart"]; !dartNeeded {
				if _, flutterNeeded := neededRuntimes["flutter"]; !flutterNeeded { // Only add dart if flutter isn't already there
					if dartVersion, ok := defaultRuntimeVersions["dart"]; ok {
						neededRuntimes["dart"] = dartVersion
					}
				}
			}
		}

		// Preserve existing runtimes and their versions if possible, only add new ones.
		existingRuntimesRaw, _ := configData["runtimes"].([]interface{})
		finalRuntimesList := []string{}
		existingRuntimeSet := make(map[string]string) // name -> name@version

		for _, r := range existingRuntimesRaw {
			if rtStr, ok := r.(string); ok {
				finalRuntimesList = append(finalRuntimesList, rtStr)
				name := strings.Split(rtStr, "@")[0]
				existingRuntimeSet[name] = rtStr
			}
		}

		for rtName, rtVersion := range neededRuntimes {
			if _, exists := existingRuntimeSet[rtName]; !exists {
				finalRuntimesList = append(finalRuntimesList, rtName+"@"+rtVersion)
				fmt.Printf("Adding runtime to codacy.yaml: %s@%s\n", rtName, rtVersion)
			}
			// We are not updating versions of existing runtimes here, just adding new ones.
		}
		sort.Strings(finalRuntimesList)
		configData["runtimes"] = finalRuntimesList
	}

	yamlData, err := yaml.Marshal(configData)
	if err != nil {
		return fmt.Errorf("error marshaling %s: %w", codacyYAMLPath, err)
	}
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(codacyYAMLPath), constants.DefaultDirPerms); err != nil {
		return fmt.Errorf("error creating .codacy directory: %w", err)
	}

	return os.WriteFile(codacyYAMLPath, yamlData, constants.DefaultFilePerms)
}

func init() {
	// Add cloud-related flags to both commands
	cmdutils.AddCloudFlags(configResetCmd, &configResetInitFlags)
	cmdutils.AddCloudFlags(configDiscoverCmd, &configResetInitFlags)

	// Add subcommands to config command
	configCmd.AddCommand(configResetCmd, configDiscoverCmd)

	// Add config command to root
	rootCmd.AddCommand(configCmd)
}
