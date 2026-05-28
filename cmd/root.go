package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "sheetpilot",
	Short: "Excel MCP Server - AI-powered Excel operations",
	Long:  "SheetPilot is an MCP server that enables AI to operate Excel files like a human user.",
}

func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(serveCmd)
}

func Execute() error {
	return rootCmd.Execute()
}
