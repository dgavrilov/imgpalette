package imgpalette

import (
	"fmt"
	"image"
	"image/color"
	// Register GIF decoder for image.Decode.
	_ "image/gif"
	// Register JPEG decoder for image.Decode.
	_ "image/jpeg"
	// Register PNG decoder for image.Decode.
	_ "image/png"
	"io"
	"math"
	"os"
	"sort"

	iconvert "github.com/dgavrilov/imgpalette/internal/convert"
	isample "github.com/dgavrilov/imgpalette/internal/sample"
)

// Extract returns representative palette colors from an image.
func Extract(img image.Image, opts ...Option) (Palette, error) {
	if img == nil {
		return nil, ErrNilImage
	}

	cfg, err := resolveConfig(opts...)
	if err != nil {
		return nil, err
	}

	palette := extractWithConfig(img, cfg)
	if len(palette) == 0 {
		return nil, ErrNoColors
	}
	return palette, nil
}

// ExtractReader decodes an image from r and returns representative palette colors.
func ExtractReader(r io.Reader, opts ...Option) (Palette, error) {
	cfg, err := resolveConfig(opts...)
	if err != nil {
		return nil, err
	}

	palette, err := extractReaderWithConfig(r, cfg)
	if err != nil {
		return nil, err
	}
	if len(palette) == 0 {
		return nil, ErrNoColors
	}
	return palette, nil
}

// ExtractFile opens an image file and returns representative palette colors.
func ExtractFile(path string, opts ...Option) (Palette, error) {
	cfg, err := resolveConfig(opts...)
	if err != nil {
		return nil, err
	}

	palette, err := extractFileWithConfig(path, cfg)
	if err != nil {
		return nil, err
	}
	if len(palette) == 0 {
		return nil, ErrNoColors
	}
	return palette, nil
}

// extractPaletteFromImage bins sampled pixels and returns representative colors.
func extractPaletteFromImage(img image.Image, cfg config) Palette {
	if img == nil || cfg.count <= 0 {
		return nil
	}
	maxColors := cfg.count
	sampleStride := cfg.sampleStride
	if sampleStride <= 1 {
		sampleStride = 1
	}

	var binPixelCounts [totalQuantizedBins]uint32
	var binRedSums [totalQuantizedBins]uint64
	var binGreenSums [totalQuantizedBins]uint64
	var binBlueSums [totalQuantizedBins]uint64

	switch typedImage := img.(type) {
	case *image.NRGBA:
		accumulateFromNRGBA(typedImage, sampleStride, cfg.colorSpace, &binPixelCounts, &binRedSums, &binGreenSums, &binBlueSums)
	case *image.RGBA:
		accumulateFromRGBA(typedImage, sampleStride, cfg.colorSpace, &binPixelCounts, &binRedSums, &binGreenSums, &binBlueSums)
	case *image.Gray:
		accumulateFromGray(typedImage, sampleStride, cfg.colorSpace, &binPixelCounts, &binRedSums, &binGreenSums, &binBlueSums)
	case *image.Paletted:
		accumulateFromPaletted(typedImage, sampleStride, cfg.colorSpace, &binPixelCounts, &binRedSums, &binGreenSums, &binBlueSums)
	default:
		accumulateFromGenericImage(img, sampleStride, cfg.colorSpace, &binPixelCounts, &binRedSums, &binGreenSums, &binBlueSums)
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

	var totalSampledPixels uint64
	for _, binStats := range nonEmptyBins {
		totalSampledPixels += uint64(binStats.pixelCount)
	}

	return selectDominantColors(nonEmptyBins, maxColors, totalSampledPixels)
}

const (
	quantizationBits              = 5
	quantizationMask              = (1 << quantizationBits) - 1
	quantizedBinsPerAxis          = 1 << quantizationBits
	totalQuantizedBins            = quantizedBinsPerAxis * quantizedBinsPerAxis * quantizedBinsPerAxis
	defaultPaletteMergeSimilarity = 0.08
)

type quantizedBinStats struct {
	binIndex   int
	pixelCount uint32
	redSum     uint64
	greenSum   uint64
	blueSum    uint64
}

// selectDominantColors merges near-duplicate bins and truncates the result.
func selectDominantColors(nonEmptyBins []quantizedBinStats, maxColors int, totalSampledPixels uint64) Palette {
	dominantColors := make(Palette, 0, maxColors)
	for _, binStats := range nonEmptyBins {
		candidate := colorFromBinStats(binStats)

		// Merge near-duplicate bins before truncating so the default palette
		// contains fewer visually redundant colors.
		merged := false
		for i := range dominantColors {
			if Distance(candidate.RGBA, dominantColors[i].RGBA) > defaultPaletteMergeSimilarity {
				continue
			}
			dominantColors[i] = mergeColors(dominantColors[i], candidate)
			merged = true
			break
		}
		if merged {
			continue
		}
		if len(dominantColors) >= maxColors {
			continue
		}
		dominantColors = append(dominantColors, candidate)
	}

	sort.Slice(dominantColors, func(i, j int) bool {
		if dominantColors[i].Count == dominantColors[j].Count {
			return dominantColors[i].Int() < dominantColors[j].Int()
		}
		return dominantColors[i].Count > dominantColors[j].Count
	})

	if totalSampledPixels > 0 {
		for i := range dominantColors {
			dominantColors[i].Ratio = float64(dominantColors[i].Count) / float64(totalSampledPixels)
		}
	}

	return dominantColors
}

func colorFromBinStats(binStats quantizedBinStats) Color {
	pixelCount := uint64(binStats.pixelCount)
	return Color{
		RGBA: color.RGBA{
			R: uint8Clamped(binStats.redSum / pixelCount),
			G: uint8Clamped(binStats.greenSum / pixelCount),
			B: uint8Clamped(binStats.blueSum / pixelCount),
			A: 255,
		},
		Count: int(binStats.pixelCount),
	}
}

func addPixelToQuantizedBin(quantizeRed, quantizeGreen, quantizeBlue, sumRed, sumGreen, sumBlue uint8, binPixelCounts *[totalQuantizedBins]uint32, binRedSums, binGreenSums, binBlueSums *[totalQuantizedBins]uint64) {
	quantizedRed := int(quantizeRed >> (8 - quantizationBits))
	quantizedGreen := int(quantizeGreen >> (8 - quantizationBits))
	quantizedBlue := int(quantizeBlue >> (8 - quantizationBits))
	binIndex := (quantizedRed << (quantizationBits * 2)) | (quantizedGreen << quantizationBits) | quantizedBlue

	binPixelCounts[binIndex]++
	binRedSums[binIndex] += uint64(sumRed)
	binGreenSums[binIndex] += uint64(sumGreen)
	binBlueSums[binIndex] += uint64(sumBlue)
}

func accumulateFromGenericImage(img image.Image, sampleStride int, colorSpace ColorSpace, binPixelCounts *[totalQuantizedBins]uint32, binRedSums, binGreenSums, binBlueSums *[totalQuantizedBins]uint64) {
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y += sampleStride {
		for x := bounds.Min.X; x < bounds.Max.X; x += sampleStride {
			red16, green16, blue16, alpha16 := img.At(x, y).RGBA()
			if alpha16 == 0 {
				continue
			}
			red := iconvert.Uint8FromRGBA32(red16)
			green := iconvert.Uint8FromRGBA32(green16)
			blue := iconvert.Uint8FromRGBA32(blue16)
			qr, qg, qb := quantizationKey(red, green, blue, colorSpace)
			addPixelToQuantizedBin(qr, qg, qb, red, green, blue, binPixelCounts, binRedSums, binGreenSums, binBlueSums)
		}
	}
}

func accumulateFromNRGBA(img *image.NRGBA, sampleStride int, colorSpace ColorSpace, binPixelCounts *[totalQuantizedBins]uint32, binRedSums, binGreenSums, binBlueSums *[totalQuantizedBins]uint64) {
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y += sampleStride {
		rowOffset := (y-bounds.Min.Y)*img.Stride - bounds.Min.X*4
		for x := bounds.Min.X; x < bounds.Max.X; x += sampleStride {
			pixelOffset := rowOffset + x*4
			alpha := img.Pix[pixelOffset+3]
			if alpha == 0 {
				continue
			}
			red := img.Pix[pixelOffset]
			green := img.Pix[pixelOffset+1]
			blue := img.Pix[pixelOffset+2]
			qr, qg, qb := quantizationKey(red, green, blue, colorSpace)
			addPixelToQuantizedBin(qr, qg, qb, red, green, blue, binPixelCounts, binRedSums, binGreenSums, binBlueSums)
		}
	}
}

func accumulateFromRGBA(img *image.RGBA, sampleStride int, colorSpace ColorSpace, binPixelCounts *[totalQuantizedBins]uint32, binRedSums, binGreenSums, binBlueSums *[totalQuantizedBins]uint64) {
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
				red = iconvert.UnpremultiplyChannel(red, alpha)
				green = iconvert.UnpremultiplyChannel(green, alpha)
				blue = iconvert.UnpremultiplyChannel(blue, alpha)
			}
			qr, qg, qb := quantizationKey(red, green, blue, colorSpace)
			addPixelToQuantizedBin(qr, qg, qb, red, green, blue, binPixelCounts, binRedSums, binGreenSums, binBlueSums)
		}
	}
}

func accumulateFromGray(img *image.Gray, sampleStride int, colorSpace ColorSpace, binPixelCounts *[totalQuantizedBins]uint32, binRedSums, binGreenSums, binBlueSums *[totalQuantizedBins]uint64) {
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y += sampleStride {
		rowOffset := (y-bounds.Min.Y)*img.Stride - bounds.Min.X
		for x := bounds.Min.X; x < bounds.Max.X; x += sampleStride {
			pixelOffset := rowOffset + x
			gray := img.Pix[pixelOffset]
			qr, qg, qb := quantizationKey(gray, gray, gray, colorSpace)
			addPixelToQuantizedBin(qr, qg, qb, gray, gray, gray, binPixelCounts, binRedSums, binGreenSums, binBlueSums)
		}
	}
}

func accumulateFromPaletted(img *image.Paletted, sampleStride int, colorSpace ColorSpace, binPixelCounts *[totalQuantizedBins]uint32, binRedSums, binGreenSums, binBlueSums *[totalQuantizedBins]uint64) {
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
				qr, qg, qb := quantizationKey(typedColor.R, typedColor.G, typedColor.B, colorSpace)
				addPixelToQuantizedBin(qr, qg, qb, typedColor.R, typedColor.G, typedColor.B, binPixelCounts, binRedSums, binGreenSums, binBlueSums)
			case color.RGBA:
				if typedColor.A == 0 {
					continue
				}
				red, green, blue := typedColor.R, typedColor.G, typedColor.B
				if typedColor.A < 255 {
					red = iconvert.UnpremultiplyChannel(red, typedColor.A)
					green = iconvert.UnpremultiplyChannel(green, typedColor.A)
					blue = iconvert.UnpremultiplyChannel(blue, typedColor.A)
				}
				qr, qg, qb := quantizationKey(red, green, blue, colorSpace)
				addPixelToQuantizedBin(qr, qg, qb, red, green, blue, binPixelCounts, binRedSums, binGreenSums, binBlueSums)
			default:
				red16, green16, blue16, alpha16 := paletteColor.RGBA()
				if alpha16 == 0 {
					continue
				}
				red := iconvert.Uint8FromRGBA32(red16)
				green := iconvert.Uint8FromRGBA32(green16)
				blue := iconvert.Uint8FromRGBA32(blue16)
				qr, qg, qb := quantizationKey(red, green, blue, colorSpace)
				addPixelToQuantizedBin(qr, qg, qb, red, green, blue, binPixelCounts, binRedSums, binGreenSums, binBlueSums)
			}
		}
	}
}

func quantizationKey(red, green, blue uint8, colorSpace ColorSpace) (uint8, uint8, uint8) {
	switch colorSpace {
	case SpaceLab:
		lab := iconvert.RGBAToLab(color.RGBA{R: red, G: green, B: blue, A: 255})
		return labToQuantizationKey(lab)
	case SpaceOKLab:
		oklab := iconvert.RGBAToOKLab(color.RGBA{R: red, G: green, B: blue, A: 255})
		return oklabToQuantizationKey(oklab)
	default:
		return red, green, blue
	}
}

func labToQuantizationKey(lab iconvert.Lab) (uint8, uint8, uint8) {
	l := clampIntToUint8(int(math.Round((lab.L / 100.0) * 255.0)))
	a := clampIntToUint8(int(math.Round(lab.A + 128.0)))
	b := clampIntToUint8(int(math.Round(lab.B + 128.0)))
	return l, a, b
}

func oklabToQuantizationKey(oklab iconvert.OKLab) (uint8, uint8, uint8) {
	// A/B are typically in roughly [-0.4, 0.4]; map that range into [0..255].
	l := clampIntToUint8(int(math.Round(oklab.L * 255.0)))
	a := clampIntToUint8(int(math.Round((oklab.A + 0.4) * (255.0 / 0.8))))
	b := clampIntToUint8(int(math.Round((oklab.B + 0.4) * (255.0 / 0.8))))
	return l, a, b
}

func uint8Clamped(value uint64) uint8 {
	if value > 255 {
		return 255
	}
	return uint8(value)
}

func extractWithConfig(img image.Image, cfg config) Palette {
	resized := isample.ResizeToMaxSide(img, cfg.resizeTo)
	return extractPaletteFromImage(resized, cfg)
}

func extractReaderWithConfig(r io.Reader, cfg config) (Palette, error) {
	decodedImage, _, err := image.Decode(r)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDecodeImage, err)
	}
	return extractWithConfig(decodedImage, cfg), nil
}

func extractFileWithConfig(path string, cfg config) (Palette, error) {
	// #nosec G304 -- API intentionally accepts caller-provided file paths.
	imageFile, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOpenImage, err)
	}
	defer func() {
		_ = imageFile.Close()
	}()

	return extractReaderWithConfig(imageFile, cfg)
}
