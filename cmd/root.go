package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "taso",
	Short: "taso — find what's silently wrong with your environment",
	Long: `
  ████████╗ █████╗ ███████╗ ██████╗
     ██╔══╝██╔══██╗██╔════╝██╔═══██╗
     ██║   ███████║███████╗██║   ██║
     ██║   ██╔══██║╚════██║██║   ██║
     ██║   ██║  ██║███████║╚██████╔╝
     ╚═╝   ╚═╝  ╚═╝╚══════╝ ╚═════╝

  Scans your source code and finds env vars
  your code calls but don't exist in your .env.

  Offline. No cloud. One binary.`,
	SilenceUsage: true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(ghostCmd)
	rootCmd.AddCommand(snapCmd)
	rootCmd.AddCommand(driftCmd)
	rootCmd.AddCommand(scoreCmd)
	rootCmd.AddCommand(shareCmd)
	rootCmd.AddCommand(versionCmd)
}
