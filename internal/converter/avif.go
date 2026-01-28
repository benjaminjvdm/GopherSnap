package converter

import (
	"image"
	"io"

	"github.com/gen2brain/avif"
)

type AVIFEncoder struct{}

func (e *AVIFEncoder) Encode(w io.Writer, img image.Image, quality int) error {
	if quality <= 0 {
		quality = 75
	}

	return avif.Encode(w, img, avif.Options{Quality: quality})
}
