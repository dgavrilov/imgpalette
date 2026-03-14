package imgpalette

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"reflect"
	"testing"
)

func TestExtractDominantColorsFromImage(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 10, 10))

	redColor := color.NRGBA{R: 240, G: 20, B: 20, A: 255}
	blueColor := color.NRGBA{R: 20, G: 30, B: 240, A: 255}

	// 70% red, 30% blue.
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			if x < 7 {
				img.SetNRGBA(x, y, redColor)
			} else {
				img.SetNRGBA(x, y, blueColor)
			}
		}
	}

	dominantColors := ExtractDominantColorsFromImage(img, 2, 1)
	if len(dominantColors) != 2 {
		t.Fatalf("expected 2 colors, got %d", len(dominantColors))
	}

	if dominantColors[0].QuantizedBinIndex() != (Color{R: redColor.R, G: redColor.G, B: redColor.B}).QuantizedBinIndex() {
		t.Fatalf("expected first dominant color to be red-like, got %+v", dominantColors[0])
	}
	if dominantColors[1].QuantizedBinIndex() != (Color{R: blueColor.R, G: blueColor.G, B: blueColor.B}).QuantizedBinIndex() {
		t.Fatalf("expected second dominant color to be blue-like, got %+v", dominantColors[1])
	}
}

func TestExtractDominantColorsSkipsFullyTransparentPixels(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 4, 1))
	img.SetNRGBA(0, 0, color.NRGBA{R: 250, G: 0, B: 0, A: 255})
	img.SetNRGBA(1, 0, color.NRGBA{R: 250, G: 0, B: 0, A: 255})
	img.SetNRGBA(2, 0, color.NRGBA{R: 0, G: 250, B: 0, A: 0})
	img.SetNRGBA(3, 0, color.NRGBA{R: 0, G: 250, B: 0, A: 0})

	dominantColors := ExtractDominantColorsFromImage(img, 2, 1)
	if len(dominantColors) != 1 {
		t.Fatalf("expected 1 dominant color, got %d", len(dominantColors))
	}

	expectedRedBinIndex := (Color{R: 250, G: 0, B: 0}).QuantizedBinIndex()
	if dominantColors[0].QuantizedBinIndex() != expectedRedBinIndex {
		t.Fatalf("expected red bin, got %+v", dominantColors[0])
	}
}

func TestExtractDominantColorsFromReader(t *testing.T) {
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

	dominantColors, err := ExtractDominantColorsFromReader(bytes.NewReader(encodedPNG.Bytes()), 1, 1)
	if err != nil {
		t.Fatalf("ExtractDominantColorsFromReader returned error: %v", err)
	}
	if len(dominantColors) != 1 {
		t.Fatalf("expected 1 color, got %d", len(dominantColors))
	}
	if dominantColors[0].QuantizedBinIndex() != (Color{R: expectedColor.R, G: expectedColor.G, B: expectedColor.B}).QuantizedBinIndex() {
		t.Fatalf("expected green-like color, got %+v", dominantColors[0])
	}
}

func TestExtractDominantColorsFromReaderInvalidMaxColors(t *testing.T) {
	_, err := ExtractDominantColorsFromReader(bytes.NewReader([]byte("not-an-image")), 0, 1)
	if err == nil {
		t.Fatal("expected ErrInvalidMaxColors, got nil")
	}
	if !errors.Is(err, ErrInvalidMaxColors) {
		t.Fatalf("expected ErrInvalidMaxColors, got %v", err)
	}
}

func TestExtractDominantColorsFromReaderDecodeError(t *testing.T) {
	_, err := ExtractDominantColorsFromReader(bytes.NewReader([]byte("not-an-image")), 1, 1)
	if err == nil {
		t.Fatal("expected decode error, got nil")
	}
}

func TestExtractDominantColorsFromPathInvalidMaxColors(t *testing.T) {
	_, err := ExtractDominantColorsFromPath("does-not-matter.png", 0, 1)
	if err == nil {
		t.Fatal("expected ErrInvalidMaxColors, got nil")
	}
	if !errors.Is(err, ErrInvalidMaxColors) {
		t.Fatalf("expected ErrInvalidMaxColors, got %v", err)
	}
}

func TestExtractDominantColorsFromPathOpenError(t *testing.T) {
	_, err := ExtractDominantColorsFromPath("/definitely/missing/file.png", 1, 1)
	if err == nil {
		t.Fatal("expected open error, got nil")
	}
}

func TestExtractDominantColorsFromPath(t *testing.T) {
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

	dominantColors, err := ExtractDominantColorsFromPath(imagePath, 1, 1)
	if err != nil {
		t.Fatalf("ExtractDominantColorsFromPath returned error: %v", err)
	}
	if len(dominantColors) != 1 {
		t.Fatalf("expected 1 color, got %d", len(dominantColors))
	}
	if dominantColors[0].QuantizedBinIndex() != (Color{R: expectedColor.R, G: expectedColor.G, B: expectedColor.B}).QuantizedBinIndex() {
		t.Fatalf("expected dominant bin for %+v, got %+v", expectedColor, dominantColors[0])
	}
}

func TestExtractDominantColorsFromImageNilAndInvalidMaxColors(t *testing.T) {
	if got := ExtractDominantColorsFromImage(nil, 1, 1); got != nil {
		t.Fatalf("expected nil for nil image, got %v", got)
	}

	img := image.NewNRGBA(image.Rect(0, 0, 1, 1))
	if got := ExtractDominantColorsFromImage(img, 0, 1); got != nil {
		t.Fatalf("expected nil for maxColors <= 0, got %v", got)
	}
}

func TestExtractDominantColorsFromImageTransparentOnly(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	img.SetNRGBA(0, 0, color.NRGBA{A: 0})
	img.SetNRGBA(1, 0, color.NRGBA{A: 0})
	img.SetNRGBA(0, 1, color.NRGBA{A: 0})
	img.SetNRGBA(1, 1, color.NRGBA{A: 0})

	got := ExtractDominantColorsFromImage(img, 3, 1)
	if got != nil {
		t.Fatalf("expected nil for transparent-only image, got %v", got)
	}
}

func TestExtractDominantColorsFromImageMaxColorsCap(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 2, 1))
	img.SetNRGBA(0, 0, color.NRGBA{R: 255, A: 255})
	img.SetNRGBA(1, 0, color.NRGBA{G: 255, A: 255})

	got := ExtractDominantColorsFromImage(img, 10, 1)
	if len(got) != 2 {
		t.Fatalf("expected 2 colors, got %d", len(got))
	}
}

func TestExtractDominantColorsFromImageRGBAWithPartialAlpha(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.SetRGBA(0, 0, color.RGBA{R: 100, G: 50, B: 25, A: 128})

	got := ExtractDominantColorsFromImage(img, 1, 1)
	if len(got) != 1 {
		t.Fatalf("expected 1 color, got %d", len(got))
	}
}

func TestExtractDominantColorsFromImageRGBAOpaqueAndTransparent(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 3, 1))
	img.SetRGBA(0, 0, color.RGBA{R: 10, G: 20, B: 30, A: 255}) // opaque path
	img.SetRGBA(1, 0, color.RGBA{R: 99, G: 88, B: 77, A: 0})   // transparent skip
	img.SetRGBA(2, 0, color.RGBA{R: 10, G: 20, B: 30, A: 255}) // opaque path

	got := ExtractDominantColorsFromImage(img, 1, 1)
	if len(got) != 1 {
		t.Fatalf("expected 1 color, got %d", len(got))
	}
}

func TestExtractDominantColorsFromImageGray(t *testing.T) {
	img := image.NewGray(image.Rect(0, 0, 2, 1))
	img.SetGray(0, 0, color.Gray{Y: 40})
	img.SetGray(1, 0, color.Gray{Y: 40})

	got := ExtractDominantColorsFromImage(img, 1, 1)
	if len(got) != 1 {
		t.Fatalf("expected 1 color, got %d", len(got))
	}
	if got[0].R != 40 || got[0].G != 40 || got[0].B != 40 {
		t.Fatalf("expected gray color [40,40,40], got %+v", got[0])
	}
}

func TestExtractDominantColorsFromImagePaletted(t *testing.T) {
	palette := color.Palette{
		color.NRGBA{R: 255, G: 0, B: 0, A: 255},
		color.RGBA{R: 0, G: 255, B: 0, A: 128},
		color.Gray{Y: 64},
	}
	img := image.NewPaletted(image.Rect(0, 0, 3, 1), palette)
	img.SetColorIndex(0, 0, 0)
	img.SetColorIndex(1, 0, 1)
	img.SetColorIndex(2, 0, 2)

	got := ExtractDominantColorsFromImage(img, 3, 1)
	if len(got) != 3 {
		t.Fatalf("expected 3 colors, got %d", len(got))
	}
}

func TestExtractDominantColorsFromImagePalettedTransparentBranches(t *testing.T) {
	palette := color.Palette{
		color.NRGBA{R: 255, G: 0, B: 0, A: 0},   // NRGBA transparent skip
		color.RGBA{R: 0, G: 255, B: 0, A: 0},    // RGBA transparent skip
		color.Gray{Y: 10},                       // default opaque
		color.NRGBA{R: 5, G: 6, B: 7, A: 255},   // NRGBA opaque
		color.RGBA{R: 20, G: 40, B: 60, A: 255}, // RGBA opaque
		color.Alpha{A: 0},                       // default transparent skip
	}
	img := image.NewPaletted(image.Rect(0, 0, 6, 1), palette)
	img.SetColorIndex(0, 0, 0)
	img.SetColorIndex(1, 0, 1)
	img.SetColorIndex(2, 0, 2)
	img.SetColorIndex(3, 0, 3)
	img.SetColorIndex(4, 0, 4)
	img.SetColorIndex(5, 0, 5)

	got := ExtractDominantColorsFromImage(img, 3, 1)
	if len(got) != 3 {
		t.Fatalf("expected 3 colors after skipping transparent palette entries, got %d", len(got))
	}
}

func TestExtractDominantColorsFromImageGenericImagePath(t *testing.T) {
	img := genericImage{
		bounds: image.Rect(0, 0, 2, 2),
		pixel:  color.NRGBA{R: 90, G: 80, B: 70, A: 255},
	}

	got := ExtractDominantColorsFromImage(img, 1, 1)
	if len(got) != 1 {
		t.Fatalf("expected 1 color, got %d", len(got))
	}
}

func TestExtractDominantColorsFromImageGenericImageTransparentPixel(t *testing.T) {
	img := genericImageWithTransparentPixel{
		bounds: image.Rect(0, 0, 2, 1),
	}

	got := ExtractDominantColorsFromImage(img, 2, 1)
	if len(got) != 1 {
		t.Fatalf("expected 1 color after skipping transparent generic pixel, got %d", len(got))
	}
}

func BenchmarkExtractDominantColorsFromImage(b *testing.B) {
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
		_ = ExtractDominantColorsFromImage(img, 5, 2)
	}
}

type genericImage struct {
	bounds image.Rectangle
	pixel  color.NRGBA
}

func (g genericImage) ColorModel() color.Model {
	return color.NRGBAModel
}

func (g genericImage) Bounds() image.Rectangle {
	return g.bounds
}

func (g genericImage) At(_, _ int) color.Color {
	return g.pixel
}

type genericImageWithTransparentPixel struct {
	bounds image.Rectangle
}

func (g genericImageWithTransparentPixel) ColorModel() color.Model {
	return color.NRGBAModel
}

func (g genericImageWithTransparentPixel) Bounds() image.Rectangle {
	return g.bounds
}

func (g genericImageWithTransparentPixel) At(x, _ int) color.Color {
	if x == g.bounds.Min.X {
		return color.NRGBA{R: 0, G: 0, B: 0, A: 0}
	}
	return color.NRGBA{R: 90, G: 80, B: 70, A: 255}
}

func TestColorRepresentations(t *testing.T) {
	c := Color{R: 255, G: 255, B: 221}

	if c.String() != "#ffffdd" {
		t.Fatalf("expected String() to be #ffffdd, got %q", c.String())
	}
	if c.HexString() != "#ffffdd" {
		t.Fatalf("expected HexString() to be #ffffdd, got %q", c.HexString())
	}
	if c.RGBString() != "rgb(255,255,221)" {
		t.Fatalf("expected RGBString() to be rgb(255,255,221), got %q", c.RGBString())
	}
	if c.RGBAString(0.25) != "rgba(255,255,221,0.25)" {
		t.Fatalf("expected RGBAString() to be rgba(255,255,221,0.25), got %q", c.RGBAString(0.25))
	}
	if c.Int() != 0xffffdd {
		t.Fatalf("expected Int() to be 0xffffdd, got %#x", c.Int())
	}

	expectedRGB := [3]uint8{255, 255, 221}
	if !reflect.DeepEqual(c.RGB(), expectedRGB) {
		t.Fatalf("expected RGB() to be %v, got %v", expectedRGB, c.RGB())
	}

	expectedRGBA := [4]float64{255, 255, 221, 0.25}
	if !reflect.DeepEqual(c.RGBA(0.25), expectedRGBA) {
		t.Fatalf("expected RGBA() to be %v, got %v", expectedRGBA, c.RGBA(0.25))
	}

	if c.RGBAString(-1) != "rgba(255,255,221,0)" {
		t.Fatalf("expected RGBAString(-1) to clamp alpha to 0, got %q", c.RGBAString(-1))
	}
	if c.RGBAString(2) != "rgba(255,255,221,1)" {
		t.Fatalf("expected RGBAString(2) to clamp alpha to 1, got %q", c.RGBAString(2))
	}

	nanRGBA := c.RGBA(math.NaN())
	expectedNaNRGBA := [4]float64{255, 255, 221, 0}
	if !reflect.DeepEqual(nanRGBA, expectedNaNRGBA) {
		t.Fatalf("expected RGBA(NaN) to clamp alpha to 0, got %v", nanRGBA)
	}
}

func TestUnpremultiplyChannelAlphaZero(t *testing.T) {
	if got := unpremultiplyChannel(200, 0); got != 0 {
		t.Fatalf("expected unpremultiplyChannel(..., 0) to be 0, got %d", got)
	}
}
