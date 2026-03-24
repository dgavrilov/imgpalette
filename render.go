package imgpalette

import (
	"image"
	"image/color"
)

const (
	defaultRenderSwatchWidth = 64
	defaultRenderHeight      = 64
)

type renderConfig struct {
	swatchWidth int
	height      int
	padding     int
	background  color.NRGBA
}

func defaultRenderConfig() renderConfig {
	return renderConfig{
		swatchWidth: defaultRenderSwatchWidth,
		height:      defaultRenderHeight,
		padding:     0,
		background:  color.NRGBA{},
	}
}

// RenderOption configures palette rendering.
type RenderOption func(*renderConfig)

// RenderSwatchWidth sets width of each color swatch in pixels.
func RenderSwatchWidth(width int) RenderOption {
	return func(cfg *renderConfig) {
		cfg.swatchWidth = width
	}
}

// RenderHeight sets preview height in pixels.
func RenderHeight(height int) RenderOption {
	return func(cfg *renderConfig) {
		cfg.height = height
	}
}

// RenderPadding sets padding between swatches in pixels.
func RenderPadding(padding int) RenderOption {
	return func(cfg *renderConfig) {
		cfg.padding = padding
	}
}

// RenderBackground sets background color used for preview and padding areas.
func RenderBackground(c color.Color) RenderOption {
	return func(cfg *renderConfig) {
		cfg.background = color.NRGBA(ToRGBA(c))
	}
}

// RenderPalette renders palette preview image with vertical swatches.
func RenderPalette(p Palette, opts ...RenderOption) image.Image {
	cfg := defaultRenderConfig()
	for _, option := range opts {
		if option == nil {
			continue
		}
		option(&cfg)
	}

	if cfg.swatchWidth <= 0 {
		cfg.swatchWidth = defaultRenderSwatchWidth
	}
	if cfg.height <= 0 {
		cfg.height = defaultRenderHeight
	}
	if cfg.padding < 0 {
		cfg.padding = 0
	}

	if len(p) == 0 {
		out := image.NewNRGBA(image.Rect(0, 0, 1, 1))
		out.SetNRGBA(0, 0, cfg.background)
		return out
	}

	width := len(p) * cfg.swatchWidth
	if len(p) > 1 {
		width += (len(p) - 1) * cfg.padding
	}

	out := image.NewNRGBA(image.Rect(0, 0, width, cfg.height))
	fillNRGBA(out, cfg.background)

	x := 0
	for _, paletteColor := range p {
		swatch := color.NRGBA{
			R: paletteColor.RGBA.R,
			G: paletteColor.RGBA.G,
			B: paletteColor.RGBA.B,
			A: paletteColor.RGBA.A,
		}
		fillRectNRGBA(out, x, 0, cfg.swatchWidth, cfg.height, swatch)
		x += cfg.swatchWidth + cfg.padding
	}

	return out
}

func fillNRGBA(img *image.NRGBA, c color.NRGBA) {
	for y := img.Rect.Min.Y; y < img.Rect.Max.Y; y++ {
		for x := img.Rect.Min.X; x < img.Rect.Max.X; x++ {
			img.SetNRGBA(x, y, c)
		}
	}
}

func fillRectNRGBA(img *image.NRGBA, startX, startY, width, height int, c color.NRGBA) {
	endX := startX + width
	endY := startY + height
	if startX < img.Rect.Min.X {
		startX = img.Rect.Min.X
	}
	if startY < img.Rect.Min.Y {
		startY = img.Rect.Min.Y
	}
	if endX > img.Rect.Max.X {
		endX = img.Rect.Max.X
	}
	if endY > img.Rect.Max.Y {
		endY = img.Rect.Max.Y
	}
	for y := startY; y < endY; y++ {
		for x := startX; x < endX; x++ {
			img.SetNRGBA(x, y, c)
		}
	}
}
