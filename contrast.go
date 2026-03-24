package imgpalette

import (
	"image/color"
	"math"
)

var (
	blackText = color.RGBA{R: 0, G: 0, B: 0, A: 255}
	whiteText = color.RGBA{R: 255, G: 255, B: 255, A: 255}
)

// Contrast returns WCAG contrast ratio between two colors.
func Contrast(a, b color.Color) float64 {
	la := relativeLuminance(ToRGBA(a))
	lb := relativeLuminance(ToRGBA(b))
	if la < lb {
		la, lb = lb, la
	}
	return (la + 0.05) / (lb + 0.05)
}

// Distance returns normalized RGB Euclidean distance in [0,1].
func Distance(a, b color.Color) float64 {
	left := ToRGBA(a)
	right := ToRGBA(b)
	redDelta := float64(int(left.R) - int(right.R))
	greenDelta := float64(int(left.G) - int(right.G))
	blueDelta := float64(int(left.B) - int(right.B))
	distance := math.Sqrt(redDelta*redDelta + greenDelta*greenDelta + blueDelta*blueDelta)
	return distance / (math.Sqrt(3) * 255.0)
}

// BestTextColor returns black or white, whichever has higher contrast on background.
func BestTextColor(background color.Color) color.RGBA {
	blackContrast := Contrast(background, blackText)
	whiteContrast := Contrast(background, whiteText)
	if blackContrast >= whiteContrast {
		return blackText
	}
	return whiteText
}

func relativeLuminance(c color.RGBA) float64 {
	red := linearizeChannel(c.R)
	green := linearizeChannel(c.G)
	blue := linearizeChannel(c.B)
	return 0.2126*red + 0.7152*green + 0.0722*blue
}

func linearizeChannel(channel uint8) float64 {
	normalized := float64(channel) / 255.0
	if normalized <= 0.04045 {
		return normalized / 12.92
	}
	return math.Pow((normalized+0.055)/1.055, 2.4)
}
