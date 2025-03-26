package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(pingCmd)
}

var pingCmd = &cobra.Command{
	Use:   "ping",
	Short: "Replies with 'pong'.",
	Long:  "Test command. Replies with 'pong'.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("pong")
	},
}
