package cmd

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/benjaminjvdm/GopherSnap/internal/converter"
)

func createTestPNG(t *testing.T, dir, filename string, width, height int) string {
	t.Helper()
	imgPath := filepath.Join(dir, filename)
	if err := os.MkdirAll(filepath.Dir(imgPath), 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			img.Set(x, y, color.RGBA{R: uint8(x % 256), G: uint8(y % 256), B: 128, A: 255})
		}
	}

	f, err := os.Create(imgPath)
	if err != nil {
		t.Fatalf("failed to create image file: %v", err)
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		t.Fatalf("failed to encode png: %v", err)
	}

	return imgPath
}

func TestFindFilesSanitization(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gophersnap-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	filesToCreate := []string{
		"test1.jpg",
		filepath.Join("subdir", "test2.PNG"),
	}

	for _, f := range filesToCreate {
		fullPath := filepath.Join(tempDir, f)
		err := os.MkdirAll(filepath.Dir(fullPath), 0755)
		if err != nil {
			t.Fatalf("Failed to create temp subdir: %v", err)
		}
		err = os.WriteFile(fullPath, []byte("dummy image content"), 0644)
		if err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}
	}

	tests := []struct {
		name          string
		inputPath     string
		expectedFiles []string
		expectError   bool
	}{
		{
			name:      "Path with dot dot",
			inputPath: tempDir + string(filepath.Separator) + "subdir" + string(filepath.Separator) + "..",
			expectedFiles: []string{
				"test1.jpg",
				"test2.PNG",
			},
			expectError: false,
		},
		{
			name:      "Path with dot",
			inputPath: tempDir + string(filepath.Separator) + ".",
			expectedFiles: []string{
				"test1.jpg",
				"test2.PNG",
			},
			expectError: false,
		},
		{
			name:      "Direct file path",
			inputPath: tempDir + string(filepath.Separator) + "subdir" + string(filepath.Separator) + ".." + string(filepath.Separator) + "test1.jpg",
			expectedFiles: []string{
				"test1.jpg",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files, err := findFiles(tt.inputPath)
			if (err != nil) != tt.expectError {
				t.Fatalf("expected error: %v, got error: %v", tt.expectError, err)
			}

			if len(files) != len(tt.expectedFiles) {
				t.Fatalf("expected %d files, got %d", len(tt.expectedFiles), len(files))
			}

			foundMap := make(map[string]bool)
			for _, f := range files {
				foundMap[filepath.Base(f)] = true
			}

			for _, ef := range tt.expectedFiles {
				if !foundMap[ef] {
					t.Errorf("expected to find file %s, but it was missing", ef)
				}
			}
		})
	}
}

func TestFormatConversions(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gophersnap-fmt-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	srcPath := createTestPNG(t, tempDir, "sample.png", 200, 200)
	outputDir := filepath.Join(tempDir, "output")

	formats := []converter.Format{
		converter.FormatJPG,
		converter.FormatPNG,
		converter.FormatWebP,
		converter.FormatAVIF,
	}

	for _, fmt := range formats {
		t.Run("Convert to "+string(fmt), func(t *testing.T) {
			opts := converter.Options{
				Format:    fmt,
				Quality:   80,
				Overwrite: true,
			}
			res := converter.Convert(srcPath, outputDir, opts)
			if res.Error != nil {
				t.Fatalf("conversion to %s failed: %v", fmt, res.Error)
			}

			if _, err := os.Stat(res.OutputPath); os.IsNotExist(err) {
				t.Fatalf("output file for %s does not exist: %s", fmt, res.OutputPath)
			}
		})
	}
}

func TestSizingAndQualityConstraints(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gophersnap-size-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	srcPath := createTestPNG(t, tempDir, "large.png", 200, 200)
	outputDir := filepath.Join(tempDir, "output")

	t.Run("Width and Height Resizing", func(t *testing.T) {
		opts := converter.Options{
			Format:    converter.FormatJPG,
			Quality:   80,
			Width:     100,
			Height:    100,
			Overwrite: true,
		}
		res := converter.Convert(srcPath, outputDir, opts)
		if res.Error != nil {
			t.Fatalf("resizing failed: %v", res.Error)
		}

		img, _, err := converter.LoadImage(res.OutputPath)
		if err != nil {
			t.Fatalf("failed to load resized image: %v", err)
		}
		bounds := img.Bounds()
		if bounds.Dx() != 100 || bounds.Dy() != 100 {
			t.Errorf("expected 100x100 resized image, got %dx%d", bounds.Dx(), bounds.Dy())
		}
	})

	t.Run("MaxSize Optimization", func(t *testing.T) {
		optsLimited := converter.Options{
			Format:    converter.FormatJPG,
			Quality:   95,
			MaxSize:   2000,
			Overwrite: true,
		}
		resLimited := converter.Convert(srcPath, filepath.Join(outputDir, "limited"), optsLimited)
		if resLimited.Error != nil {
			t.Fatalf("max-size conversion failed: %v", resLimited.Error)
		}

		optsUnlimited := converter.Options{
			Format:    converter.FormatJPG,
			Quality:   95,
			MaxSize:   0,
			Overwrite: true,
		}
		resUnlimited := converter.Convert(srcPath, filepath.Join(outputDir, "unlimited"), optsUnlimited)
		if resUnlimited.Error != nil {
			t.Fatalf("unlimited conversion failed: %v", resUnlimited.Error)
		}

		infoLimited, _ := os.Stat(resLimited.OutputPath)
		infoUnlimited, _ := os.Stat(resUnlimited.OutputPath)

		if infoLimited.Size() >= infoUnlimited.Size() {
			t.Errorf("expected limited size (%d) to be smaller than unlimited high-quality size (%d)", infoLimited.Size(), infoUnlimited.Size())
		}
	})

	t.Run("Quality Setting Comparison", func(t *testing.T) {
		optsLow := converter.Options{Format: converter.FormatJPG, Quality: 20, Overwrite: true}
		optsHigh := converter.Options{Format: converter.FormatJPG, Quality: 95, Overwrite: true}

		outLowDir := filepath.Join(outputDir, "low")
		outHighDir := filepath.Join(outputDir, "high")

		resLow := converter.Convert(srcPath, outLowDir, optsLow)
		resHigh := converter.Convert(srcPath, outHighDir, optsHigh)

		if resLow.Error != nil || resHigh.Error != nil {
			t.Fatalf("quality conversion error: low=%v, high=%v", resLow.Error, resHigh.Error)
		}

		infoLow, _ := os.Stat(resLow.OutputPath)
		infoHigh, _ := os.Stat(resHigh.OutputPath)

		if infoLow.Size() >= infoHigh.Size() {
			t.Errorf("expected low quality size (%d) to be smaller than high quality size (%d)", infoLow.Size(), infoHigh.Size())
		}
	})
}

func TestBatchAndSubdirectoryConversion(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gophersnap-batch-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	src1 := createTestPNG(t, tempDir, "img1.png", 100, 100)
	src2 := createTestPNG(t, tempDir, filepath.Join("sub", "img2.PNG"), 150, 150)

	outputDir := filepath.Join(tempDir, "output")

	jobs := []converter.ConverterJob{
		{InputPath: src1, OutputDir: outputDir},
		{InputPath: src2, OutputDir: filepath.Join(outputDir, "sub")},
	}

	opts := converter.Options{
		Format:    converter.FormatWebP,
		Quality:   80,
		Overwrite: true,
	}

	progress := make(chan converter.Result)
	go converter.BatchConvert(jobs, opts, 2, progress)

	count := 0
	for res := range progress {
		if res.Error != nil {
			t.Errorf("batch convert job failed for %s: %v", res.InputPath, res.Error)
		}
		count++
	}

	if count != 2 {
		t.Fatalf("expected 2 converted results, got %d", count)
	}
}

func TestErrorHandling(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gophersnap-err-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	t.Run("Non-existent path in findFiles", func(t *testing.T) {
		_, err := findFiles(filepath.Join(tempDir, "does-not-exist"))
		if err == nil {
			t.Error("expected error for non-existent path, got nil")
		}
	})

	t.Run("Corrupted image conversion", func(t *testing.T) {
		corruptPath := filepath.Join(tempDir, "corrupt.jpg")
		if err := os.WriteFile(corruptPath, []byte("this is not an image file"), 0644); err != nil {
			t.Fatalf("failed to write corrupted file: %v", err)
		}

		res := converter.Convert(corruptPath, filepath.Join(tempDir, "out"), converter.Options{Format: converter.FormatWebP})
		if res.Error == nil {
			t.Error("expected error when converting corrupted file, got nil")
		}
	})
}
