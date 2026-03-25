package sample

import (
	"image"
	"testing"
)

func TestResizeToMaxSide(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 400, 200))
	resized := ResizeToMaxSide(img, 100)
	if resized.Bounds().Dx() != 100 || resized.Bounds().Dy() != 50 {
		t.Fatalf("unexpected resized bounds: %v", resized.Bounds())
	}
}

func TestResizeToMaxSideNoUpscale(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 40, 20))
	resized := ResizeToMaxSide(img, 100)
	if resized.Bounds().Dx() != 40 || resized.Bounds().Dy() != 20 {
		t.Fatalf("unexpected bounds for no-upscale: %v", resized.Bounds())
	}
}

// Portrait image: height > width → newHeight = maxSide, newWidth computed.
func TestResizeToMaxSidePortrait(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 100, 400))
	resized := ResizeToMaxSide(img, 100)
	if resized.Bounds().Dy() != 100 {
		t.Fatalf("expected height=100 for portrait, got %d", resized.Bounds().Dy())
	}
	if resized.Bounds().Dx() != 25 {
		t.Fatalf("expected width=25 for portrait, got %d", resized.Bounds().Dx())
	}
}

// nil image → returns nil as-is.
func TestResizeToMaxSideNilImage(t *testing.T) {
	if ResizeToMaxSide(nil, 100) != nil {
		t.Fatal("expected nil for nil image")
	}
}

// maxSide <= 0 → returns image as-is.
func TestResizeToMaxSideZeroMaxSide(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 400, 200))
	if ResizeToMaxSide(img, 0) != img {
		t.Fatal("expected original image for maxSide=0")
	}
}

// Zero-dimension image (width=0) → returns as-is.
func TestResizeToMaxSideZeroDimension(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 0, 100))
	if ResizeToMaxSide(img, 50) != img {
		t.Fatal("expected original image for zero-width image")
	}
}

// Very tall narrow image → newWidth rounds down to 0 → clamped to 1.
func TestResizeToMaxSideNewWidthClamp(t *testing.T) {
	// 1px wide, 10000px tall: newHeight=100, newWidth = 1*100/10000 = 0 → clamped to 1.
	img := image.NewNRGBA(image.Rect(0, 0, 1, 10000))
	resized := ResizeToMaxSide(img, 100)
	if resized.Bounds().Dx() < 1 {
		t.Fatalf("expected newWidth >= 1, got %d", resized.Bounds().Dx())
	}
}

// Very wide short image → newHeight rounds down to 0 → clamped to 1.
func TestResizeToMaxSideNewHeightClamp(t *testing.T) {
	// 10000px wide, 1px tall: newWidth=100, newHeight = 1*100/10000 = 0 → clamped to 1.
	img := image.NewNRGBA(image.Rect(0, 0, 10000, 1))
	resized := ResizeToMaxSide(img, 100)
	if resized.Bounds().Dy() < 1 {
		t.Fatalf("expected newHeight >= 1, got %d", resized.Bounds().Dy())
	}
}
