package imgpalette

import (
	"image/color"
	"math"
	"testing"
)

func TestContrast(t *testing.T) {
	black := color.RGBA{R: 0, G: 0, B: 0, A: 255}
	white := color.RGBA{R: 255, G: 255, B: 255, A: 255}

	ratio := Contrast(black, white)
	if math.Abs(ratio-21.0) > 1e-9 {
		t.Fatalf("expected contrast(black, white)=21, got %f", ratio)
	}
	if math.Abs(Contrast(black, black)-1.0) > 1e-9 {
		t.Fatalf("expected contrast(black, black)=1, got %f", Contrast(black, black))
	}
}

func TestDistance(t *testing.T) {
	black := color.RGBA{R: 0, G: 0, B: 0, A: 255}
	white := color.RGBA{R: 255, G: 255, B: 255, A: 255}

	if math.Abs(Distance(black, black)-0.0) > 1e-9 {
		t.Fatalf("expected distance(black, black)=0, got %f", Distance(black, black))
	}
	if math.Abs(Distance(black, white)-1.0) > 1e-9 {
		t.Fatalf("expected distance(black, white)=1, got %f", Distance(black, white))
	}
}

func TestBestTextColor(t *testing.T) {
	if got := BestTextColor(color.RGBA{R: 255, G: 255, B: 255, A: 255}); got != (color.RGBA{R: 0, G: 0, B: 0, A: 255}) {
		t.Fatalf("expected black text on white bg, got %v", got)
	}
	if got := BestTextColor(color.RGBA{R: 0, G: 0, B: 0, A: 255}); got != (color.RGBA{R: 255, G: 255, B: 255, A: 255}) {
		t.Fatalf("expected white text on black bg, got %v", got)
	}
}
