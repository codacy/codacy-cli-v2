package cmd

import (
	"codacy/cli-v2/config"
	config_file "codacy/cli-v2/config-file"
	"codacy/cli-v2/utils/logger"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/schollz/progressbar/v3"
	"github.com/sirupsen/logrus"
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

		// Create necessary directories
		if err := config.Config.CreateCodacyDirs(); err != nil {
			logger.Error("Failed to create Codacy directories", logrus.Fields{
				"error": err.Error(),
			})
			log.Fatal(err)
		}

		// Load config file
		if err := config_file.ReadConfigFile(config.Config.ProjectConfigFile()); err != nil {
			logger.Warn("Configuration file not found", logrus.Fields{
				"error": err.Error(),
			})
			fmt.Println()
			color.Red("⚠️  Warning: Could not find configuration file!")
			fmt.Println("Please run 'codacy-cli init' first to create a configuration file.")
			fmt.Println()
			os.Exit(1)
		}

		// Check if anything needs to be installed
		needsInstallation := false
		for name, runtime := range config.Config.Runtimes() {
			if !config.Config.IsRuntimeInstalled(name, runtime) {
				needsInstallation = true
				break
			}
		}
		if !needsInstallation {
			for name, tool := range config.Config.Tools() {
				if !config.Config.IsToolInstalled(name, tool) {
					needsInstallation = true
					break
				}
			}
		}

		if !needsInstallation {
			logger.Info("All components are already installed", nil)
			fmt.Println()
			bold.Println("✅ All components are already installed!")
			return
		}

		logger.Info("Starting installation process", nil)
		fmt.Println()
		bold.Println("🚀 Starting installation process...")
		fmt.Println()

		// Calculate total items to install
		totalItems := 0
		for name, runtime := range config.Config.Runtimes() {
			if !config.Config.IsRuntimeInstalled(name, runtime) {
				totalItems++
			}
		}
		for name, tool := range config.Config.Tools() {
			if !config.Config.IsToolInstalled(name, tool) {
				totalItems++
			}
		}

		if totalItems == 0 {
			logger.Info("All components are already installed", nil)
			fmt.Println()
			bold.Println("✅ All components are already installed!")
			return
		}

		// Print list of items to install
		fmt.Println("📦 Items to install:")
		for name, runtime := range config.Config.Runtimes() {
			if !config.Config.IsRuntimeInstalled(name, runtime) {
				logger.Info("Runtime scheduled for installation", logrus.Fields{
					"runtime": name,
					"version": runtime.Version,
				})
				fmt.Printf("  • Runtime: %s v%s\n", name, runtime.Version)
			}
		}
		for name, tool := range config.Config.Tools() {
			if !config.Config.IsToolInstalled(name, tool) {
				logger.Info("Tool scheduled for installation", logrus.Fields{
					"tool":    name,
					"version": tool.Version,
				})
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

		// Install runtimes first
		for name, runtime := range config.Config.Runtimes() {
			if !config.Config.IsRuntimeInstalled(name, runtime) {
				progressBar.Describe(fmt.Sprintf("Installing runtime: %s v%s...", name, runtime.Version))
				logger.Info("Installing runtime", logrus.Fields{
					"runtime": name,
					"version": runtime.Version,
				})
				err := config.InstallRuntime(name, runtime)
				if err != nil {
					logger.Error("Failed to install runtime", logrus.Fields{
						"runtime": name,
						"version": runtime.Version,
						"error":   err.Error(),
					})
					fmt.Printf("\n⚠️  Warning: Failed to install runtime %s v%s: %v\n", name, runtime.Version, err)
					// Continue with next runtime instead of fatal
					progressBar.Add(1)
					continue
				}
				logger.Info("Successfully installed runtime", logrus.Fields{
					"runtime": name,
					"version": runtime.Version,
				})
				progressBar.Add(1)
			}
		}

		// Install tools
		for name, tool := range config.Config.Tools() {
			if !config.Config.IsToolInstalled(name, tool) {
				progressBar.Describe(fmt.Sprintf("Installing tool: %s v%s...", name, tool.Version))
				logger.Info("Installing tool", logrus.Fields{
					"tool":    name,
					"version": tool.Version,
				})
				err := config.InstallTool(name, tool, registry)
				if err != nil {
					logger.Error("Failed to install tool", logrus.Fields{
						"tool":    name,
						"version": tool.Version,
						"error":   err.Error(),
					})
					fmt.Printf("\n⚠️  Warning: Failed to install tool %s v%s: %v\n", name, tool.Version, err)
					// Continue with next tool instead of fatal
					progressBar.Add(1)
					continue
				}
				logger.Info("Successfully installed tool", logrus.Fields{
					"tool":    name,
					"version": tool.Version,
				})
				progressBar.Add(1)
			}
		}

		// Restore output
		os.Stdout = oldStdout
		devNull.Close()
		log.SetOutput(os.Stderr)

		// Print completion status with warnings for failed installations
		fmt.Println()
		var hasFailures bool
		for name, runtime := range config.Config.Runtimes() {
			if !config.Config.IsRuntimeInstalled(name, runtime) {
				color.Yellow("  ⚠️  Runtime: %s v%s (installation failed)", name, runtime.Version)
				hasFailures = true
			} else {
				green.Printf("  ✓ Runtime: %s v%s\n", name, runtime.Version)
			}
		}
		for name, tool := range config.Config.Tools() {
			if !config.Config.IsToolInstalled(name, tool) {
				color.Yellow("  ⚠️  Tool: %s v%s (installation failed)", name, tool.Version)
				hasFailures = true
			} else {
				green.Printf("  ✓ Tool: %s v%s\n", name, tool.Version)
			}
		}
		fmt.Println()
		if hasFailures {
			logger.Warn("Installation completed with some failures", nil)
			bold.Println("⚠️  Installation completed with some failures!")
			fmt.Println("Some components failed to install. You can try installing them again with:")
			fmt.Println("  codacy-cli install")
		} else {
			logger.Info("Installation completed successfully", nil)
			bold.Println("✅ Installation completed successfully!")
		}
	},
}
