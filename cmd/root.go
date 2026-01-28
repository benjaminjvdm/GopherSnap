package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gophersnap",
	Short: "Go-powered image converter CLI",
	Long: `Batch process images with Go-powered efficiency.
Supports JPG, PNG, WebP, and AVIF formats.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {

}
