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
	maxSize      string
	width        int
	height       int
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
		inputPath = filepath.Clean(inputPath)
		outputDir = filepath.Clean(outputDir)

		files, err := findFiles(inputPath)
		if err != nil {
			fmt.Printf("Error finding files: %v\n", err)
			os.Exit(1)
		}

		if len(files) == 0 {
			fmt.Println("No supported image files found.")
			return
		}

		baseInputDir := inputPath
		info, err := os.Stat(inputPath)
		if err == nil && !info.IsDir() {
			baseInputDir = filepath.Dir(inputPath)
		}

		var jobsList []converter.ConverterJob
		for _, file := range files {
			relPath, err := filepath.Rel(baseInputDir, filepath.Dir(file))
			if err != nil {
				relPath = "."
			}
			jobOutputDir := filepath.Clean(filepath.Join(outputDir, relPath))
			jobsList = append(jobsList, converter.ConverterJob{
				InputPath: file,
				OutputDir: jobOutputDir,
			})
		}

		// Pre-create directory structure
		uniqueDirs := make(map[string]bool)
		for _, job := range jobsList {
			uniqueDirs[job.OutputDir] = true
		}
		for dir := range uniqueDirs {
			if err := os.MkdirAll(dir, 0755); err != nil {
				fmt.Printf("Error creating directory %s: %v\n", dir, err)
				os.Exit(1)
			}
		}

		var parsedMaxSize int64
		if width < 0 || height < 0 {
			fmt.Println(errorStyle.Render("Error: width and height must be non-negative"))
			os.Exit(1)
		}

		if maxSize != "" {
			var err error
			parsedMaxSize, err = converter.ParseSize(maxSize)
			if err != nil {
				fmt.Printf("Error parsing max-size: %v\n", err)
				os.Exit(1)
			}
			if !cmd.Flags().Changed("quality") {
				quality = 90
			}
		}

		opts := converter.Options{
			Format:    converter.Format(targetFormat),
			Quality:   quality,
			MaxSize:   parsedMaxSize,
			Overwrite: overwrite,
			Width:     width,
			Height:    height,
		}

		fmt.Println(titleStyle.Render("🚀 GopherSnap: Starting Batch Processing"))
		fmt.Printf("Converting %d files to %s (quality: %d, workers: %d)\n\n", len(jobsList), targetFormat, quality, jobs)

		bar := progressbar.NewOptions(len(jobsList),
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
		go converter.BatchConvert(jobsList, opts, jobs, progress)

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
	cleanPath := filepath.Clean(path)
	var files []string
	info, err := os.Stat(cleanPath)
	if err != nil {
		return nil, err
	}

	if !info.IsDir() {
		return []string{cleanPath}, nil
	}

	err = filepath.Walk(cleanPath, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			ext := strings.ToLower(filepath.Ext(p))
			if ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".webp" || ext == ".avif" || ext == ".gif" {
				files = append(files, filepath.Clean(p))
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
	convertCmd.Flags().StringVar(&maxSize, "max-size", "", "Maximum allowed output file size (e.g., 200kb, 1mb)")
	convertCmd.Flags().IntVar(&width, "width", 0, "Target width (maintaining aspect ratio)")
	convertCmd.Flags().IntVar(&height, "height", 0, "Target height (maintaining aspect ratio)")
	convertCmd.Flags().IntVarP(&jobs, "jobs", "j", 4, "Number of concurrent jobs")
	convertCmd.Flags().BoolVar(&overwrite, "overwrite", false, "Overwrite existing files")

	_ = convertCmd.MarkFlagRequired("input")
}
