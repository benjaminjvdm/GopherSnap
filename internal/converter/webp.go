package converter

import (
	"image"
	"io"

	"github.com/gen2brain/webp"
)

type WebPEncoder struct{}

func (e *WebPEncoder) Encode(w io.Writer, img image.Image, quality int) error {
	if quality <= 0 {
		quality = 75
	}
	return webp.Encode(w, img, webp.Options{Lossless: false, Quality: quality})
}
