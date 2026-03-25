package imgpalette

import (
	"errors"
	"image"
	"image/color"
)

// Quantize returns a copy of img redrawn with colors from the extracted palette.
func Quantize(img image.Image, opts ...Option) (image.Image, error) {
	palette, err := Extract(img, opts...)
	if err != nil {
		if errors.Is(err, ErrNoColors) {
			return image.NewNRGBA(img.Bounds()), nil
		}
		return nil, err
	}

	bounds := img.Bounds()
	quantized := image.NewNRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			source := ToRGBA(img.At(x, y))
			if source.A == 0 {
				quantized.Set(x, y, color.NRGBA{})
				continue
			}

			nearest := palette.Nearest(source)
			quantized.SetNRGBA(x, y, color.NRGBA{
				R: nearest.RGBA.R,
				G: nearest.RGBA.G,
				B: nearest.RGBA.B,
				A: source.A,
			})
		}
	}

	return quantized, nil
}
