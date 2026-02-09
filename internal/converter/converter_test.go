package converter

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestBatchConvert(t *testing.T) {

	tempDir, err := os.MkdirTemp("", "goconv_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	inputDir := filepath.Join(tempDir, "input")
	outputDir := filepath.Join(tempDir, "output")
	os.Mkdir(inputDir, 0755)

	imgPath := filepath.Join(inputDir, "test.png")
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			img.Set(x, y, color.RGBA{255, 0, 0, 255})
		}
	}
	f, err := os.Create(imgPath)
	if err != nil {
		t.Fatal(err)
	}
	png.Encode(f, img)
	f.Close()

	opts := Options{
		Format:  FormatJPG,
		Quality: 80,
	}

	progress := make(chan Result)
	go BatchConvert([]string{imgPath}, outputDir, opts, 1, progress)

	count := 0
	for res := range progress {
		if res.Error != nil {
			t.Errorf("Conversion failed for %s: %v", res.InputPath, res.Error)
		}
		count++
	}

	if count != 1 {
		t.Errorf("Expected 1 conversion, got %d", count)
	}

	if _, err := os.Stat(filepath.Join(outputDir, "test.jpg")); os.IsNotExist(err) {
		t.Errorf("Output file was not created")
	}

	formats := []Format{FormatPNG, FormatWebP, FormatAVIF}
	for _, f := range formats {
		opts.Format = f
		res := Convert(imgPath, outputDir, opts)
		if res.Error != nil {
			t.Errorf("Conversion failed for %s: %v", f, res.Error)
		}
		outputExt := "." + string(f)
		if _, err := os.Stat(filepath.Join(outputDir, "test"+outputExt)); os.IsNotExist(err) {
			t.Errorf("Output file for %s was not created", f)
		}
	}
}

func TestConvertWithMaxSize(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "goconv_maxsize_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, color.RGBA{uint8(x % 256), uint8(y % 256), uint8((x + y) % 256), 255})
		}
	}

	imgPath := filepath.Join(tempDir, "large.png")
	f, err := os.Create(imgPath)
	if err != nil {
		t.Fatal(err)
	}
	png.Encode(f, img)
	f.Close()

	outputDir := filepath.Join(tempDir, "output")

	maxSize := int64(1 * 1024) // 1KB - small enough for 100x100 noise to likely exceed at high quality
	opts := Options{
		Format:  FormatJPG,
		Quality: 95,
		MaxSize: maxSize,
	}

	res := Convert(imgPath, outputDir, opts)
	if res.Error != nil {
		t.Fatalf("Conversion failed: %v", res.Error)
	}

	info, err := os.Stat(res.OutputPath)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Output size with MaxSize: %d", info.Size())

	optsNoLimit := Options{
		Format:  FormatJPG,
		Quality: 95,
		MaxSize: 0,
	}
	resNoLimit := Convert(imgPath, filepath.Join(tempDir, "output_nolimit"), optsNoLimit)
	infoNoLimit, _ := os.Stat(resNoLimit.OutputPath)
	t.Logf("Output size without MaxSize: %d", infoNoLimit.Size())

	if info.Size() >= infoNoLimit.Size() {
		// Only error if we actually needed to reduce quality
		if infoNoLimit.Size() > maxSize {
			t.Errorf("MaxSize version (%d) should be smaller than high quality version (%d) which exceeded the limit", info.Size(), infoNoLimit.Size())
		}
	}

	if info.Size() > maxSize {
		t.Logf("Note: final size %d still exceeds maxSize %d because min quality reached", info.Size(), maxSize)
	}
}

func TestConvertWithResizing(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "goconv_resize_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	img := image.NewRGBA(image.Rect(0, 0, 100, 200))
	imgPath := filepath.Join(tempDir, "original.png")
	f, err := os.Create(imgPath)
	if err != nil {
		t.Fatal(err)
	}
	png.Encode(f, img)
	f.Close()

	outputDir := filepath.Join(tempDir, "output")

	tests := []struct {
		name           string
		opts           Options
		expectedWidth  int
		expectedHeight int
	}{
		{
			name:           "Resize width only",
			opts:           Options{Width: 50, Format: FormatPNG, Overwrite: true},
			expectedWidth:  50,
			expectedHeight: 100,
		},
		{
			name:           "Resize height only",
			opts:           Options{Height: 100, Format: FormatPNG, Overwrite: true},
			expectedWidth:  50,
			expectedHeight: 100,
		},
		{
			name:           "Resize both (fit width)",
			opts:           Options{Width: 40, Height: 100, Format: FormatPNG, Overwrite: true},
			expectedWidth:  40,
			expectedHeight: 80,
		},
		{
			name:           "Resize both (fit height)",
			opts:           Options{Width: 100, Height: 80, Format: FormatPNG, Overwrite: true},
			expectedWidth:  40,
			expectedHeight: 80,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := Convert(imgPath, outputDir, tt.opts)
			if res.Error != nil {
				t.Fatalf("Conversion failed: %v", res.Error)
			}

			outputImg, _, err := LoadImage(res.OutputPath)
			if err != nil {
				t.Fatalf("Failed to load output image: %v", err)
			}

			bounds := outputImg.Bounds()
			if bounds.Dx() != tt.expectedWidth || bounds.Dy() != tt.expectedHeight {
				t.Errorf("Expected size %dx%d, got %dx%d", tt.expectedWidth, tt.expectedHeight, bounds.Dx(), bounds.Dy())
			}
		})
	}
}
