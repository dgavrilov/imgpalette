package imgpalette

import (
	"image/color"
	"math"
	"testing"
)

func TestToRGBAAndHex(t *testing.T) {
	input := color.NRGBA{R: 12, G: 34, B: 56, A: 78}
	got := ToRGBA(input)
	if got != (color.RGBA{R: 12, G: 34, B: 56, A: 78}) {
		t.Fatalf("expected %v, got %v", color.RGBA{R: 12, G: 34, B: 56, A: 78}, got)
	}
	if Hex(input) != "#0c2238" {
		t.Fatalf("expected #0c2238, got %q", Hex(input))
	}
}

func TestBrightness(t *testing.T) {
	white := Brightness(color.RGBA{R: 255, G: 255, B: 255, A: 255})
	black := Brightness(color.RGBA{R: 0, G: 0, B: 0, A: 255})
	if math.Abs(white-1.0) > 1e-9 {
		t.Fatalf("expected brightness(white)=1, got %f", white)
	}
	if math.Abs(black-0.0) > 1e-9 {
		t.Fatalf("expected brightness(black)=0, got %f", black)
	}
}

func TestSaturationAndIsGray(t *testing.T) {
	gray := color.RGBA{R: 120, G: 120, B: 120, A: 255}
	red := color.RGBA{R: 255, G: 0, B: 0, A: 255}

	if Saturation(gray) != 0 {
		t.Fatalf("expected gray saturation to be 0, got %f", Saturation(gray))
	}
	if Saturation(red) != 1 {
		t.Fatalf("expected red saturation to be 1, got %f", Saturation(red))
	}
	if !IsGray(gray, 0.01) {
		t.Fatal("expected gray to be considered gray")
	}
	if IsGray(red, 0.2) {
		t.Fatal("expected red to not be considered gray")
	}
}

// Saturation: maxValue == 0 branch (pure black → 0).
func TestSaturationBlack(t *testing.T) {
	if Saturation(color.RGBA{R: 0, G: 0, B: 0, A: 255}) != 0 {
		t.Fatal("expected saturation of black to be 0")
	}
}

// IsGray: threshold clamping branches.
func TestIsGrayClamps(t *testing.T) {
	red := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	// threshold < 0 → clamped to 0 → Saturation(red)=1 > 0 → not gray
	if IsGray(red, -1) {
		t.Fatal("expected IsGray with negative threshold to return false for red")
	}
	// threshold > 1 → clamped to 1 → Saturation(red)=1 <= 1 → gray
	if !IsGray(red, 2) {
		t.Fatal("expected IsGray with threshold>1 to return true (clamped to 1)")
	}
}

// maxFloat: cover the branch where the third arg is the largest.
func TestMaxFloatThirdWins(t *testing.T) {
	if maxFloat(1, 2, 3) != 3 {
		t.Fatal("expected maxFloat(1,2,3)=3")
	}
}

// minFloat: cover the branch where the third arg is the smallest.
func TestMinFloatThirdWins(t *testing.T) {
	if minFloat(3, 2, 1) != 1 {
		t.Fatal("expected minFloat(3,2,1)=1")
	}
}
