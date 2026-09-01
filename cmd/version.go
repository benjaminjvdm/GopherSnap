package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	Version = "v0.1.0"
	Commit  = "dev"
	Date    = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information for GopherSnap",
	Long:  `Print version information including version number, git commit hash, and build date.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("GopherSnap %s (commit: %s, built at: %s)\n", Version, Commit, Date)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.Version = Version
	rootCmd.SetVersionTemplate("GopherSnap {{.Version}}\n")
}
