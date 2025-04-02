package cmd

import (
	cfg "codacy/cli-v2/config"
	config_file "codacy/cli-v2/config-file"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

var registry string

func init() {
	installCmd.Flags().StringVarP(&registry, "registry", "r", "", "Registry to use for installing tools")
	rootCmd.AddCommand(installCmd)
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Installs the tools specified in the project's config-file.",
	Long:  "Installs all runtimes and tools specified in the project's config-file file.",
	Run: func(cmd *cobra.Command, args []string) {
		bold := color.New(color.Bold)
		green := color.New(color.FgGreen)

		// Initialize config
		cfg.Init()

		// Load config file
		if err := config_file.ReadConfigFile(cfg.Config.ProjectConfigFile()); err != nil {
			log.Fatalf("Failed to load config file: %v", err)
		}

		// Check if anything needs to be installed
		needsInstallation := false
		for name, runtime := range cfg.Config.Runtimes() {
			if !cfg.Config.IsRuntimeInstalled(name, runtime) {
				needsInstallation = true
				break
			}
		}
		if !needsInstallation {
			for name, tool := range cfg.Config.Tools() {
				if !cfg.Config.IsToolInstalled(name, tool) {
					needsInstallation = true
					break
				}
			}
		}

		if !needsInstallation {
			fmt.Println()
			bold.Println("✅ All components are already installed!")
			return
		}

		fmt.Println()
		bold.Println("🚀 Starting installation process...")
		fmt.Println()

		// Calculate total items to install
		totalItems := 0
		for name, runtime := range cfg.Config.Runtimes() {
			if !cfg.Config.IsRuntimeInstalled(name, runtime) {
				totalItems++
			}
		}
		for name, tool := range cfg.Config.Tools() {
			if !cfg.Config.IsToolInstalled(name, tool) {
				totalItems++
			}
		}

		if totalItems == 0 {
			fmt.Println()
			bold.Println("✅ All components are already installed!")
			return
		}

		// Print list of items to install
		fmt.Println("📦 Items to install:")
		for name, runtime := range cfg.Config.Runtimes() {
			if !cfg.Config.IsRuntimeInstalled(name, runtime) {
				fmt.Printf("  • Runtime: %s v%s\n", name, runtime.Version)
			}
		}
		for name, tool := range cfg.Config.Tools() {
			if !cfg.Config.IsToolInstalled(name, tool) {
				fmt.Printf("  • Tool: %s v%s\n", name, tool.Version)
			}
		}
		fmt.Println()

		// Create a single progress bar for the entire installation
		progressBar := progressbar.NewOptions(totalItems,
			progressbar.OptionSetDescription("Installing components..."),
			progressbar.OptionSetTheme(progressbar.Theme{
				Saucer:        "█",
				SaucerHead:    "█",
				SaucerPadding: "░",
				BarStart:      "│",
				BarEnd:        "│",
			}),
			progressbar.OptionShowCount(),
			progressbar.OptionShowIts(),
			progressbar.OptionSetWidth(50),
			progressbar.OptionThrottle(100*time.Millisecond),
			progressbar.OptionSpinnerType(14),
			progressbar.OptionFullWidth(),
			progressbar.OptionSetRenderBlankState(true),
			progressbar.OptionOnCompletion(func() {
				fmt.Println()
			}),
		)

		// Redirect all output to /dev/null during installation
		oldStdout := os.Stdout
		devNull, _ := os.Open(os.DevNull)
		os.Stdout = devNull
		log.SetOutput(io.Discard)

		// Install runtimes
		for name, runtime := range cfg.Config.Runtimes() {
			if !cfg.Config.IsRuntimeInstalled(name, runtime) {
				progressBar.Describe(fmt.Sprintf("Installing runtime: %s v%s...", name, runtime.Version))
				err := cfg.InstallRuntime(name, runtime)
				if err != nil {
					log.Fatal(err)
				}
				progressBar.Add(1)
			}
		}

		// Install tools
		for name, tool := range cfg.Config.Tools() {
			if !cfg.Config.IsToolInstalled(name, tool) {
				progressBar.Describe(fmt.Sprintf("Installing tool: %s v%s...", name, tool.Version))
				err := cfg.InstallTool(name, tool)
				if err != nil {
					log.Fatal(err)
				}
				progressBar.Add(1)
			}
		}

		// Restore output
		os.Stdout = oldStdout
		devNull.Close()
		log.SetOutput(os.Stderr)

		// Print completion status
		fmt.Println()
		for name, runtime := range cfg.Config.Runtimes() {
			if !cfg.Config.IsRuntimeInstalled(name, runtime) {
				green.Printf("  ✓ Runtime: %s v%s\n", name, runtime.Version)
			}
		}
		for name, tool := range cfg.Config.Tools() {
			if !cfg.Config.IsToolInstalled(name, tool) {
				green.Printf("  ✓ Tool: %s v%s\n", name, tool.Version)
			}
		}
		fmt.Println()
		bold.Println("✅ Installation completed successfully!")
	},
}

func installRuntimes(config *cfg.ConfigType) {
	err := cfg.InstallRuntimes()
	if err != nil {
		log.Fatal(err)
	}
}

func installTools(config *cfg.ConfigType) {
	// Use the new tools-installer instead of manual installation
	err := cfg.InstallTools()
	if err != nil {
		log.Fatal(err)
	}
}
