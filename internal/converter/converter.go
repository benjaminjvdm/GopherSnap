package converter

import (
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"

	_ "github.com/gen2brain/avif"
	_ "github.com/gen2brain/webp"
)

type Format string

const (
	FormatJPG  Format = "jpg"
	FormatPNG  Format = "png"
	FormatWebP Format = "webp"
	FormatAVIF Format = "avif"
)

type Options struct {
	Format       Format
	Quality      int
	Overwrite    bool
	PreserveMeta bool
}

type ImageEncoder interface {
	Encode(w io.Writer, img image.Image, quality int) error
}

func LoadImage(path string) (image.Image, string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, "", err
	}
	defer file.Close()

	img, format, err := image.Decode(file)
	if err != nil {
		return nil, "", err
	}
	return img, format, nil
}
