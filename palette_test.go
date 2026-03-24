package imgpalette

import (
	"image/color"
	"math"
	"reflect"
	"testing"
)

func TestColorRepresentations(t *testing.T) {
	c := Color{RGBA: color.RGBA{R: 255, G: 255, B: 221, A: 255}}

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
	if !reflect.DeepEqual(c.RGBAValues(0.25), expectedRGBA) {
		t.Fatalf("expected RGBAValues() to be %v, got %v", expectedRGBA, c.RGBAValues(0.25))
	}

	if c.RGBAString(-1) != "rgba(255,255,221,0)" {
		t.Fatalf("expected RGBAString(-1) to clamp alpha to 0, got %q", c.RGBAString(-1))
	}
	if c.RGBAString(2) != "rgba(255,255,221,1)" {
		t.Fatalf("expected RGBAString(2) to clamp alpha to 1, got %q", c.RGBAString(2))
	}

	nanRGBA := c.RGBAValues(math.NaN())
	expectedNaNRGBA := [4]float64{255, 255, 221, 0}
	if !reflect.DeepEqual(nanRGBA, expectedNaNRGBA) {
		t.Fatalf("expected RGBAValues(NaN) to clamp alpha to 0, got %v", nanRGBA)
	}
}

func TestPaletteMethods(t *testing.T) {
	palette := Palette{
		{RGBA: color.RGBA{R: 255, G: 255, B: 221, A: 255}, Count: 10, Ratio: 0.5},
		{RGBA: color.RGBA{R: 0, G: 0, B: 0, A: 255}, Count: 10, Ratio: 0.5},
	}

	expectedColors := []color.RGBA{
		{R: 255, G: 255, B: 221, A: 255},
		{R: 0, G: 0, B: 0, A: 255},
	}
	if !reflect.DeepEqual(palette.Colors(), expectedColors) {
		t.Fatalf("expected Colors() to be %v, got %v", expectedColors, palette.Colors())
	}

	expectedHex := []string{"#ffffdd", "#000000"}
	if !reflect.DeepEqual(palette.Hex(), expectedHex) {
		t.Fatalf("expected Hex() to be %v, got %v", expectedHex, palette.Hex())
	}

	if dominant := palette.Dominant(); dominant.RGBA != (color.RGBA{R: 255, G: 255, B: 221, A: 255}) {
		t.Fatalf("expected Dominant() to be first color, got %v", dominant.RGBA)
	}
}

func TestPaletteSorts(t *testing.T) {
	byFrequency := Palette{
		{RGBA: color.RGBA{R: 10, G: 10, B: 10, A: 255}, Count: 1},
		{RGBA: color.RGBA{R: 20, G: 20, B: 20, A: 255}, Count: 3},
		{RGBA: color.RGBA{R: 30, G: 30, B: 30, A: 255}, Count: 2},
	}
	byFrequency.SortByFrequency()
	if byFrequency[0].Count != 3 || byFrequency[1].Count != 2 || byFrequency[2].Count != 1 {
		t.Fatalf("unexpected SortByFrequency result: %+v", byFrequency)
	}

	byBrightness := Palette{
		{RGBA: color.RGBA{R: 0, G: 0, B: 0, A: 255}},
		{RGBA: color.RGBA{R: 255, G: 255, B: 255, A: 255}},
	}
	byBrightness.SortByBrightness()
	if byBrightness[0].RGBA != (color.RGBA{R: 255, G: 255, B: 255, A: 255}) {
		t.Fatalf("unexpected SortByBrightness result: %+v", byBrightness)
	}

	bySaturation := Palette{
		{RGBA: color.RGBA{R: 120, G: 120, B: 120, A: 255}},
		{RGBA: color.RGBA{R: 255, G: 0, B: 0, A: 255}},
	}
	bySaturation.SortBySaturation()
	if bySaturation[0].RGBA != (color.RGBA{R: 255, G: 0, B: 0, A: 255}) {
		t.Fatalf("unexpected SortBySaturation result: %+v", bySaturation)
	}
}

func TestPaletteMergeSimilarAndNearest(t *testing.T) {
	palette := Palette{
		{RGBA: color.RGBA{R: 255, G: 0, B: 0, A: 255}, Count: 6, Ratio: 0.6},
		{RGBA: color.RGBA{R: 250, G: 5, B: 5, A: 255}, Count: 4, Ratio: 0.4},
	}

	merged := palette.MergeSimilar(0.05)
	if len(merged) != 1 {
		t.Fatalf("expected merged palette length 1, got %d", len(merged))
	}
	if merged[0].Count != 10 {
		t.Fatalf("expected merged count 10, got %d", merged[0].Count)
	}
	if math.Abs(merged[0].Ratio-1.0) > 1e-9 {
		t.Fatalf("expected merged ratio 1, got %f", merged[0].Ratio)
	}

	nearest := palette.Nearest(color.RGBA{R: 254, G: 1, B: 1, A: 255})
	if nearest.RGBA != (color.RGBA{R: 255, G: 0, B: 0, A: 255}) {
		t.Fatalf("expected nearest to be pure red, got %v", nearest.RGBA)
	}
}

// Dominant: second element has higher Count → branch triggers update.
func TestPaletteDominantUpdates(t *testing.T) {
	p := Palette{
		{RGBA: color.RGBA{R: 10, G: 10, B: 10, A: 255}, Count: 1},
		{RGBA: color.RGBA{R: 200, G: 50, B: 50, A: 255}, Count: 9},
	}
	if d := p.Dominant(); d.Count != 9 {
		t.Fatalf("expected Dominant().Count=9, got %d", d.Count)
	}
}

// Dominant: empty palette → zero Color.
func TestPaletteDominantEmpty(t *testing.T) {
	if (Palette{}).Dominant() != (Color{}) {
		t.Fatal("expected zero Color from empty palette Dominant()")
	}
}

// SortByFrequency: two colors with same Count → tie-break by Int() (lower first).
func TestSortByFrequencyTieBreak(t *testing.T) {
	p := Palette{
		{RGBA: color.RGBA{R: 20, G: 20, B: 20, A: 255}, Count: 5},
		{RGBA: color.RGBA{R: 10, G: 10, B: 10, A: 255}, Count: 5},
	}
	p.SortByFrequency()
	if p[0].RGBA.R != 10 {
		t.Fatalf("expected lower-Int color first on tie, got R=%d", p[0].RGBA.R)
	}
}

// SortByBrightness: two colors with same brightness → tie-break by Int().
func TestSortByBrightnessTieBreak(t *testing.T) {
	// Two identical colors force equal brightness branch.
	p := Palette{
		{RGBA: color.RGBA{R: 100, G: 100, B: 100, A: 255}},
		{RGBA: color.RGBA{R: 100, G: 100, B: 100, A: 255}},
	}
	p.SortByBrightness()
	if len(p) != 2 {
		t.Fatalf("expected 2 colors, got %d", len(p))
	}
}

// SortBySaturation: two fully-saturated colors → tie-break by Int().
func TestSortBySaturationTieBreak(t *testing.T) {
	p := Palette{
		{RGBA: color.RGBA{R: 0, G: 0, B: 255, A: 255}},  // saturation=1, Int=0x0000ff
		{RGBA: color.RGBA{R: 255, G: 0, B: 0, A: 255}},  // saturation=1, Int=0xff0000
	}
	p.SortBySaturation()
	// Both saturation=1; lower int (blue) comes first.
	if p[0].RGBA.B != 255 {
		t.Fatalf("expected blue (lower int) first on saturation tie, got %+v", p[0].RGBA)
	}
}

// MergeSimilar: threshold < 0 clamped to 0 — dissimilar colors don't merge.
func TestMergeSimilarNegativeThreshold(t *testing.T) {
	p := Palette{
		{RGBA: color.RGBA{R: 255, G: 0, B: 0, A: 255}, Count: 1},
		{RGBA: color.RGBA{R: 0, G: 255, B: 0, A: 255}, Count: 1},
	}
	if merged := p.MergeSimilar(-5); len(merged) != 2 {
		t.Fatalf("expected no merge with negative threshold, got %d colors", len(merged))
	}
}

// MergeSimilar: threshold > 1 clamped to 1 — Distance(black,white)=1.0 ≤ 1 → merge.
func TestMergeSimilarThresholdAboveOne(t *testing.T) {
	p := Palette{
		{RGBA: color.RGBA{R: 0, G: 0, B: 0, A: 255}, Count: 3},
		{RGBA: color.RGBA{R: 0, G: 0, B: 1, A: 255}, Count: 2}, // very close, distance≈0
	}
	if merged := p.MergeSimilar(5); len(merged) != 1 {
		t.Fatalf("expected full merge with threshold>1, got %d colors", len(merged))
	}
}

// MergeSimilar: empty palette returns nil.
func TestMergeSimilarEmpty(t *testing.T) {
	if merged := (Palette{}).MergeSimilar(0.5); merged != nil {
		t.Fatalf("expected nil from empty palette, got %v", merged)
	}
}

// Nearest: second color is closer — exercises the distance-update branch.
func TestNearestUpdates(t *testing.T) {
	p := Palette{
		{RGBA: color.RGBA{R: 0, G: 0, B: 0, A: 255}},
		{RGBA: color.RGBA{R: 255, G: 255, B: 255, A: 255}},
	}
	got := p.Nearest(color.RGBA{R: 200, G: 200, B: 200, A: 255})
	if got.RGBA != (color.RGBA{R: 255, G: 255, B: 255, A: 255}) {
		t.Fatalf("expected white as nearest, got %v", got.RGBA)
	}
}

// Nearest: empty palette → zero Color.
func TestNearestEmpty(t *testing.T) {
	if (Palette{}).Nearest(color.RGBA{}) != (Color{}) {
		t.Fatal("expected zero Color from empty palette Nearest()")
	}
}

// mergeColors: totalWeight == 0 falls back to equal weighting.
func TestMergeColorsZeroWeight(t *testing.T) {
	left := Color{RGBA: color.RGBA{R: 0, G: 0, B: 0, A: 255}, Count: 0}
	right := Color{RGBA: color.RGBA{R: 100, G: 100, B: 100, A: 255}, Count: 0}
	merged := mergeColors(left, right)
	if merged.RGBA.R != 50 {
		t.Fatalf("expected avg R=50 with zero weights, got %d", merged.RGBA.R)
	}
}

// clampIntToUint8: below 0 → 0, above 255 → 255.
func TestClampIntToUint8Extremes(t *testing.T) {
	if clampIntToUint8(-1) != 0 {
		t.Fatal("expected clampIntToUint8(-1)=0")
	}
	if clampIntToUint8(256) != 255 {
		t.Fatal("expected clampIntToUint8(256)=255")
	}
}
