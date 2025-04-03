package cmd

import (
	"github.com/spf13/cobra"
	"go-figure/internal"
)

var rootCmd = &cobra.Command{
	Use:   "zavrsni",
	Short: "System backup and restore tool",
	Long:  "A tool for detecting installed packages, backing up configurations, and restoring them.",
}

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Go figure, it's a TUI",
	Run: func(cmd *cobra.Command, args []string) {
	},
}

var systemReportCmd = &cobra.Command{
	Use:   "sys-info",
	Short: "testing report",
	Run: func(cmd *cobra.Command, args []string) {
		internal.FetchSystemInfo()
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(tuiCmd)
	rootCmd.AddCommand(systemReportCmd)
}
