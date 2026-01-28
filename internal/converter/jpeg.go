package converter

import (
	"image"
	"image/jpeg"
	"io"
)

type JPEGEncoder struct{}

func (e *JPEGEncoder) Encode(w io.Writer, img image.Image, quality int) error {
	if quality <= 0 {
		quality = 75
	}
	return jpeg.Encode(w, img, &jpeg.Options{Quality: quality})
}
