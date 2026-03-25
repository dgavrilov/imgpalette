package imgpalette

import (
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
)

const (
	defaultRenderSwatchWidth  = 64
	defaultRenderSwatchHeight = 64
	demoRenderSwatchWidth     = 96
	demoRenderSwatchHeight    = 96
	demoRenderGap             = 12
	demoRenderRadius          = 16
)

type renderConfig struct {
	swatchWidth  int
	swatchHeight int
	padding      int
	background   color.NRGBA
}

func defaultRenderConfig() renderConfig {
	return renderConfig{
		swatchWidth:  defaultRenderSwatchWidth,
		swatchHeight: defaultRenderSwatchHeight,
		padding:      0,
		background:   color.NRGBA{},
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

// RenderSwatchHeight sets height of each color swatch in pixels.
func RenderSwatchHeight(height int) RenderOption {
	return func(cfg *renderConfig) {
		cfg.swatchHeight = height
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

// RenderPalette renders a left-to-right palette preview image.
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
	if cfg.swatchHeight <= 0 {
		cfg.swatchHeight = defaultRenderSwatchHeight
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

	out := image.NewNRGBA(image.Rect(0, 0, width, cfg.swatchHeight))
	fillNRGBA(out, cfg.background)

	x := 0
	for _, paletteColor := range p {
		swatch := color.NRGBA{
			R: paletteColor.RGBA.R,
			G: paletteColor.RGBA.G,
			B: paletteColor.RGBA.B,
			A: paletteColor.RGBA.A,
		}
		fillRectNRGBA(out, x, 0, cfg.swatchWidth, cfg.swatchHeight, swatch)
		x += cfg.swatchWidth + cfg.padding
	}

	return out
}

// RenderFileDemo saves a demo palette PNG with transparent background and rounded swatches.
func RenderFileDemo(path string, p Palette) error {
	// #nosec G304 -- API intentionally accepts caller-provided output file paths.
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()

	return renderDemoPNG(file, p)
}

func fillNRGBA(img *image.NRGBA, c color.NRGBA) {
	for y := img.Rect.Min.Y; y < img.Rect.Max.Y; y++ {
		for x := img.Rect.Min.X; x < img.Rect.Max.X; x++ {
			img.SetNRGBA(x, y, c)
		}
	}
}

func renderDemoPalette(p Palette) image.Image {
	if len(p) == 0 {
		return image.NewNRGBA(image.Rect(0, 0, 1, 1))
	}

	width := len(p)*demoRenderSwatchWidth + (len(p)+1)*demoRenderGap
	height := demoRenderSwatchHeight + 2*demoRenderGap
	out := image.NewNRGBA(image.Rect(0, 0, width, height))

	x := demoRenderGap
	for _, paletteColor := range p {
		fillRoundedRectNRGBA(
			out,
			x,
			demoRenderGap,
			demoRenderSwatchWidth,
			demoRenderSwatchHeight,
			demoRenderRadius,
			color.NRGBA{
				R: paletteColor.RGBA.R,
				G: paletteColor.RGBA.G,
				B: paletteColor.RGBA.B,
				A: paletteColor.RGBA.A,
			},
		)
		x += demoRenderSwatchWidth + demoRenderGap
	}

	return out
}

// renderDemoPNG encodes the demo palette preview as PNG.
func renderDemoPNG(w io.Writer, p Palette) error {
	if err := png.Encode(w, renderDemoPalette(p)); err != nil {
		return err
	}
	return nil
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

// fillRoundedRectNRGBA fills a rounded rectangle clipped to img bounds.
func fillRoundedRectNRGBA(img *image.NRGBA, startX, startY, width, height, radius int, c color.NRGBA) {
	if radius < 0 {
		radius = 0
	}
	maxRadius := width / 2
	if height/2 < maxRadius {
		maxRadius = height / 2
	}
	if radius > maxRadius {
		radius = maxRadius
	}

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

	radiusSquared := radius * radius
	for y := startY; y < endY; y++ {
		for x := startX; x < endX; x++ {
			if radius == 0 || pointInsideRoundedRect(x, y, startX, startY, endX, endY, radius, radiusSquared) {
				img.SetNRGBA(x, y, c)
			}
		}
	}
}

// pointInsideRoundedRect reports whether a pixel lies inside the rounded shape.
func pointInsideRoundedRect(x, y, startX, startY, endX, endY, radius, radiusSquared int) bool {
	if x >= startX+radius && x < endX-radius {
		return true
	}
	if y >= startY+radius && y < endY-radius {
		return true
	}

	cornerX := startX + radius
	if x >= endX-radius {
		cornerX = endX - radius - 1
	}
	cornerY := startY + radius
	if y >= endY-radius {
		cornerY = endY - radius - 1
	}

	dx := x - cornerX
	dy := y - cornerY
	return dx*dx+dy*dy <= radiusSquared
}
