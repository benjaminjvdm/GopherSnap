package converter

import (
	"bytes"
	"fmt"
	"image"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/image/draw"
)

type Result struct {
	InputPath  string
	OutputPath string
	Error      error
}

func resizeImage(img image.Image, width, height int) image.Image {
	bounds := img.Bounds()
	origWidth := bounds.Dx()
	origHeight := bounds.Dy()

	if width == 0 && height == 0 {
		return img
	}

	targetWidth := width
	targetHeight := height

	if width > 0 && height == 0 {
		targetHeight = int(float64(origHeight) * float64(width) / float64(origWidth))
	} else if width == 0 && height > 0 {
		targetWidth = int(float64(origWidth) * float64(height) / float64(origHeight))
	} else {
		ratioW := float64(width) / float64(origWidth)
		ratioH := float64(height) / float64(origHeight)
		ratio := ratioW
		if ratioH < ratioW {
			ratio = ratioH
		}
		targetWidth = int(float64(origWidth) * ratio)
		targetHeight = int(float64(origHeight) * ratio)
	}

	newImg := image.NewRGBA(image.Rect(0, 0, targetWidth, targetHeight))
	draw.CatmullRom.Scale(newImg, newImg.Bounds(), img, img.Bounds(), draw.Over, nil)
	return newImg
}

func Convert(inputPath, outputDir string, opts Options) Result {
	img, _, err := LoadImage(inputPath)
	if err != nil {
		return Result{InputPath: inputPath, Error: err}
	}

	if opts.Width > 0 || opts.Height > 0 {
		img = resizeImage(img, opts.Width, opts.Height)
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return Result{InputPath: inputPath, Error: err}
	}

	ext := "." + string(opts.Format)
	base := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
	outputPath := filepath.Join(outputDir, base+ext)

	if !opts.Overwrite {
		if _, err := os.Stat(outputPath); err == nil {
			return Result{InputPath: inputPath, OutputPath: outputPath, Error: fmt.Errorf("file already exists")}
		}
	}

	var encoder ImageEncoder
	switch opts.Format {
	case FormatJPG:
		encoder = &JPEGEncoder{}
	case FormatPNG:
		encoder = &PNGEncoder{}
	case FormatWebP:
		encoder = &WebPEncoder{}
	case FormatAVIF:
		encoder = &AVIFEncoder{}
	default:
		return Result{InputPath: inputPath, Error: fmt.Errorf("unsupported format: %s", opts.Format)}
	}

	var buf bytes.Buffer
	currentQuality := opts.Quality

	for {
		buf.Reset()
		err := encoder.Encode(&buf, img, currentQuality)
		if err != nil {
			return Result{InputPath: inputPath, Error: err}
		}

		if opts.MaxSize <= 0 || int64(buf.Len()) <= opts.MaxSize || currentQuality <= 10 {
			break
		}

		currentQuality -= 5
		if currentQuality < 10 {
			currentQuality = 10
		}
	}

	out, err := os.Create(outputPath)
	if err != nil {
		return Result{InputPath: inputPath, Error: err}
	}
	defer out.Close()

	_, err = io.Copy(out, &buf)
	if err != nil {
		return Result{InputPath: inputPath, Error: err}
	}

	return Result{InputPath: inputPath, OutputPath: outputPath}
}

func BatchConvert(inputPaths []string, outputDir string, opts Options, jobs int, progress chan<- Result) {
	if jobs <= 0 {
		jobs = 1
	}

	var wg sync.WaitGroup
	paths := make(chan string, len(inputPaths))

	for i := 0; i < jobs; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range paths {
				res := Convert(path, outputDir, opts)
				if progress != nil {
					progress <- res
				}
			}
		}()
	}

	for _, path := range inputPaths {
		paths <- path
	}
	close(paths)

	wg.Wait()
	if progress != nil {
		close(progress)
	}
}
