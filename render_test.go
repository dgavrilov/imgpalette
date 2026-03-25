package imgpalette

import (
	"errors"
	"image"
	"image/color"
	"image/png"
	"os"
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
		RenderSwatchHeight(6),
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
		RenderSwatchHeight(-1),
		RenderPadding(-3),
	)
	// Negative values clamped → defaults (64×64).
	if out.Bounds().Dx() != defaultRenderSwatchWidth || out.Bounds().Dy() != defaultRenderSwatchHeight {
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

func TestRenderFileDemo(t *testing.T) {
	p := Palette{
		{RGBA: color.RGBA{R: 255, G: 0, B: 0, A: 255}},
		{RGBA: color.RGBA{R: 0, G: 255, B: 0, A: 255}},
	}

	path := t.TempDir() + "/demo.png"
	if err := RenderFileDemo(path, p); err != nil {
		t.Fatalf("RenderFileDemo returned error: %v", err)
	}

	// #nosec G304 -- test reads a temp file path created within the test.
	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("open demo png: %v", err)
	}
	defer func() {
		_ = file.Close()
	}()

	img, err := png.Decode(file)
	if err != nil {
		t.Fatalf("decode demo png: %v", err)
	}

	expectedWidth := len(p)*demoRenderSwatchWidth + (len(p)+1)*demoRenderGap
	expectedHeight := demoRenderSwatchHeight + 2*demoRenderGap
	if img.Bounds().Dx() != expectedWidth || img.Bounds().Dy() != expectedHeight {
		t.Fatalf("unexpected demo bounds: %v", img.Bounds())
	}

	if got := ToRGBA(img.At(0, 0)); got.A != 0 {
		t.Fatalf("expected transparent background at image corner, got %v", got)
	}
	if got := ToRGBA(img.At(demoRenderGap/2, expectedHeight/2)); got.A != 0 {
		t.Fatalf("expected transparent left margin, got %v", got)
	}
	if got := ToRGBA(img.At(demoRenderGap+demoRenderSwatchWidth+demoRenderGap/2, expectedHeight/2)); got.A != 0 {
		t.Fatalf("expected transparent gap between swatches, got %v", got)
	}
	if got := ToRGBA(img.At(expectedWidth-demoRenderGap/2-1, expectedHeight/2)); got.A != 0 {
		t.Fatalf("expected transparent right margin, got %v", got)
	}

	firstTileCenter := ToRGBA(img.At(demoRenderGap+demoRenderSwatchWidth/2, demoRenderGap+demoRenderSwatchHeight/2))
	if firstTileCenter != (color.RGBA{R: 255, G: 0, B: 0, A: 255}) {
		t.Fatalf("expected first swatch center to be red, got %v", firstTileCenter)
	}

	roundedCorner := ToRGBA(img.At(demoRenderGap, demoRenderGap))
	if roundedCorner.A != 0 {
		t.Fatalf("expected rounded tile corner to stay transparent, got %v", roundedCorner)
	}

	insideCorner := ToRGBA(img.At(demoRenderGap+demoRenderRadius, demoRenderGap+1))
	if insideCorner != (color.RGBA{R: 255, G: 0, B: 0, A: 255}) {
		t.Fatalf("expected tile interior near rounded corner to be painted, got %v", insideCorner)
	}
}

func TestRenderFileDemoCreateError(t *testing.T) {
	err := RenderFileDemo(t.TempDir(), Palette{{RGBA: color.RGBA{R: 255, A: 255}}})
	if err == nil {
		t.Fatal("expected error when trying to create demo png at directory path")
	}
}

func TestRenderDemoPaletteEmpty(t *testing.T) {
	out := renderDemoPalette(nil)
	if out.Bounds().Dx() != 1 || out.Bounds().Dy() != 1 {
		t.Fatalf("expected 1x1 image for empty demo palette, got %v", out.Bounds())
	}
	if got := ToRGBA(out.At(0, 0)); got.A != 0 {
		t.Fatalf("expected empty demo palette image to stay transparent, got %v", got)
	}
}

func TestRenderDemoPNGWriteError(t *testing.T) {
	expectedErr := errors.New("write failed")
	err := renderDemoPNG(failingWriter{err: expectedErr}, Palette{{RGBA: color.RGBA{R: 255, A: 255}}})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected write error %v, got %v", expectedErr, err)
	}
}

func TestFillRoundedRectNRGBAClampAndZeroRadius(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 6, 4))
	c := color.NRGBA{R: 1, G: 2, B: 3, A: 255}

	fillRoundedRectNRGBA(img, -2, -1, 10, 8, -5, c)

	if got := img.NRGBAAt(0, 0); got != c {
		t.Fatalf("expected negative radius to clamp to filled rectangle, got %v", got)
	}
	if got := img.NRGBAAt(5, 3); got != c {
		t.Fatalf("expected clamped rectangle to paint far corner, got %v", got)
	}
}

func TestFillRoundedRectNRGBARadiusClamp(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 10, 10))
	c := color.NRGBA{R: 20, G: 30, B: 40, A: 255}

	fillRoundedRectNRGBA(img, 2, 2, 6, 4, 99, c)

	if got := img.NRGBAAt(2, 2); got.A != 0 {
		t.Fatalf("expected oversized radius to keep rounded corner transparent, got %v", got)
	}
	if got := img.NRGBAAt(5, 3); got != c {
		t.Fatalf("expected interior pixel to be painted after radius clamp, got %v", got)
	}
}

type failingWriter struct {
	err error
}

func (w failingWriter) Write(_ []byte) (int, error) {
	return 0, w.err
}
