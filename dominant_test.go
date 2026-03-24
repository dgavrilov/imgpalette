package imgpalette

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/png"
	"os"
	"testing"
)

func TestDominantFunctions(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 10, 1))
	for x := 0; x < 10; x++ {
		if x < 7 {
			img.SetNRGBA(x, 0, color.NRGBA{R: 220, G: 20, B: 20, A: 255})
		} else {
			img.SetNRGBA(x, 0, color.NRGBA{R: 20, G: 20, B: 220, A: 255})
		}
	}

	expectedBin := (Color{RGBA: color.RGBA{R: 220, G: 20, B: 20, A: 255}}).QuantizedBinIndex()

	d1, err := Dominant(img, Count(2))
	if err != nil {
		t.Fatalf("Dominant returned error: %v", err)
	}
	if d1.QuantizedBinIndex() != expectedBin {
		t.Fatalf("expected red-like dominant color, got %+v", d1)
	}

	var encodedPNG bytes.Buffer
	if err := png.Encode(&encodedPNG, img); err != nil {
		t.Fatalf("png encode: %v", err)
	}
	d2, err := DominantReader(bytes.NewReader(encodedPNG.Bytes()), Count(2))
	if err != nil {
		t.Fatalf("DominantReader returned error: %v", err)
	}
	if d2.QuantizedBinIndex() != expectedBin {
		t.Fatalf("expected red-like dominant color from reader, got %+v", d2)
	}

	tempImageFile, err := os.CreateTemp(t.TempDir(), "dominant-*.png")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	imagePath := tempImageFile.Name()
	if err := png.Encode(tempImageFile, img); err != nil {
		_ = tempImageFile.Close()
		t.Fatalf("png encode: %v", err)
	}
	if err := tempImageFile.Close(); err != nil {
		t.Fatalf("close temp file: %v", err)
	}

	d3, err := DominantFile(imagePath, Count(2))
	if err != nil {
		t.Fatalf("DominantFile returned error: %v", err)
	}
	if d3.QuantizedBinIndex() != expectedBin {
		t.Fatalf("expected red-like dominant color from file, got %+v", d3)
	}
}

// Dominant/DominantReader/DominantFile: fully transparent image → ErrNoColors.
func TestDominantErrNoColors(t *testing.T) {
	transparentImg := image.NewNRGBA(image.Rect(0, 0, 2, 2))

	if _, err := Dominant(transparentImg); !errors.Is(err, ErrNoColors) {
		t.Fatalf("expected ErrNoColors from Dominant, got %v", err)
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, transparentImg); err != nil {
		t.Fatalf("png encode: %v", err)
	}
	if _, err := DominantReader(bytes.NewReader(buf.Bytes())); !errors.Is(err, ErrNoColors) {
		t.Fatalf("expected ErrNoColors from DominantReader, got %v", err)
	}

	f, err := os.CreateTemp(t.TempDir(), "dominant-nocolor-*.png")
	if err != nil {
		t.Fatalf("create temp: %v", err)
	}
	if err := png.Encode(f, transparentImg); err != nil {
		_ = f.Close()
		t.Fatalf("encode: %v", err)
	}
	_ = f.Close()
	if _, err := DominantFile(f.Name()); !errors.Is(err, ErrNoColors) {
		t.Fatalf("expected ErrNoColors from DominantFile, got %v", err)
	}
}

func TestDominantErrorPaths(t *testing.T) {
	if _, err := Dominant(nil); !errors.Is(err, ErrNilImage) {
		t.Fatalf("expected ErrNilImage from Dominant(nil), got %v", err)
	}

	if _, err := DominantReader(bytes.NewReader([]byte("not-an-image"))); !errors.Is(err, ErrDecodeImage) {
		t.Fatalf("expected ErrDecodeImage from DominantReader, got %v", err)
	}

	if _, err := DominantFile("/definitely/missing/file.png"); !errors.Is(err, ErrOpenImage) {
		t.Fatalf("expected ErrOpenImage from DominantFile, got %v", err)
	}
}

func TestExtractPaletteFromImage(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 10, 10))

	redColor := color.NRGBA{R: 240, G: 20, B: 20, A: 255}
	blueColor := color.NRGBA{R: 20, G: 30, B: 240, A: 255}

	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			if x < 7 {
				img.SetNRGBA(x, y, redColor)
			} else {
				img.SetNRGBA(x, y, blueColor)
			}
		}
	}

	dominantColors := extractPaletteFromImage(img, config{count: 2, sampleStride: 1, resizeTo: 256, colorSpace: SpaceOKLab})
	if len(dominantColors) != 2 {
		t.Fatalf("expected 2 colors, got %d", len(dominantColors))
	}

	if dominantColors[0].QuantizedBinIndex() != (Color{RGBA: color.RGBA{R: redColor.R, G: redColor.G, B: redColor.B, A: 255}}).QuantizedBinIndex() {
		t.Fatalf("expected first dominant color to be red-like, got %+v", dominantColors[0])
	}
	if dominantColors[1].QuantizedBinIndex() != (Color{RGBA: color.RGBA{R: blueColor.R, G: blueColor.G, B: blueColor.B, A: 255}}).QuantizedBinIndex() {
		t.Fatalf("expected second dominant color to be blue-like, got %+v", dominantColors[1])
	}
}

func TestExtractDominantColorsSkipsFullyTransparentPixels(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 4, 1))
	img.SetNRGBA(0, 0, color.NRGBA{R: 250, G: 0, B: 0, A: 255})
	img.SetNRGBA(1, 0, color.NRGBA{R: 250, G: 0, B: 0, A: 255})
	img.SetNRGBA(2, 0, color.NRGBA{R: 0, G: 250, B: 0, A: 0})
	img.SetNRGBA(3, 0, color.NRGBA{R: 0, G: 250, B: 0, A: 0})

	dominantColors := extractPaletteFromImage(img, config{count: 2, sampleStride: 1, resizeTo: 256, colorSpace: SpaceOKLab})
	if len(dominantColors) != 1 {
		t.Fatalf("expected 1 dominant color, got %d", len(dominantColors))
	}

	expectedRedBinIndex := (Color{RGBA: color.RGBA{R: 250, G: 0, B: 0, A: 255}}).QuantizedBinIndex()
	if dominantColors[0].QuantizedBinIndex() != expectedRedBinIndex {
		t.Fatalf("expected red bin, got %+v", dominantColors[0])
	}
}

func TestExtractPaletteFromReader(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	expectedColor := color.NRGBA{R: 40, G: 210, B: 80, A: 255}
	for y := 0; y < 2; y++ {
		for x := 0; x < 2; x++ {
			img.SetNRGBA(x, y, expectedColor)
		}
	}

	var encodedPNG bytes.Buffer
	if err := png.Encode(&encodedPNG, img); err != nil {
		t.Fatalf("png encode: %v", err)
	}

	dominantColors, err := ExtractReader(bytes.NewReader(encodedPNG.Bytes()), Count(1))
	if err != nil {
		t.Fatalf("extractPaletteFromReader returned error: %v", err)
	}
	if len(dominantColors) != 1 {
		t.Fatalf("expected 1 color, got %d", len(dominantColors))
	}
	if dominantColors[0].QuantizedBinIndex() != (Color{RGBA: color.RGBA{R: expectedColor.R, G: expectedColor.G, B: expectedColor.B, A: 255}}).QuantizedBinIndex() {
		t.Fatalf("expected green-like color, got %+v", dominantColors[0])
	}
}

func TestExtractPaletteFromReaderInvalidMaxColors(t *testing.T) {
	_, err := ExtractReader(bytes.NewReader([]byte("not-an-image")), Count(0))
	if err == nil {
		t.Fatal("expected ErrInvalidCount, got nil")
	}
	if !errors.Is(err, ErrInvalidCount) {
		t.Fatalf("expected ErrInvalidCount, got %v", err)
	}
}

func TestExtractPaletteFromReaderDecodeError(t *testing.T) {
	_, err := ExtractReader(bytes.NewReader([]byte("not-an-image")), Count(1))
	if err == nil {
		t.Fatal("expected decode error, got nil")
	}
}

func TestExtractPaletteFromPathInvalidMaxColors(t *testing.T) {
	_, err := ExtractFile("does-not-matter.png", Count(0))
	if err == nil {
		t.Fatal("expected ErrInvalidCount, got nil")
	}
	if !errors.Is(err, ErrInvalidCount) {
		t.Fatalf("expected ErrInvalidCount, got %v", err)
	}
}

func TestExtractPaletteFromPathOpenError(t *testing.T) {
	_, err := ExtractFile("/definitely/missing/file.png", Count(1))
	if err == nil {
		t.Fatal("expected open error, got nil")
	}
}

func TestExtractPaletteFromPath(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 3, 3))
	expectedColor := color.NRGBA{R: 210, G: 60, B: 30, A: 255}
	for y := 0; y < 3; y++ {
		for x := 0; x < 3; x++ {
			img.SetNRGBA(x, y, expectedColor)
		}
	}

	tempImageFile, err := os.CreateTemp(t.TempDir(), "palette-*.png")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	imagePath := tempImageFile.Name()
	if err := png.Encode(tempImageFile, img); err != nil {
		_ = tempImageFile.Close()
		t.Fatalf("png encode: %v", err)
	}
	if err := tempImageFile.Close(); err != nil {
		t.Fatalf("close temp file: %v", err)
	}

	dominantColors, err := ExtractFile(imagePath, Count(1))
	if err != nil {
		t.Fatalf("extractPaletteFromPath returned error: %v", err)
	}
	if len(dominantColors) != 1 {
		t.Fatalf("expected 1 color, got %d", len(dominantColors))
	}
	if dominantColors[0].QuantizedBinIndex() != (Color{RGBA: color.RGBA{R: expectedColor.R, G: expectedColor.G, B: expectedColor.B, A: 255}}).QuantizedBinIndex() {
		t.Fatalf("expected dominant bin for %+v, got %+v", expectedColor, dominantColors[0])
	}
}

func TestExtractPaletteFromImageNilAndInvalidMaxColors(t *testing.T) {
	if got := extractPaletteFromImage(nil, config{count: 1, sampleStride: 1, resizeTo: 256, colorSpace: SpaceOKLab}); got != nil {
		t.Fatalf("expected nil for nil image, got %v", got)
	}

	img := image.NewNRGBA(image.Rect(0, 0, 1, 1))
	if got := extractPaletteFromImage(img, config{count: 0, sampleStride: 1, resizeTo: 256, colorSpace: SpaceOKLab}); got != nil {
		t.Fatalf("expected nil for maxColors <= 0, got %v", got)
	}
}

func TestExtractPaletteFromImageTransparentOnly(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	img.SetNRGBA(0, 0, color.NRGBA{A: 0})
	img.SetNRGBA(1, 0, color.NRGBA{A: 0})
	img.SetNRGBA(0, 1, color.NRGBA{A: 0})
	img.SetNRGBA(1, 1, color.NRGBA{A: 0})

	got := extractPaletteFromImage(img, config{count: 3, sampleStride: 1, resizeTo: 256, colorSpace: SpaceOKLab})
	if got != nil {
		t.Fatalf("expected nil for transparent-only image, got %v", got)
	}
}

func TestExtractPaletteFromImageMaxColorsCap(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 2, 1))
	img.SetNRGBA(0, 0, color.NRGBA{R: 255, A: 255})
	img.SetNRGBA(1, 0, color.NRGBA{G: 255, A: 255})

	got := extractPaletteFromImage(img, config{count: 10, sampleStride: 1, resizeTo: 256, colorSpace: SpaceOKLab})
	if len(got) != 2 {
		t.Fatalf("expected 2 colors, got %d", len(got))
	}
}

func TestExtractPaletteFromImageRGBAWithPartialAlpha(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.SetRGBA(0, 0, color.RGBA{R: 100, G: 50, B: 25, A: 128})

	got := extractPaletteFromImage(img, config{count: 1, sampleStride: 1, resizeTo: 256, colorSpace: SpaceOKLab})
	if len(got) != 1 {
		t.Fatalf("expected 1 color, got %d", len(got))
	}
}

func TestExtractPaletteFromImageRGBAOpaqueAndTransparent(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 3, 1))
	img.SetRGBA(0, 0, color.RGBA{R: 10, G: 20, B: 30, A: 255})
	img.SetRGBA(1, 0, color.RGBA{R: 99, G: 88, B: 77, A: 0})
	img.SetRGBA(2, 0, color.RGBA{R: 10, G: 20, B: 30, A: 255})

	got := extractPaletteFromImage(img, config{count: 1, sampleStride: 1, resizeTo: 256, colorSpace: SpaceOKLab})
	if len(got) != 1 {
		t.Fatalf("expected 1 color, got %d", len(got))
	}
}

func TestExtractPaletteFromImageGray(t *testing.T) {
	img := image.NewGray(image.Rect(0, 0, 2, 1))
	img.SetGray(0, 0, color.Gray{Y: 40})
	img.SetGray(1, 0, color.Gray{Y: 40})

	got := extractPaletteFromImage(img, config{count: 1, sampleStride: 1, resizeTo: 256, colorSpace: SpaceOKLab})
	if len(got) != 1 {
		t.Fatalf("expected 1 color, got %d", len(got))
	}
	if got[0].RGBA.R != 40 || got[0].RGBA.G != 40 || got[0].RGBA.B != 40 {
		t.Fatalf("expected gray color [40,40,40], got %+v", got[0])
	}
}

func TestExtractPaletteFromImagePaletted(t *testing.T) {
	palette := color.Palette{
		color.NRGBA{R: 255, G: 0, B: 0, A: 255},
		color.RGBA{R: 0, G: 255, B: 0, A: 128},
		color.Gray{Y: 64},
	}
	img := image.NewPaletted(image.Rect(0, 0, 3, 1), palette)
	img.SetColorIndex(0, 0, 0)
	img.SetColorIndex(1, 0, 1)
	img.SetColorIndex(2, 0, 2)

	got := extractPaletteFromImage(img, config{count: 3, sampleStride: 1, resizeTo: 256, colorSpace: SpaceOKLab})
	if len(got) != 3 {
		t.Fatalf("expected 3 colors, got %d", len(got))
	}
}

func TestExtractPaletteFromImagePalettedTransparentBranches(t *testing.T) {
	palette := color.Palette{
		color.NRGBA{R: 255, G: 0, B: 0, A: 0},
		color.RGBA{R: 0, G: 255, B: 0, A: 0},
		color.Gray{Y: 10},
		color.NRGBA{R: 5, G: 6, B: 7, A: 255},
		color.RGBA{R: 20, G: 40, B: 60, A: 255},
		color.Alpha{A: 0},
	}
	img := image.NewPaletted(image.Rect(0, 0, 6, 1), palette)
	img.SetColorIndex(0, 0, 0)
	img.SetColorIndex(1, 0, 1)
	img.SetColorIndex(2, 0, 2)
	img.SetColorIndex(3, 0, 3)
	img.SetColorIndex(4, 0, 4)
	img.SetColorIndex(5, 0, 5)

	got := extractPaletteFromImage(img, config{count: 3, sampleStride: 1, resizeTo: 256, colorSpace: SpaceOKLab})
	if len(got) != 3 {
		t.Fatalf("expected 3 colors after skipping transparent palette entries, got %d", len(got))
	}
}

func TestExtractPaletteFromImageGenericImagePath(t *testing.T) {
	img := dominantGenericImage{
		bounds: image.Rect(0, 0, 2, 2),
		pixel:  color.NRGBA{R: 90, G: 80, B: 70, A: 255},
	}

	got := extractPaletteFromImage(img, config{count: 1, sampleStride: 1, resizeTo: 256, colorSpace: SpaceOKLab})
	if len(got) != 1 {
		t.Fatalf("expected 1 color, got %d", len(got))
	}
}

func TestExtractPaletteFromImageGenericImageTransparentPixel(t *testing.T) {
	img := dominantGenericImageWithTransparentPixel{bounds: image.Rect(0, 0, 2, 1)}
	got := extractPaletteFromImage(img, config{count: 2, sampleStride: 1, resizeTo: 256, colorSpace: SpaceOKLab})
	if len(got) != 1 {
		t.Fatalf("expected 1 color after skipping transparent generic pixel, got %d", len(got))
	}
}

func BenchmarkExtractPaletteFromImage(b *testing.B) {
	img := image.NewNRGBA(image.Rect(0, 0, 1920, 1080))

	testPalette := []color.NRGBA{
		{R: 220, G: 40, B: 40, A: 255},
		{R: 40, G: 220, B: 120, A: 255},
		{R: 40, G: 120, B: 220, A: 255},
		{R: 240, G: 200, B: 60, A: 255},
	}

	for y := 0; y < 1080; y++ {
		for x := 0; x < 1920; x++ {
			img.SetNRGBA(x, y, testPalette[(x/480)%len(testPalette)])
		}
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = extractPaletteFromImage(img, config{count: 5, sampleStride: 2, resizeTo: 256, colorSpace: SpaceOKLab})
	}
}

type dominantGenericImage struct {
	bounds image.Rectangle
	pixel  color.NRGBA
}

func (g dominantGenericImage) ColorModel() color.Model { return color.NRGBAModel }
func (g dominantGenericImage) Bounds() image.Rectangle { return g.bounds }
func (g dominantGenericImage) At(_, _ int) color.Color { return g.pixel }

type dominantGenericImageWithTransparentPixel struct{ bounds image.Rectangle }

func (g dominantGenericImageWithTransparentPixel) ColorModel() color.Model { return color.NRGBAModel }
func (g dominantGenericImageWithTransparentPixel) Bounds() image.Rectangle { return g.bounds }
func (g dominantGenericImageWithTransparentPixel) At(x, _ int) color.Color {
	if x == g.bounds.Min.X {
		return color.NRGBA{R: 0, G: 0, B: 0, A: 0}
	}
	return color.NRGBA{R: 90, G: 80, B: 70, A: 255}
}
