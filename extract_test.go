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

func TestExtractUsesDefaults(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 2, 1))
	img.SetNRGBA(0, 0, color.NRGBA{R: 255, A: 255})
	img.SetNRGBA(1, 0, color.NRGBA{G: 255, A: 255})

	got, err := Extract(img)
	if err != nil {
		t.Fatalf("Extract returned error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 colors with default maxColors, got %d", len(got))
	}
}

func TestExtractNilImage(t *testing.T) {
	_, err := Extract(nil)
	if err == nil {
		t.Fatal("expected ErrNilImage, got nil")
	}
	if !errors.Is(err, ErrNilImage) {
		t.Fatalf("expected ErrNilImage, got %v", err)
	}
}

func TestExtractInvalidCountOption(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 1, 1))

	_, err := Extract(img, Count(0))
	if err == nil {
		t.Fatal("expected ErrInvalidCount, got nil")
	}
	if !errors.Is(err, ErrInvalidCount) {
		t.Fatalf("expected ErrInvalidCount, got %v", err)
	}
}

func TestExtractNoColors(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	_, err := Extract(img)
	if !errors.Is(err, ErrNoColors) {
		t.Fatalf("expected ErrNoColors, got %v", err)
	}
}

func TestExtractReaderAndExtractFile(t *testing.T) {
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

	fromReader, err := ExtractReader(bytes.NewReader(encodedPNG.Bytes()), Count(1), Space(SpaceRGB))
	if err != nil {
		t.Fatalf("ExtractReader returned error: %v", err)
	}
	if len(fromReader) != 1 {
		t.Fatalf("expected 1 color from reader, got %d", len(fromReader))
	}

	tempImageFile, err := os.CreateTemp(t.TempDir(), "palette-new-api-*.png")
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

	fromFile, err := ExtractFile(imagePath, Count(1), Space(SpaceRGB))
	if err != nil {
		t.Fatalf("ExtractFile returned error: %v", err)
	}
	if len(fromFile) != 1 {
		t.Fatalf("expected 1 color from file, got %d", len(fromFile))
	}
}

func TestExtractRespectsColorSpaceOption(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 4, 1))
	img.SetNRGBA(0, 0, color.NRGBA{R: 255, G: 0, B: 0, A: 255})
	img.SetNRGBA(1, 0, color.NRGBA{R: 250, G: 10, B: 10, A: 255})
	img.SetNRGBA(2, 0, color.NRGBA{R: 0, G: 255, B: 0, A: 255})
	img.SetNRGBA(3, 0, color.NRGBA{R: 10, G: 250, B: 10, A: 255})

	for _, space := range []ColorSpace{SpaceRGB, SpaceLab, SpaceOKLab} {
		palette, err := Extract(img, Count(2), Space(space))
		if err != nil {
			t.Fatalf("Extract returned error for space %v: %v", space, err)
		}
		if len(palette) != 2 {
			t.Fatalf("expected 2 colors for space %v, got %d", space, len(palette))
		}
	}
}

func TestExtractMergesNearDuplicateDominantBins(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 20, 10))

	for y := 0; y < 10; y++ {
		for x := 0; x < 20; x++ {
			switch {
			case x < 8:
				img.SetNRGBA(x, y, color.NRGBA{R: 3, G: 3, B: 3, A: 255})
			case x < 14:
				img.SetNRGBA(x, y, color.NRGBA{R: 8, G: 8, B: 7, A: 255})
			case x < 17:
				img.SetNRGBA(x, y, color.NRGBA{R: 12, G: 15, B: 10, A: 255})
			default:
				img.SetNRGBA(x, y, color.NRGBA{R: 90, G: 160, B: 50, A: 255})
			}
		}
	}

	palette, err := Extract(img, Count(3))
	if err != nil {
		t.Fatalf("Extract returned error: %v", err)
	}
	if len(palette) != 2 {
		t.Fatalf("expected merged palette of 2 colors, got %d", len(palette))
	}

	if palette[0].Count != 170 {
		t.Fatalf("expected merged dark color count 170, got %d", palette[0].Count)
	}
	if palette[1].RGBA.G <= palette[1].RGBA.R || palette[1].RGBA.G <= palette[1].RGBA.B {
		t.Fatalf("expected second palette color to remain green-like, got %+v", palette[1].RGBA)
	}
}

func TestExtractReaderDecodeErrorType(t *testing.T) {
	_, err := ExtractReader(bytes.NewReader([]byte("not-an-image")))
	if err == nil {
		t.Fatal("expected decode error, got nil")
	}
	if !errors.Is(err, ErrDecodeImage) {
		t.Fatalf("expected ErrDecodeImage, got %v", err)
	}
}

func TestExtractReaderNoColors(t *testing.T) {
	transparent := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	var buf bytes.Buffer
	if err := png.Encode(&buf, transparent); err != nil {
		t.Fatalf("png encode: %v", err)
	}

	_, err := ExtractReader(bytes.NewReader(buf.Bytes()))
	if !errors.Is(err, ErrNoColors) {
		t.Fatalf("expected ErrNoColors, got %v", err)
	}
}

func TestExtractFileOpenErrorType(t *testing.T) {
	_, err := ExtractFile("/definitely/missing/file.png")
	if err == nil {
		t.Fatal("expected open error, got nil")
	}
	if !errors.Is(err, ErrOpenImage) {
		t.Fatalf("expected ErrOpenImage, got %v", err)
	}
}

func TestExtractFileNoColors(t *testing.T) {
	transparent := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	f, err := os.CreateTemp(t.TempDir(), "extract-nocolor-*.png")
	if err != nil {
		t.Fatalf("create temp: %v", err)
	}
	if err := png.Encode(f, transparent); err != nil {
		_ = f.Close()
		t.Fatalf("encode: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	_, err = ExtractFile(f.Name())
	if !errors.Is(err, ErrNoColors) {
		t.Fatalf("expected ErrNoColors, got %v", err)
	}
}

func TestAccentNilImage(t *testing.T) {
	_, err := Accent(nil)
	if err == nil {
		t.Fatal("expected ErrNilImage, got nil")
	}
	if !errors.Is(err, ErrNilImage) {
		t.Fatalf("expected ErrNilImage, got %v", err)
	}
}

func TestAccentTransparentImage(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	_, err := Accent(img)
	if err == nil {
		t.Fatal("expected ErrNoColors, got nil")
	}
	if !errors.Is(err, ErrNoColors) {
		t.Fatalf("expected ErrNoColors, got %v", err)
	}
}

func TestAccentInvalidCount(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 1, 1))
	_, err := Accent(img, Count(0))
	if !errors.Is(err, ErrInvalidCount) {
		t.Fatalf("expected ErrInvalidCount, got %v", err)
	}
}

func TestAccentPrefersSaturatedColor(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 10, 10))
	gray := color.NRGBA{R: 140, G: 140, B: 140, A: 255}
	red := color.NRGBA{R: 230, G: 20, B: 20, A: 255}

	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			if x < 7 {
				img.SetNRGBA(x, y, gray)
			} else {
				img.SetNRGBA(x, y, red)
			}
		}
	}

	accent, err := Accent(img, Count(2))
	if err != nil {
		t.Fatalf("Accent returned error: %v", err)
	}

	redBin := (Color{RGBA: color.RGBA{R: red.R, G: red.G, B: red.B, A: 255}}).QuantizedBinIndex()
	if accent.QuantizedBinIndex() != redBin {
		t.Fatalf("expected red-like accent color, got %+v", accent)
	}
}

func TestPickAccentThresholdSkips(t *testing.T) {
	palette := Palette{
		{RGBA: color.RGBA{R: 255, G: 0, B: 0, A: 255}, Count: 10, Ratio: 0.05},    // skipped by minCoverage
		{RGBA: color.RGBA{R: 120, G: 120, B: 120, A: 255}, Count: 9, Ratio: 0.95}, // skipped by filterGray
	}
	_, ok := pickAccent(palette, config{
		minSaturation: 0.1,
		minCoverage:   0.1,
		filterGray:    true,
	}, true)
	if ok {
		t.Fatal("expected no accent when all colors are filtered by thresholds")
	}
}

func TestAccentReaderSuccess(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 4, 4))
	saturated := color.NRGBA{R: 220, G: 20, B: 20, A: 255}
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			img.SetNRGBA(x, y, saturated)
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("png encode: %v", err)
	}

	accent, err := AccentReader(bytes.NewReader(buf.Bytes()), Count(2))
	if err != nil {
		t.Fatalf("AccentReader returned error: %v", err)
	}

	expectedBin := (Color{RGBA: color.RGBA{R: saturated.R, G: saturated.G, B: saturated.B, A: 255}}).QuantizedBinIndex()
	if accent.QuantizedBinIndex() != expectedBin {
		t.Fatalf("expected saturated accent color, got %+v", accent)
	}
}

func TestAccentReaderDecodeError(t *testing.T) {
	_, err := AccentReader(bytes.NewReader([]byte("not-an-image")))
	if err == nil {
		t.Fatal("expected decode error, got nil")
	}
	if !errors.Is(err, ErrDecodeImage) {
		t.Fatalf("expected ErrDecodeImage, got %v", err)
	}
}

func TestAccentReaderInvalidCount(t *testing.T) {
	_, err := AccentReader(bytes.NewReader([]byte("x")), Count(0))
	if err == nil {
		t.Fatal("expected ErrInvalidCount, got nil")
	}
	if !errors.Is(err, ErrInvalidCount) {
		t.Fatalf("expected ErrInvalidCount, got %v", err)
	}
}

func TestAccentFileSuccess(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 4, 4))
	saturated := color.NRGBA{R: 20, G: 200, B: 20, A: 255}
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			img.SetNRGBA(x, y, saturated)
		}
	}

	f, err := os.CreateTemp(t.TempDir(), "accent-*.png")
	if err != nil {
		t.Fatalf("create temp: %v", err)
	}
	if err := png.Encode(f, img); err != nil {
		_ = f.Close()
		t.Fatalf("png encode: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	accent, err := AccentFile(f.Name(), Count(2))
	if err != nil {
		t.Fatalf("AccentFile returned error: %v", err)
	}

	expectedBin := (Color{RGBA: color.RGBA{R: saturated.R, G: saturated.G, B: saturated.B, A: 255}}).QuantizedBinIndex()
	if accent.QuantizedBinIndex() != expectedBin {
		t.Fatalf("expected saturated accent color from file, got %+v", accent)
	}
}

func TestAccentFileOpenError(t *testing.T) {
	_, err := AccentFile("/definitely/missing/file.png")
	if err == nil {
		t.Fatal("expected open error, got nil")
	}
	if !errors.Is(err, ErrOpenImage) {
		t.Fatalf("expected ErrOpenImage, got %v", err)
	}
}

func TestAccentFileInvalidCount(t *testing.T) {
	_, err := AccentFile("does-not-matter.png", Count(0))
	if err == nil {
		t.Fatal("expected ErrInvalidCount, got nil")
	}
	if !errors.Is(err, ErrInvalidCount) {
		t.Fatalf("expected ErrInvalidCount, got %v", err)
	}
}

func TestFilterGrayOption(t *testing.T) {
	// 7 gray pixels, 3 saturated red pixels — gray is dominant by count.
	img := image.NewNRGBA(image.Rect(0, 0, 10, 1))
	gray := color.NRGBA{R: 140, G: 140, B: 140, A: 255}
	red := color.NRGBA{R: 230, G: 20, B: 20, A: 255}
	for x := 0; x < 10; x++ {
		if x < 7 {
			img.SetNRGBA(x, 0, gray)
		} else {
			img.SetNRGBA(x, 0, red)
		}
	}

	// With FilterGray enabled: accent should be the saturated red, not gray.
	accent, err := Accent(img, Count(2), FilterGray(true), MinSaturation(0.1))
	if err != nil {
		t.Fatalf("Accent with FilterGray returned error: %v", err)
	}
	redBin := (Color{RGBA: color.RGBA{R: red.R, G: red.G, B: red.B, A: 255}}).QuantizedBinIndex()
	if accent.QuantizedBinIndex() != redBin {
		t.Fatalf("expected red accent with FilterGray=true, got %+v", accent)
	}

	// With FilterGray disabled (default): should still return a color without error.
	_, err = Accent(img, Count(2), FilterGray(false))
	if err != nil {
		t.Fatalf("Accent with FilterGray(false) returned error: %v", err)
	}
}

// uint8Clamped: value > 255 → returns 255.
func TestUint8Clamped(t *testing.T) {
	if uint8Clamped(256) != 255 {
		t.Fatal("expected uint8Clamped(256)=255")
	}
	if uint8Clamped(0) != 0 {
		t.Fatal("expected uint8Clamped(0)=0")
	}
}

// Accent: all colors are gray + FilterGray → first pass fails → second pass succeeds.
func TestAccentFallbackToAllColors(t *testing.T) {
	// All gray — first pass (applyThresholds=true with FilterGray) skips all,
	// second pass (applyThresholds=false) returns something.
	img := image.NewNRGBA(image.Rect(0, 0, 4, 1))
	gray := color.NRGBA{R: 128, G: 128, B: 128, A: 255}
	for x := 0; x < 4; x++ {
		img.SetNRGBA(x, 0, gray)
	}
	// Use FilterGray + high MinSaturation so first pass finds nothing.
	_, err := Accent(img, FilterGray(true), MinSaturation(0.9))
	if err != nil {
		t.Fatalf("expected fallback to succeed for all-gray image, got: %v", err)
	}
}

// AccentReader ErrNoColors: transparent image produces empty palette.
func TestAccentReaderErrNoColors(t *testing.T) {
	transparent := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	var buf bytes.Buffer
	if err := png.Encode(&buf, transparent); err != nil {
		t.Fatalf("png encode: %v", err)
	}
	_, err := AccentReader(bytes.NewReader(buf.Bytes()))
	if !errors.Is(err, ErrNoColors) {
		t.Fatalf("expected ErrNoColors from AccentReader, got %v", err)
	}
}

func TestAccentReaderFallbackToAllColors(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 3, 1))
	gray := color.NRGBA{R: 140, G: 140, B: 140, A: 255}
	for x := 0; x < 3; x++ {
		img.SetNRGBA(x, 0, gray)
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("png encode: %v", err)
	}

	accent, err := AccentReader(bytes.NewReader(buf.Bytes()), FilterGray(true), MinSaturation(0.9))
	if err != nil {
		t.Fatalf("expected fallback to succeed in AccentReader, got %v", err)
	}
	if accent == (Color{}) {
		t.Fatal("expected non-zero accent color from fallback pass")
	}
}

// AccentFile ErrNoColors: transparent image produces empty palette.
func TestAccentFileErrNoColors(t *testing.T) {
	transparent := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	f, err := os.CreateTemp(t.TempDir(), "accent-nocolor-*.png")
	if err != nil {
		t.Fatalf("create temp: %v", err)
	}
	if err := png.Encode(f, transparent); err != nil {
		_ = f.Close()
		t.Fatalf("encode: %v", err)
	}
	_ = f.Close()
	_, err = AccentFile(f.Name())
	if !errors.Is(err, ErrNoColors) {
		t.Fatalf("expected ErrNoColors from AccentFile, got %v", err)
	}
}

func TestPickAccentGrayFilterBranch(t *testing.T) {
	palette := Palette{
		{RGBA: color.RGBA{R: 120, G: 120, B: 120, A: 255}, Count: 10, Ratio: 1},
	}

	_, ok := pickAccent(palette, config{
		minSaturation: -0.1,
		minCoverage:   0,
		filterGray:    true,
	}, true)
	if ok {
		t.Fatal("expected gray filter branch to reject gray color")
	}
}

func TestPickAccentTieBreakByCount(t *testing.T) {
	palette := Palette{
		{RGBA: color.RGBA{R: 255, G: 0, B: 0, A: 255}, Count: 2, Ratio: 0.5},
		{RGBA: color.RGBA{R: 255, G: 0, B: 0, A: 255}, Count: 10, Ratio: 0.5},
	}

	accent, ok := pickAccent(palette, config{}, true)
	if !ok {
		t.Fatal("expected accent to be found")
	}
	if accent.Count != 10 {
		t.Fatalf("expected tie-break to choose higher count, got %d", accent.Count)
	}
}
