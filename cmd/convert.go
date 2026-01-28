package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
	"github.com/benjaminjvdm/GopherSnap/internal/converter"
)

var (
	inputPath    string
	outputDir    string
	targetFormat string
	quality      int
	jobs         int
	overwrite    bool
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#135bec")).
			MarginTop(1).
			MarginBottom(1)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00")).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true)
)

var convertCmd = &cobra.Command{
	Use:   "convert",
	Short: "Convert images to a specified format",
	Run: func(cmd *cobra.Command, args []string) {
		files, err := findFiles(inputPath)
		if err != nil {
			fmt.Printf("Error finding files: %v\n", err)
			os.Exit(1)
		}

		if len(files) == 0 {
			fmt.Println("No supported image files found.")
			return
		}

		opts := converter.Options{
			Format:    converter.Format(targetFormat),
			Quality:   quality,
			Overwrite: overwrite,
		}

		fmt.Println(titleStyle.Render("🚀 GopherSnap: Starting Batch Processing"))
		fmt.Printf("Converting %d files to %s (quality: %d, workers: %d)\n\n", len(files), targetFormat, quality, jobs)

		bar := progressbar.NewOptions(len(files),
			progressbar.OptionSetDescription("Processing"),
			progressbar.OptionSetWriter(os.Stderr),
			progressbar.OptionShowCount(),
			progressbar.OptionSetWidth(30),
			progressbar.OptionClearOnFinish(),
			progressbar.OptionSetTheme(progressbar.Theme{
				Saucer:        "█",
				SaucerHead:    "█",
				SaucerPadding: "░",
				BarStart:      "╢",
				BarEnd:        "╟",
			}))

		progress := make(chan converter.Result)
		go converter.BatchConvert(files, outputDir, opts, jobs, progress)

		var results []converter.Result
		for res := range progress {
			results = append(results, res)
			_ = bar.Add(1)
		}

		fmt.Println("\n" + titleStyle.Render("📊 Conversion Summary"))
		successCount := 0
		for _, res := range results {
			if res.Error == nil {
				successCount++
				fmt.Printf("%s %s -> %s\n", successStyle.Render("✔"), filepath.Base(res.InputPath), filepath.Base(res.OutputPath))
			} else {
				fmt.Printf("%s %s: %v\n", errorStyle.Render("✘"), filepath.Base(res.InputPath), res.Error)
			}
		}

		fmt.Printf("\nDone! Successfully converted %d/%d files.\n", successCount, len(files))
	},
}

func findFiles(path string) ([]string, error) {
	var files []string
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if !info.IsDir() {
		return []string{path}, nil
	}

	err = filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			ext := strings.ToLower(filepath.Ext(p))
			if ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".webp" || ext == ".avif" || ext == ".gif" {
				files = append(files, p)
			}
		}
		return nil
	})

	return files, err
}

func init() {
	rootCmd.AddCommand(convertCmd)

	convertCmd.Flags().StringVarP(&inputPath, "input", "i", "", "Input file or directory")
	convertCmd.Flags().StringVarP(&outputDir, "output", "o", "./output", "Output directory")
	convertCmd.Flags().StringVarP(&targetFormat, "format", "f", "webp", "Output format (jpg, png, webp, avif)")
	convertCmd.Flags().IntVarP(&quality, "quality", "q", 80, "Image quality (0-100)")
	convertCmd.Flags().IntVarP(&jobs, "jobs", "j", 4, "Number of concurrent jobs")
	convertCmd.Flags().BoolVar(&overwrite, "overwrite", false, "Overwrite existing files")

	_ = convertCmd.MarkFlagRequired("input")
}
