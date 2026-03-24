// Package convert contains internal color conversion helpers.
package convert

import (
	"image/color"
	"math"
)

// Lab is CIE L*a*b* color.
type Lab struct {
	L float64
	A float64
	B float64
}

// RGBAToLab converts RGBA to CIE L*a*b* (D65).
func RGBAToLab(c color.RGBA) Lab {
	x, y, z := rgbaToXYZ(c)
	return xyzToLab(x, y, z)
}

func rgbaToXYZ(c color.RGBA) (float64, float64, float64) {
	r := srgbToLinear(float64(c.R) / 255.0)
	g := srgbToLinear(float64(c.G) / 255.0)
	b := srgbToLinear(float64(c.B) / 255.0)

	x := r*0.4124564 + g*0.3575761 + b*0.1804375
	y := r*0.2126729 + g*0.7151522 + b*0.0721750
	z := r*0.0193339 + g*0.1191920 + b*0.9503041
	return x, y, z
}

func xyzToLab(x, y, z float64) Lab {
	const (
		xn = 0.95047
		yn = 1.00000
		zn = 1.08883
	)

	fx := labF(x / xn)
	fy := labF(y / yn)
	fz := labF(z / zn)

	return Lab{
		L: 116*fy - 16,
		A: 500 * (fx - fy),
		B: 200 * (fy - fz),
	}
}

func labF(t float64) float64 {
	const delta = 6.0 / 29.0
	if t > delta*delta*delta {
		return math.Cbrt(t)
	}
	return t/(3*delta*delta) + 4.0/29.0
}

func srgbToLinear(v float64) float64 {
	if v <= 0.04045 {
		return v / 12.92
	}
	return math.Pow((v+0.055)/1.055, 2.4)
}
