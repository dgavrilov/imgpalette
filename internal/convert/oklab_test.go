package convert

import (
	"image/color"
	"testing"
)

func TestRGBAToOKLabBlack(t *testing.T) {
	ok := RGBAToOKLab(color.RGBA{A: 255})
	if ok.L != 0 || ok.A != 0 || ok.B != 0 {
		t.Fatalf("unexpected OKLab for black: %+v", ok)
	}
}
