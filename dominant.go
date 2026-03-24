package imgpalette

import (
	"image"
	"io"
)

// Dominant returns the first (most frequent) color from Extract.
func Dominant(img image.Image, opts ...Option) (Color, error) {
	palette, err := Extract(img, opts...)
	if err != nil {
		return Color{}, err
	}
	return palette[0], nil
}

// DominantReader returns the first (most frequent) color from ExtractReader.
func DominantReader(r io.Reader, opts ...Option) (Color, error) {
	palette, err := ExtractReader(r, opts...)
	if err != nil {
		return Color{}, err
	}
	return palette[0], nil
}

// DominantFile returns the first (most frequent) color from ExtractFile.
func DominantFile(path string, opts ...Option) (Color, error) {
	palette, err := ExtractFile(path, opts...)
	if err != nil {
		return Color{}, err
	}
	return palette[0], nil
}
