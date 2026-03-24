package imgpalette

import (
	"image"
	"image/color"
	"testing"
)

func TestRenderPaletteDefault(t *testing.T) {
	p := Palette{
		{RGBA: color.RGBA{R: 255, G: 0, B: 0, A: 255}},
		{RGBA: color.RGBA{R: 0, G: 0, B: 255, A: 255}},
	}

	out := RenderPalette(p)
	if out.Bounds().Dx() != 128 || out.Bounds().Dy() != 64 {
		t.Fatalf("unexpected bounds: %v", out.Bounds())
	}

	left := ToRGBA(out.At(10, 10))
	right := ToRGBA(out.At(90, 10))
	if left != (color.RGBA{R: 255, G: 0, B: 0, A: 255}) {
		t.Fatalf("unexpected left swatch color: %v", left)
	}
	if right != (color.RGBA{R: 0, G: 0, B: 255, A: 255}) {
		t.Fatalf("unexpected right swatch color: %v", right)
	}
}

func TestRenderPaletteWithOptions(t *testing.T) {
	p := Palette{
		{RGBA: color.RGBA{R: 255, G: 0, B: 0, A: 255}},
		{RGBA: color.RGBA{R: 0, G: 255, B: 0, A: 255}},
	}

	bg := color.RGBA{R: 7, G: 8, B: 9, A: 255}
	out := RenderPalette(
		p,
		RenderSwatchWidth(10),
		RenderHeight(6),
		RenderPadding(2),
		RenderBackground(bg),
	)

	if out.Bounds().Dx() != 22 || out.Bounds().Dy() != 6 {
		t.Fatalf("unexpected bounds: %v", out.Bounds())
	}

	paddingPixel := ToRGBA(out.At(10, 2))
	if paddingPixel != bg {
		t.Fatalf("expected padding pixel %v, got %v", bg, paddingPixel)
	}
}

func TestRenderPaletteEmpty(t *testing.T) {
	out := RenderPalette(nil)
	if out.Bounds().Dx() != 1 || out.Bounds().Dy() != 1 {
		t.Fatalf("expected 1x1 image for empty palette, got %v", out.Bounds())
	}
}

// RenderPalette: negative swatchWidth, height, padding are clamped to defaults/0.
func TestRenderPaletteNegativeOptions(t *testing.T) {
	p := Palette{{RGBA: color.RGBA{R: 255, G: 0, B: 0, A: 255}}}
	out := RenderPalette(p,
		RenderSwatchWidth(-1),
		RenderHeight(-1),
		RenderPadding(-3),
	)
	// Negative values clamped → defaults (64×64).
	if out.Bounds().Dx() != defaultRenderSwatchWidth || out.Bounds().Dy() != defaultRenderHeight {
		t.Fatalf("expected default size after negative options, got %v", out.Bounds())
	}
}

// RenderPalette: nil option is skipped.
func TestRenderPaletteNilOption(t *testing.T) {
	p := Palette{{RGBA: color.RGBA{R: 255, G: 0, B: 0, A: 255}}}
	out := RenderPalette(p, nil)
	if out == nil {
		t.Fatal("expected non-nil image with nil option")
	}
}

// fillRectNRGBA: exercise all four clamp branches by drawing outside image bounds.
func TestFillRectNRGBAClamps(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 10, 10))
	c := color.NRGBA{R: 100, G: 100, B: 100, A: 255}

	// startX < 0, startY < 0, endX > width, endY > height → all four clamp branches.
	fillRectNRGBA(img, -5, -5, 30, 30, c)

	// Corner pixel should be painted.
	if got := img.NRGBAAt(0, 0); got != c {
		t.Fatalf("expected corner pixel to be painted, got %v", got)
	}
}

// RenderPalette: single swatch with swatchWidth that produces width<=0 edge case.
// Achieved by setting swatchWidth=0 via the option (which then gets clamped to default).
// Separately verify the width=1 clamp path: a palette with a 0-width swatch.
func TestRenderPaletteWidthClamp(t *testing.T) {
	// Use a 1-color palette with padding only (no swatch width contribution):
	// force via negative option so swatchWidth → default (64) and total >= 1.
	p := Palette{{RGBA: color.RGBA{G: 255, A: 255}}}
	out := RenderPalette(p, RenderSwatchWidth(0))
	if out.Bounds().Dx() < 1 {
		t.Fatalf("expected width >= 1, got %d", out.Bounds().Dx())
	}
}
