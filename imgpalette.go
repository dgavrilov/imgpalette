// Package imgpalette extracts dominant colors from images using standard library decoders only.
package imgpalette

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"io"
	"math"
	"os"
	"sort"

	// Register GIF decoder for image.Decode.
	_ "image/gif"
	// Register JPEG decoder for image.Decode.
	_ "image/jpeg"
	// Register PNG decoder for image.Decode.
	_ "image/png"
)

const (
	quantizationBits     = 5
	quantizationMask     = (1 << quantizationBits) - 1
	quantizedBinsPerAxis = 1 << quantizationBits
	totalQuantizedBins   = quantizedBinsPerAxis * quantizedBinsPerAxis * quantizedBinsPerAxis
)

// ErrInvalidMaxColors is returned when maxColors <= 0 in reader/path APIs.
var ErrInvalidMaxColors = errors.New("maxColors must be > 0")

// Color represents an RGB color.
type Color struct {
	R uint8
	G uint8
	B uint8
}

// String returns the lowercase hex representation (for example: #ffffdd).
func (c Color) String() string {
	return c.HexString()
}

// HexString returns the lowercase hex representation (for example: #ffffdd).
func (c Color) HexString() string {
	return fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B)
}

// RGBString returns CSS-like rgb() representation (for example: rgb(255,255,221)).
func (c Color) RGBString() string {
	return fmt.Sprintf("rgb(%d,%d,%d)", c.R, c.G, c.B)
}

// RGBAString returns CSS-like rgba() representation (for example: rgba(255,255,221,0.25)).
func (c Color) RGBAString(alpha float64) string {
	return fmt.Sprintf("rgba(%d,%d,%d,%s)", c.R, c.G, c.B, formatAlpha(alpha))
}

// Int returns packed RGB integer value (for example: 0xffffdd == 16777181).
func (c Color) Int() uint32 {
	return uint32(c.R)<<16 | uint32(c.G)<<8 | uint32(c.B)
}

// RGB returns channel values as a 3-element array.
func (c Color) RGB() [3]uint8 {
	return [3]uint8{c.R, c.G, c.B}
}

// RGBA returns channel values + alpha as a 4-element float array.
func (c Color) RGBA(alpha float64) [4]float64 {
	return [4]float64{float64(c.R), float64(c.G), float64(c.B), clampAlpha(alpha)}
}

type quantizedBinStats struct {
	binIndex   int
	pixelCount uint32
	redSum     uint64
	greenSum   uint64
	blueSum    uint64
}

// ExtractDominantColorsFromReader decodes an image and returns up to maxColors dominant colors.
// sampleStride controls sampling density; sampleStride <= 1 means every pixel.
func ExtractDominantColorsFromReader(imageReader io.Reader, maxColors int, sampleStride int) ([]Color, error) {
	if maxColors <= 0 {
		return nil, ErrInvalidMaxColors
	}

	decodedImage, _, err := image.Decode(imageReader)
	if err != nil {
		return nil, err
	}

	return ExtractDominantColorsFromImage(decodedImage, maxColors, sampleStride), nil
}

// ExtractDominantColorsFromPath opens an image file and returns up to maxColors dominant colors.
// sampleStride controls sampling density; sampleStride <= 1 means every pixel.
func ExtractDominantColorsFromPath(imagePath string, maxColors int, sampleStride int) ([]Color, error) {
	if maxColors <= 0 {
		return nil, ErrInvalidMaxColors
	}

	// #nosec G304 -- API intentionally accepts caller-provided file paths.
	imageFile, err := os.Open(imagePath)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = imageFile.Close()
	}()

	return ExtractDominantColorsFromReader(imageFile, maxColors, sampleStride)
}

// ExtractDominantColorsFromImage returns up to maxColors dominant colors from an image.
// sampleStride controls sampling density; sampleStride <= 1 means every pixel.
func ExtractDominantColorsFromImage(img image.Image, maxColors int, sampleStride int) []Color {
	if img == nil || maxColors <= 0 {
		return nil
	}
	if sampleStride <= 1 {
		sampleStride = 1
	}

	var binPixelCounts [totalQuantizedBins]uint32
	var binRedSums [totalQuantizedBins]uint64
	var binGreenSums [totalQuantizedBins]uint64
	var binBlueSums [totalQuantizedBins]uint64

	switch typedImage := img.(type) {
	case *image.NRGBA:
		accumulateFromNRGBA(typedImage, sampleStride, &binPixelCounts, &binRedSums, &binGreenSums, &binBlueSums)
	case *image.RGBA:
		accumulateFromRGBA(typedImage, sampleStride, &binPixelCounts, &binRedSums, &binGreenSums, &binBlueSums)
	case *image.Gray:
		accumulateFromGray(typedImage, sampleStride, &binPixelCounts, &binRedSums, &binGreenSums, &binBlueSums)
	case *image.Paletted:
		accumulateFromPaletted(typedImage, sampleStride, &binPixelCounts, &binRedSums, &binGreenSums, &binBlueSums)
	default:
		accumulateFromGenericImage(img, sampleStride, &binPixelCounts, &binRedSums, &binGreenSums, &binBlueSums)
	}

	nonEmptyBins := make([]quantizedBinStats, 0, totalQuantizedBins)
	for binIndex, pixelCount := range binPixelCounts {
		if pixelCount == 0 {
			continue
		}
		nonEmptyBins = append(nonEmptyBins, quantizedBinStats{
			binIndex:   binIndex,
			pixelCount: pixelCount,
			redSum:     binRedSums[binIndex],
			greenSum:   binGreenSums[binIndex],
			blueSum:    binBlueSums[binIndex],
		})
	}

	if len(nonEmptyBins) == 0 {
		return nil
	}

	sort.Slice(nonEmptyBins, func(i, j int) bool {
		if nonEmptyBins[i].pixelCount == nonEmptyBins[j].pixelCount {
			return nonEmptyBins[i].binIndex < nonEmptyBins[j].binIndex
		}
		return nonEmptyBins[i].pixelCount > nonEmptyBins[j].pixelCount
	})

	if maxColors > len(nonEmptyBins) {
		maxColors = len(nonEmptyBins)
	}

	dominantColors := make([]Color, 0, maxColors)
	for i := 0; i < maxColors; i++ {
		binStats := nonEmptyBins[i]
		pixelCount := uint64(binStats.pixelCount)
		dominantColors = append(dominantColors, Color{
			R: uint8Clamped(binStats.redSum / pixelCount),
			G: uint8Clamped(binStats.greenSum / pixelCount),
			B: uint8Clamped(binStats.blueSum / pixelCount),
		})
	}

	return dominantColors
}

// QuantizedBinIndex returns the 5-5-5 quantized bin index for this color.
func (c Color) QuantizedBinIndex() int {
	quantizedRed := int(c.R>>(8-quantizationBits)) & quantizationMask
	quantizedGreen := int(c.G>>(8-quantizationBits)) & quantizationMask
	quantizedBlue := int(c.B>>(8-quantizationBits)) & quantizationMask
	return (quantizedRed << (quantizationBits * 2)) | (quantizedGreen << quantizationBits) | quantizedBlue
}

func clampAlpha(alpha float64) float64 {
	if math.IsNaN(alpha) || alpha < 0 {
		return 0
	}
	if alpha > 1 {
		return 1
	}
	return alpha
}

func formatAlpha(alpha float64) string {
	clamped := clampAlpha(alpha)
	return fmt.Sprintf("%g", clamped)
}

func addPixelToQuantizedBin(red, green, blue uint8, binPixelCounts *[totalQuantizedBins]uint32, binRedSums, binGreenSums, binBlueSums *[totalQuantizedBins]uint64) {
	quantizedRed := int(red >> (8 - quantizationBits))
	quantizedGreen := int(green >> (8 - quantizationBits))
	quantizedBlue := int(blue >> (8 - quantizationBits))
	binIndex := (quantizedRed << (quantizationBits * 2)) | (quantizedGreen << quantizationBits) | quantizedBlue

	binPixelCounts[binIndex]++
	binRedSums[binIndex] += uint64(red)
	binGreenSums[binIndex] += uint64(green)
	binBlueSums[binIndex] += uint64(blue)
}

func accumulateFromGenericImage(img image.Image, sampleStride int, binPixelCounts *[totalQuantizedBins]uint32, binRedSums, binGreenSums, binBlueSums *[totalQuantizedBins]uint64) {
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y += sampleStride {
		for x := bounds.Min.X; x < bounds.Max.X; x += sampleStride {
			red16, green16, blue16, alpha16 := img.At(x, y).RGBA()
			if alpha16 == 0 {
				continue
			}
			addPixelToQuantizedBin(uint8FromRGBA32(red16), uint8FromRGBA32(green16), uint8FromRGBA32(blue16), binPixelCounts, binRedSums, binGreenSums, binBlueSums)
		}
	}
}

func accumulateFromNRGBA(img *image.NRGBA, sampleStride int, binPixelCounts *[totalQuantizedBins]uint32, binRedSums, binGreenSums, binBlueSums *[totalQuantizedBins]uint64) {
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y += sampleStride {
		rowOffset := (y-bounds.Min.Y)*img.Stride - bounds.Min.X*4
		for x := bounds.Min.X; x < bounds.Max.X; x += sampleStride {
			pixelOffset := rowOffset + x*4
			alpha := img.Pix[pixelOffset+3]
			if alpha == 0 {
				continue
			}
			addPixelToQuantizedBin(img.Pix[pixelOffset], img.Pix[pixelOffset+1], img.Pix[pixelOffset+2], binPixelCounts, binRedSums, binGreenSums, binBlueSums)
		}
	}
}

func accumulateFromRGBA(img *image.RGBA, sampleStride int, binPixelCounts *[totalQuantizedBins]uint32, binRedSums, binGreenSums, binBlueSums *[totalQuantizedBins]uint64) {
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y += sampleStride {
		rowOffset := (y-bounds.Min.Y)*img.Stride - bounds.Min.X*4
		for x := bounds.Min.X; x < bounds.Max.X; x += sampleStride {
			pixelOffset := rowOffset + x*4
			alpha := img.Pix[pixelOffset+3]
			if alpha == 0 {
				continue
			}

			red, green, blue := img.Pix[pixelOffset], img.Pix[pixelOffset+1], img.Pix[pixelOffset+2]
			if alpha < 255 {
				// RGBA stores alpha-premultiplied channels; undo to recover display color.
				red = unpremultiplyChannel(red, alpha)
				green = unpremultiplyChannel(green, alpha)
				blue = unpremultiplyChannel(blue, alpha)
			}
			addPixelToQuantizedBin(red, green, blue, binPixelCounts, binRedSums, binGreenSums, binBlueSums)
		}
	}
}

func accumulateFromGray(img *image.Gray, sampleStride int, binPixelCounts *[totalQuantizedBins]uint32, binRedSums, binGreenSums, binBlueSums *[totalQuantizedBins]uint64) {
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y += sampleStride {
		rowOffset := (y-bounds.Min.Y)*img.Stride - bounds.Min.X
		for x := bounds.Min.X; x < bounds.Max.X; x += sampleStride {
			pixelOffset := rowOffset + x
			gray := img.Pix[pixelOffset]
			addPixelToQuantizedBin(gray, gray, gray, binPixelCounts, binRedSums, binGreenSums, binBlueSums)
		}
	}
}

func accumulateFromPaletted(img *image.Paletted, sampleStride int, binPixelCounts *[totalQuantizedBins]uint32, binRedSums, binGreenSums, binBlueSums *[totalQuantizedBins]uint64) {
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y += sampleStride {
		rowOffset := (y-bounds.Min.Y)*img.Stride - bounds.Min.X
		for x := bounds.Min.X; x < bounds.Max.X; x += sampleStride {
			pixelOffset := rowOffset + x
			paletteColor := img.Palette[img.Pix[pixelOffset]]

			switch typedColor := paletteColor.(type) {
			case color.NRGBA:
				if typedColor.A == 0 {
					continue
				}
				addPixelToQuantizedBin(typedColor.R, typedColor.G, typedColor.B, binPixelCounts, binRedSums, binGreenSums, binBlueSums)
			case color.RGBA:
				if typedColor.A == 0 {
					continue
				}
				red, green, blue := typedColor.R, typedColor.G, typedColor.B
				if typedColor.A < 255 {
					red = unpremultiplyChannel(red, typedColor.A)
					green = unpremultiplyChannel(green, typedColor.A)
					blue = unpremultiplyChannel(blue, typedColor.A)
				}
				addPixelToQuantizedBin(red, green, blue, binPixelCounts, binRedSums, binGreenSums, binBlueSums)
			default:
				red16, green16, blue16, alpha16 := paletteColor.RGBA()
				if alpha16 == 0 {
					continue
				}
				addPixelToQuantizedBin(uint8FromRGBA32(red16), uint8FromRGBA32(green16), uint8FromRGBA32(blue16), binPixelCounts, binRedSums, binGreenSums, binBlueSums)
			}
		}
	}
}

func uint8FromRGBA32(channel uint32) uint8 {
	return uint8Clamped(uint64(channel >> 8))
}

func unpremultiplyChannel(channel, alpha uint8) uint8 {
	if alpha == 0 {
		return 0
	}
	value := (uint32(channel) * 255) / uint32(alpha)
	return uint8Clamped(uint64(value))
}

func uint8Clamped(value uint64) uint8 {
	if value > 255 {
		return 255
	}
	return uint8(value)
}
