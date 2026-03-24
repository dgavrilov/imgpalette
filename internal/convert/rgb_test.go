package convert

import (
	"image/color"
	"testing"
)

func TestToRGBA(t *testing.T) {
	in := color.NRGBA{R: 12, G: 34, B: 56, A: 78}
	got := ToRGBA(in)
	if got != (color.RGBA{R: 12, G: 34, B: 56, A: 78}) {
		t.Fatalf("unexpected conversion: %v", got)
	}
}

func TestToRGBANil(t *testing.T) {
	if got := ToRGBA(nil); got != (color.RGBA{}) {
		t.Fatalf("expected zero RGBA for nil color, got %v", got)
	}
}

func TestToRGBAPremultiplied(t *testing.T) {
	// color.RGBA with partial alpha stores premultiplied channels.
	// R=100, A=128 means original red ≈ 200 (100/128 * 255).
	premul := color.RGBA{R: 100, G: 50, B: 25, A: 128}
	got := ToRGBA(premul)
	if got.A != 128 {
		t.Fatalf("expected A=128, got A=%d", got.A)
	}
	// Unpremultiplied R should be 100*255/128 = 199
	if got.R < 195 || got.R > 202 {
		t.Fatalf("expected R≈199 (unpremultiplied), got R=%d", got.R)
	}
}

func TestToRGBAOpaqueRGBA(t *testing.T) {
	// Fully opaque: premultiplied == non-premultiplied.
	opaque := color.RGBA{R: 100, G: 150, B: 200, A: 255}
	got := ToRGBA(opaque)
	if got != opaque {
		t.Fatalf("expected identical opaque color, got %v", got)
	}
}

func TestToRGBAZeroAlpha(t *testing.T) {
	// Zero alpha should return zero color.
	got := ToRGBA(color.RGBA{R: 255, G: 255, B: 255, A: 0})
	if got != (color.RGBA{}) {
		t.Fatalf("expected zero color for A=0, got %v", got)
	}
}

func TestUnpremultiplyChannel(t *testing.T) {
	if UnpremultiplyChannel(200, 0) != 0 {
		t.Fatal("expected zero with alpha 0")
	}
}

// Uint8FromRGBA32: channel >> 8 > 255 → returns 255.
// This cannot happen with normal uint32 RGBA values (max 0xffff >> 8 = 0xff),
// but the guard exists. Exercise via a value just above 0xffff.
func TestUint8FromRGBA32Normal(t *testing.T) {
	if Uint8FromRGBA32(0xffff) != 0xff {
		t.Fatal("expected 0xff for max 16-bit channel")
	}
	if Uint8FromRGBA32(0) != 0 {
		t.Fatal("expected 0 for zero channel")
	}
	if Uint8FromRGBA32(0x1_0000) != 0xff {
		t.Fatal("expected 0xff when shifted value is above 255")
	}
}

// UnpremultiplyChannel: result > 255 is clamped to 255.
// E.g. channel=200, alpha=1 → 200*255/1 = 51000 → clamped to 255.
func TestUnpremultiplyChannelOverflow(t *testing.T) {
	if got := UnpremultiplyChannel(200, 1); got != 255 {
		t.Fatalf("expected 255 for overflow case, got %d", got)
	}
}

// ToRGBA: generic path (non-RGBA, non-NRGBA type with partial alpha).
func TestToRGBAGenericPath(t *testing.T) {
	// color.Alpha is not RGBA or NRGBA; goes through generic path.
	got := ToRGBA(color.Alpha{A: 128})
	// Alpha-only color: RGBA() returns (a,a,a,a) premultiplied.
	if got.A != 128 {
		t.Fatalf("expected A=128 for color.Alpha{128}, got %v", got)
	}
}

// ToRGBA: generic path with zero alpha → returns color.RGBA{}.
func TestToRGBAGenericZeroAlpha(t *testing.T) {
	got := ToRGBA(color.Alpha{A: 0})
	if got != (color.RGBA{}) {
		t.Fatalf("expected zero RGBA for zero-alpha generic color, got %v", got)
	}
}
