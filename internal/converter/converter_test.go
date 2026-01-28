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
