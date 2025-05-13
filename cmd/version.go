package cmd

import (
	"codacy/cli-v2/version"
	"fmt"
	"runtime"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display version information",
	Run: func(cmd *cobra.Command, args []string) {
		bold := color.New(color.Bold)
		cyan := color.New(color.FgCyan)

		fmt.Println()
		bold.Println("Codacy CLI Version Information")
		fmt.Println("-----------------------------")
		cyan.Printf("Version:    ")
		fmt.Println(version.Version)
		cyan.Printf("Commit:     ")
		fmt.Println(version.GitCommit)
		cyan.Printf("Built:      ")
		fmt.Println(version.BuildTime)
		cyan.Printf("Go version: ")
		fmt.Println(runtime.Version())
		cyan.Printf("OS/Arch:    ")
		fmt.Printf("%s/%s\n", runtime.GOOS, runtime.GOARCH)
		fmt.Println()
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
