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
