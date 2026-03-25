package imgpalette

import (
	"fmt"
	"image"
	"io"
	"math"
	"os"
)

// Accent returns a saturated, visually noticeable color from an image.
func Accent(img image.Image, opts ...Option) (Color, error) {
	if img == nil {
		return Color{}, ErrNilImage
	}

	cfg, err := resolveConfig(opts...)
	if err != nil {
		return Color{}, err
	}

	palette := extractWithConfig(img, cfg)
	if len(palette) == 0 {
		return Color{}, ErrNoColors
	}

	if accent, ok := pickAccent(palette, cfg, true); ok {
		return accent, nil
	}
	accent, _ := pickAccent(palette, cfg, false)
	return accent, nil
}

// AccentReader decodes an image from r and returns an accent color.
func AccentReader(r io.Reader, opts ...Option) (Color, error) {
	cfg, err := resolveConfig(opts...)
	if err != nil {
		return Color{}, err
	}

	return accentReaderWithConfig(r, cfg)
}

// AccentFile opens an image file and returns an accent color.
func AccentFile(path string, opts ...Option) (Color, error) {
	cfg, err := resolveConfig(opts...)
	if err != nil {
		return Color{}, err
	}

	return accentFileWithConfig(path, cfg)
}

func accentReaderWithConfig(r io.Reader, cfg config) (Color, error) {
	palette, err := extractReaderWithConfig(r, cfg)
	if err != nil {
		return Color{}, err
	}
	if len(palette) == 0 {
		return Color{}, ErrNoColors
	}

	if accent, ok := pickAccent(palette, cfg, true); ok {
		return accent, nil
	}
	accent, _ := pickAccent(palette, cfg, false)
	return accent, nil
}

func accentFileWithConfig(path string, cfg config) (Color, error) {
	// #nosec G304 -- API intentionally accepts caller-provided file paths.
	imageFile, err := os.Open(path)
	if err != nil {
		return Color{}, fmt.Errorf("%w: %v", ErrOpenImage, err)
	}
	defer func() {
		_ = imageFile.Close()
	}()

	return accentReaderWithConfig(imageFile, cfg)
}

// pickAccent scores palette colors by saturation, coverage, and readability.
func pickAccent(palette Palette, cfg config, applyThresholds bool) (Color, bool) {
	bestIndex := -1
	bestScore := -1.0
	bestCount := 0

	for i, paletteColor := range palette {
		saturation := Saturation(paletteColor.RGBA)
		coverage := paletteColor.Ratio
		brightness := Brightness(paletteColor.RGBA)

		if applyThresholds {
			if saturation < cfg.minSaturation {
				continue
			}
			if coverage < cfg.minCoverage {
				continue
			}
			if cfg.filterGray && IsGray(paletteColor.RGBA, cfg.minSaturation) {
				continue
			}
		}

		// Keep accent readable by penalizing colors too close to black or white.
		contrastPotential := 1 - math.Abs(2*brightness-1)
		score := 0.55*saturation + 0.30*coverage + 0.15*contrastPotential

		if bestIndex < 0 {
			bestIndex = i
			bestScore = score
			bestCount = paletteColor.Count
			continue
		}

		if score > bestScore {
			bestIndex = i
			bestScore = score
			bestCount = paletteColor.Count
			continue
		}

		if score == bestScore && paletteColor.Count > bestCount {
			bestIndex = i
			bestScore = score
			bestCount = paletteColor.Count
		}
	}

	if bestIndex < 0 {
		return Color{}, false
	}
	return palette[bestIndex], true
}
