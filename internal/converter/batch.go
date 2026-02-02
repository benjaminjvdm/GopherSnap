package converter

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Result struct {
	InputPath  string
	OutputPath string
	Error      error
}

func Convert(inputPath, outputDir string, opts Options) Result {
	img, _, err := LoadImage(inputPath)
	if err != nil {
		return Result{InputPath: inputPath, Error: err}
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
