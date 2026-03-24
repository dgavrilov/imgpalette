package imgpalette

import (
	"errors"
	"image"
	"image/color"
	"testing"
)

func TestQuantizeNilImage(t *testing.T) {
	_, err := Quantize(nil)
	if err == nil {
		t.Fatal("expected ErrNilImage, got nil")
	}
	if !errors.Is(err, ErrNilImage) {
		t.Fatalf("expected ErrNilImage, got %v", err)
	}
}

func TestQuantizeInvalidCount(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 1, 1))
	_, err := Quantize(img, Count(0))
	if err == nil {
		t.Fatal("expected ErrInvalidCount, got nil")
	}
	if !errors.Is(err, ErrInvalidCount) {
		t.Fatalf("expected ErrInvalidCount, got %v", err)
	}
}

func TestQuantizeReducesColors(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 2, 1))
	img.SetNRGBA(0, 0, color.NRGBA{R: 255, G: 0, B: 0, A: 255})
	img.SetNRGBA(1, 0, color.NRGBA{R: 0, G: 0, B: 255, A: 255})

	out, err := Quantize(img, Count(1))
	if err != nil {
		t.Fatalf("Quantize returned error: %v", err)
	}

	left := ToRGBA(out.At(0, 0))
	right := ToRGBA(out.At(1, 0))
	if left != right {
		t.Fatalf("expected single quantized color, got left=%v right=%v", left, right)
	}
}

// Quantize: fully transparent source pixel → written as transparent NRGBA{}.
func TestQuantizeTransparentPixel(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 2, 1))
	img.SetNRGBA(0, 0, color.NRGBA{R: 255, G: 0, B: 0, A: 255})
	img.SetNRGBA(1, 0, color.NRGBA{A: 0}) // fully transparent

	out, err := Quantize(img, Count(1))
	if err != nil {
		t.Fatalf("Quantize returned error: %v", err)
	}
	transparent := ToRGBA(out.At(1, 0))
	if transparent.A != 0 {
		t.Fatalf("expected transparent pixel to remain transparent, got A=%d", transparent.A)
	}
}

// Quantize: all pixels transparent → palette is empty, returns blank NRGBA image.
func TestQuantizeEmptyPalette(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	// all pixels are zero/transparent by default
	out, err := Quantize(img)
	if err != nil {
		t.Fatalf("Quantize returned error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil image for empty palette")
	}
	if out.Bounds() != img.Bounds() {
		t.Fatalf("expected same bounds, got %v", out.Bounds())
	}
}
