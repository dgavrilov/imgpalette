package convert

import (
	"image/color"
	"math"
	"testing"
)

func TestRGBAToLabWhite(t *testing.T) {
	lab := RGBAToLab(color.RGBA{R: 255, G: 255, B: 255, A: 255})
	if math.Abs(lab.L-100) > 1 {
		t.Fatalf("unexpected L for white: %f", lab.L)
	}
}

// labF: cover the linear branch (t <= delta^3 ≈ 0.00886).
// Black has XYZ ≈ (0,0,0), so all three labF calls take the linear branch.
func TestRGBAToLabBlack(t *testing.T) {
	lab := RGBAToLab(color.RGBA{R: 0, G: 0, B: 0, A: 255})
	if math.Abs(lab.L-0) > 0.1 {
		t.Fatalf("expected L≈0 for black, got %f", lab.L)
	}
}

// RGBAToLab: mid-gray exercises both branches of labF for different channels.
func TestRGBAToLabGray(t *testing.T) {
	lab := RGBAToLab(color.RGBA{R: 128, G: 128, B: 128, A: 255})
	// L should be roughly 53 for mid-gray in CIE Lab.
	if lab.L < 40 || lab.L > 65 {
		t.Fatalf("unexpected L for mid-gray: %f", lab.L)
	}
}
