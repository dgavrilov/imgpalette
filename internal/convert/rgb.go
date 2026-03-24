// Package convert contains internal color conversion helpers.
package convert

import "image/color"

// ToRGBA converts any color.Color to 8-bit non-premultiplied RGBA.
func ToRGBA(c color.Color) color.RGBA {
	if c == nil {
		return color.RGBA{}
	}

	switch typedColor := c.(type) {
	case color.RGBA:
		// color.RGBA stores alpha-premultiplied channels.
		// Only return it directly when fully opaque (A==255) — premultiplied == non-premultiplied.
		if typedColor.A == 255 {
			return typedColor
		}
		if typedColor.A == 0 {
			return color.RGBA{}
		}
		// Fallthrough to generic path for partial-alpha premultiplied values.
	case color.NRGBA:
		return color.RGBA(typedColor)
	}

	r, g, b, a := c.RGBA()
	if a == 0 {
		return color.RGBA{}
	}

	alpha := Uint8FromRGBA32(a)
	red := Uint8FromRGBA32(r)
	green := Uint8FromRGBA32(g)
	blue := Uint8FromRGBA32(b)
	if alpha < 255 {
		red = UnpremultiplyChannel(red, alpha)
		green = UnpremultiplyChannel(green, alpha)
		blue = UnpremultiplyChannel(blue, alpha)
	}

	return color.RGBA{R: red, G: green, B: blue, A: alpha}
}

// Uint8FromRGBA32 converts a 16-bit color channel (0..65535) to 8-bit (0..255).
func Uint8FromRGBA32(channel uint32) uint8 {
	value := channel >> 8
	if value > 255 {
		return 255
	}
	return uint8(value)
}

// UnpremultiplyChannel restores channel from alpha-premultiplied representation.
func UnpremultiplyChannel(channel, alpha uint8) uint8 {
	if alpha == 0 {
		return 0
	}
	value := (uint32(channel) * 255) / uint32(alpha)
	if value > 255 {
		return 255
	}
	return uint8(value)
}
