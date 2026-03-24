// Package sample contains internal image sampling helpers.
package sample

import "image"

// ResizeToMaxSide downsamples image to fit maxSide preserving aspect ratio.
// It never upscales: images already within bounds are returned as-is.
func ResizeToMaxSide(img image.Image, maxSide int) image.Image {
	if img == nil || maxSide <= 0 {
		return img
	}

	b := img.Bounds()
	width := b.Dx()
	height := b.Dy()
	if width <= 0 || height <= 0 {
		return img
	}
	if width <= maxSide && height <= maxSide {
		return img
	}

	var newWidth, newHeight int
	if width >= height {
		newWidth = maxSide
		newHeight = int(float64(height) * float64(maxSide) / float64(width))
	} else {
		newHeight = maxSide
		newWidth = int(float64(width) * float64(maxSide) / float64(height))
	}
	if newWidth < 1 {
		newWidth = 1
	}
	if newHeight < 1 {
		newHeight = 1
	}

	dst := image.NewNRGBA(image.Rect(0, 0, newWidth, newHeight))
	for y := 0; y < newHeight; y++ {
		srcY := b.Min.Y + y*height/newHeight
		for x := 0; x < newWidth; x++ {
			srcX := b.Min.X + x*width/newWidth
			dst.Set(x, y, img.At(srcX, srcY))
		}
	}
	return dst
}
