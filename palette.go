package imgpalette

import (
	"image/color"
	"sort"
)

// Palette is a collection of palette colors.
type Palette []Color

// Colors returns RGBA colors from the palette preserving order.
func (p Palette) Colors() []color.RGBA {
	colors := make([]color.RGBA, 0, len(p))
	for _, paletteColor := range p {
		colors = append(colors, paletteColor.RGBA)
	}
	return colors
}

// Hex returns lowercase hex colors from the palette preserving order.
func (p Palette) Hex() []string {
	hexColors := make([]string, 0, len(p))
	for _, paletteColor := range p {
		hexColors = append(hexColors, paletteColor.HexString())
	}
	return hexColors
}

// Dominant returns the most frequent color from the palette.
func (p Palette) Dominant() Color {
	if len(p) == 0 {
		return Color{}
	}

	dominant := p[0]
	for _, paletteColor := range p[1:] {
		if paletteColor.Count > dominant.Count {
			dominant = paletteColor
		}
	}
	return dominant
}

// SortByFrequency sorts the palette in place by descending Count.
func (p Palette) SortByFrequency() {
	sort.Slice(p, func(i, j int) bool {
		if p[i].Count == p[j].Count {
			return p[i].Int() < p[j].Int()
		}
		return p[i].Count > p[j].Count
	})
}

// SortByBrightness sorts the palette in place by descending brightness.
func (p Palette) SortByBrightness() {
	sort.Slice(p, func(i, j int) bool {
		leftBrightness := Brightness(p[i].RGBA)
		rightBrightness := Brightness(p[j].RGBA)
		if leftBrightness == rightBrightness {
			return p[i].Int() < p[j].Int()
		}
		return leftBrightness > rightBrightness
	})
}

// SortBySaturation sorts the palette in place by descending saturation.
func (p Palette) SortBySaturation() {
	sort.Slice(p, func(i, j int) bool {
		leftSaturation := Saturation(p[i].RGBA)
		rightSaturation := Saturation(p[j].RGBA)
		if leftSaturation == rightSaturation {
			return p[i].Int() < p[j].Int()
		}
		return leftSaturation > rightSaturation
	})
}

// MergeSimilar merges colors within threshold RGB distance and returns a new palette.
func (p Palette) MergeSimilar(threshold float64) Palette {
	if len(p) == 0 {
		return nil
	}

	if threshold < 0 {
		threshold = 0
	}
	if threshold > 1 {
		threshold = 1
	}

	merged := make(Palette, 0, len(p))
	for _, currentColor := range p {
		mergedIndex := -1
		for i, mergedColor := range merged {
			if Distance(currentColor.RGBA, mergedColor.RGBA) <= threshold {
				mergedIndex = i
				break
			}
		}

		if mergedIndex < 0 {
			merged = append(merged, currentColor)
			continue
		}

		merged[mergedIndex] = mergeColors(merged[mergedIndex], currentColor)
	}

	totalCount := 0
	for _, paletteColor := range merged {
		totalCount += paletteColor.Count
	}
	if totalCount > 0 {
		for i := range merged {
			merged[i].Ratio = float64(merged[i].Count) / float64(totalCount)
		}
	}

	return merged
}

// Nearest returns the nearest palette color to provided color in RGB space.
func (p Palette) Nearest(c color.Color) Color {
	if len(p) == 0 {
		return Color{}
	}

	target := ToRGBA(c)
	nearestColor := p[0]
	minDistance := Distance(nearestColor.RGBA, target)
	for _, paletteColor := range p[1:] {
		distance := Distance(paletteColor.RGBA, target)
		if distance < minDistance {
			minDistance = distance
			nearestColor = paletteColor
		}
	}

	return nearestColor
}

// mergeColors combines two colors using count-weighted average of their channels.
func mergeColors(left, right Color) Color {
	leftWeight := left.Count
	rightWeight := right.Count
	totalWeight := leftWeight + rightWeight

	if totalWeight <= 0 {
		totalWeight = 2
		leftWeight = 1
		rightWeight = 1
	}

	return Color{
		RGBA: color.RGBA{
			R: clampIntToUint8((int(left.RGBA.R)*leftWeight + int(right.RGBA.R)*rightWeight) / totalWeight),
			G: clampIntToUint8((int(left.RGBA.G)*leftWeight + int(right.RGBA.G)*rightWeight) / totalWeight),
			B: clampIntToUint8((int(left.RGBA.B)*leftWeight + int(right.RGBA.B)*rightWeight) / totalWeight),
			A: clampIntToUint8((int(left.RGBA.A)*leftWeight + int(right.RGBA.A)*rightWeight) / totalWeight),
		},
		Count: left.Count + right.Count,
		Ratio: left.Ratio + right.Ratio,
	}
}

// clampIntToUint8 clamps an integer to the [0, 255] range.
func clampIntToUint8(value int) uint8 {
	if value < 0 {
		return 0
	}
	if value > 255 {
		return 255
	}
	return uint8(value)
}
