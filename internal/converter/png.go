package converter

import (
	"image"
	"image/png"
	"io"
)

type PNGEncoder struct{}

func (e *PNGEncoder) Encode(w io.Writer, img image.Image, quality int) error {

	encoder := png.Encoder{
		CompressionLevel: png.DefaultCompression,
	}
	if quality > 80 {
		encoder.CompressionLevel = png.BestCompression
	} else if quality < 20 {
		encoder.CompressionLevel = png.BestSpeed
	}
	return encoder.Encode(w, img)
}
