package imgpalette

import (
	"fmt"
	"image/color"
	"math"

	iconvert "github.com/dgavrilov/imgpalette/internal/convert"
)

// Color represents a palette color with occurrence stats.
type Color struct {
	RGBA  color.RGBA
	Count int
	Ratio float64
}

// String returns the lowercase hex representation (for example: #ffffdd).
func (c Color) String() string {
	return c.Hex()
}

// Hex returns the lowercase hex representation (for example: #ffffdd).
func (c Color) Hex() string {
	return fmt.Sprintf("#%02x%02x%02x", c.RGBA.R, c.RGBA.G, c.RGBA.B)
}

// RGBString returns CSS-like rgb() representation (for example: rgb(255,255,221)).
func (c Color) RGBString() string {
	return fmt.Sprintf("rgb(%d,%d,%d)", c.RGBA.R, c.RGBA.G, c.RGBA.B)
}

// RGBAString returns CSS-like rgba() representation (for example: rgba(255,255,221,0.25)).
func (c Color) RGBAString(alpha float64) string {
	return fmt.Sprintf("rgba(%d,%d,%d,%s)", c.RGBA.R, c.RGBA.G, c.RGBA.B, formatAlpha(alpha))
}

// Int returns packed RGB integer value (for example: 0xffffdd == 16777181).
func (c Color) Int() uint32 {
	return uint32(c.RGBA.R)<<16 | uint32(c.RGBA.G)<<8 | uint32(c.RGBA.B)
}

// RGB returns channel values as a 3-element array.
func (c Color) RGB() [3]uint8 {
	return [3]uint8{c.RGBA.R, c.RGBA.G, c.RGBA.B}
}

// RGBAValues returns channel values + alpha as a 4-element float array.
func (c Color) RGBAValues(alpha float64) [4]float64 {
	return [4]float64{float64(c.RGBA.R), float64(c.RGBA.G), float64(c.RGBA.B), clampAlpha(alpha)}
}

// QuantizedBinIndex returns the 5-5-5 quantized bin index for this color.
func (c Color) QuantizedBinIndex() int {
	quantizedRed := int(c.RGBA.R>>(8-quantizationBits)) & quantizationMask
	quantizedGreen := int(c.RGBA.G>>(8-quantizationBits)) & quantizationMask
	quantizedBlue := int(c.RGBA.B>>(8-quantizationBits)) & quantizationMask
	return (quantizedRed << (quantizationBits * 2)) | (quantizedGreen << quantizationBits) | quantizedBlue
}

// ToRGBA converts any color.Color to 8-bit non-premultiplied RGBA.
func ToRGBA(c color.Color) color.RGBA {
	return iconvert.ToRGBA(c)
}

// Hex returns lowercase hex representation of RGB channels (#rrggbb).
func Hex(c color.Color) string {
	rgba := ToRGBA(c)
	return fmt.Sprintf("#%02x%02x%02x", rgba.R, rgba.G, rgba.B)
}

// Brightness returns perceived luminance in [0,1].
func Brightness(c color.Color) float64 {
	rgba := ToRGBA(c)
	red := float64(rgba.R) / 255.0
	green := float64(rgba.G) / 255.0
	blue := float64(rgba.B) / 255.0
	return 0.299*red + 0.587*green + 0.114*blue
}

// Saturation returns HSV saturation in [0,1].
func Saturation(c color.Color) float64 {
	rgba := ToRGBA(c)
	red := float64(rgba.R) / 255.0
	green := float64(rgba.G) / 255.0
	blue := float64(rgba.B) / 255.0

	maxValue := maxFloat(red, green, blue)
	minValue := minFloat(red, green, blue)
	if maxValue == 0 {
		return 0
	}
	return (maxValue - minValue) / maxValue
}

// IsGray reports whether saturation is below or equal to threshold.
func IsGray(c color.Color, threshold float64) bool {
	if threshold < 0 {
		threshold = 0
	}
	if threshold > 1 {
		threshold = 1
	}
	return Saturation(c) <= threshold
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
	return fmt.Sprintf("%g", clampAlpha(alpha))
}

func maxFloat(a, b, c float64) float64 {
	maxValue := a
	if b > maxValue {
		maxValue = b
	}
	if c > maxValue {
		maxValue = c
	}
	return maxValue
}

func minFloat(a, b, c float64) float64 {
	minValue := a
	if b < minValue {
		minValue = b
	}
	if c < minValue {
		minValue = c
	}
	return minValue
}
