package convert

import (
	"image/color"
	"math"
)

// OKLab is OKLab color.
type OKLab struct {
	L float64
	A float64
	B float64
}

// RGBAToOKLab converts RGBA color to OKLab.
func RGBAToOKLab(c color.RGBA) OKLab {
	r := srgbToLinear(float64(c.R) / 255.0)
	g := srgbToLinear(float64(c.G) / 255.0)
	b := srgbToLinear(float64(c.B) / 255.0)

	l := 0.4122214708*r + 0.5363325363*g + 0.0514459929*b
	m := 0.2119034982*r + 0.6806995451*g + 0.1073969566*b
	s := 0.0883024619*r + 0.2817188376*g + 0.6299787005*b

	lc := math.Cbrt(l)
	mc := math.Cbrt(m)
	sc := math.Cbrt(s)

	return OKLab{
		L: 0.2104542553*lc + 0.7936177850*mc - 0.0040720468*sc,
		A: 1.9779984951*lc - 2.4285922050*mc + 0.4505937099*sc,
		B: 0.0259040371*lc + 0.7827717662*mc - 0.8086757660*sc,
	}
}
