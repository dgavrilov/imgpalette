# imgpalette

[![Go Reference](https://pkg.go.dev/badge/github.com/dgavrilov/imgpalette.svg)](https://pkg.go.dev/github.com/dgavrilov/imgpalette)
[![Go Report Card](https://goreportcard.com/badge/github.com/dgavrilov/imgpalette)](https://goreportcard.com/report/github.com/dgavrilov/imgpalette)
[![Go Version](https://img.shields.io/github/go-mod/go-version/dgavrilov/imgpalette)](https://github.com/dgavrilov/imgpalette/blob/master/go.mod)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Package `imgpalette` extracts dominant palettes and accent colors from images using only the Go standard library.

It supports `png`, `jpeg`, and `gif` through `image.Decode`, can quantize images to a reduced palette, and can render palette previews, including a demo PNG with transparent background and rounded swatches.

Requires Go `1.21+`.

## Installation

```bash
go get github.com/dgavrilov/imgpalette
```

## Quick Start

### Extract a palette and save a demo PNG

```go
package main

import (
	"fmt"
	"log"

	"github.com/dgavrilov/imgpalette"
)

func main() {
	palette, err := imgpalette.ExtractFile("image.jpg", imgpalette.Count(5))
	if err != nil {
		log.Fatal(err)
	}

	for _, c := range palette {
		fmt.Println(c.Hex(), c.RGBString(), c.Int())
	}

	if err := imgpalette.RenderFileDemo("palette.png", palette); err != nil {
		log.Fatal(err)
	}
}
```

### Pick an accent color

```go
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/dgavrilov/imgpalette"
)

func main() {
	f, err := os.Open("image.png")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	accent, err := imgpalette.AccentReader(
		f,
		imgpalette.Count(8),
		imgpalette.FilterGray(true),
		imgpalette.MinSaturation(0.15),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(accent.Hex(), "text:", imgpalette.BestTextColor(accent.RGBA))
}
```

## Package Overview

### Extraction

```go
func Extract(img image.Image, opts ...Option) (Palette, error)
func ExtractReader(r io.Reader, opts ...Option) (Palette, error)
func ExtractFile(path string, opts ...Option) (Palette, error)

func Dominant(img image.Image, opts ...Option) (Color, error)
func DominantReader(r io.Reader, opts ...Option) (Color, error)
func DominantFile(path string, opts ...Option) (Color, error)

func Accent(img image.Image, opts ...Option) (Color, error)
func AccentReader(r io.Reader, opts ...Option) (Color, error)
func AccentFile(path string, opts ...Option) (Color, error)
```

### Rendering and Utilities

```go
func Quantize(img image.Image, opts ...Option) (image.Image, error)
func RenderPalette(p Palette, opts ...RenderOption) image.Image
func RenderFileDemo(path string, p Palette) error

func BestTextColor(bg color.Color) color.RGBA
func Contrast(a, b color.Color) float64
func Distance(a, b color.Color) float64
func ToRGBA(c color.Color) color.RGBA
func Hex(c color.Color) string
func Brightness(c color.Color) float64
func Saturation(c color.Color) float64
func IsGray(c color.Color, threshold float64) bool
```

## Options

Extraction options:

| Option | Default | Description |
| --- | --- | --- |
| `Count(n int)` | `5` | Maximum number of colors to return. |
| `Resize(n int)` | `256` | Maximum image side before extraction; never upscales. |
| `SampleStride(n int)` | `1` | Pixel sampling step. Values `<= 1` are treated as `1`. |
| `Space(s ColorSpace)` | `SpaceOKLab` | Bin keying space: `SpaceRGB`, `SpaceLab`, `SpaceOKLab`. |
| `MinSaturation(v float64)` | `0` | Minimum HSV saturation for accent selection in `[0,1]`. |
| `MinCoverage(v float64)` | `0` | Minimum coverage ratio for accent selection in `[0,1]`. |
| `FilterGray(enabled bool)` | `false` | Exclude near-gray colors during accent selection. |

Render options for `RenderPalette`:

| Option | Default | Description |
| --- | --- | --- |
| `RenderSwatchWidth(width int)` | `64` | Swatch width in pixels. |
| `RenderSwatchHeight(height int)` | `64` | Swatch height in pixels. |
| `RenderPadding(padding int)` | `0` | Padding between swatches in pixels. |
| `RenderBackground(c color.Color)` | transparent | Background color for preview and padding areas. |

`RenderFileDemo` uses a fixed demo layout: transparent background, equal outer and inner spacing, and rounded swatch corners.

## Types

### Palette

```go
type Palette []Color
```

Methods:

```go
func (p Palette) Colors() []color.RGBA
func (p Palette) Hex() []string
func (p Palette) Dominant() Color
func (p Palette) Nearest(c color.Color) Color
func (p Palette) MergeSimilar(threshold float64) Palette
func (p Palette) SortByFrequency()
func (p Palette) SortByBrightness()
func (p Palette) SortBySaturation()
```

### Color

```go
type Color struct {
	RGBA  color.RGBA
	Count int
	Ratio float64
}
```

Methods:

```go
func (c Color) String() string
func (c Color) Hex() string
func (c Color) RGBString() string
func (c Color) RGBAString(alpha float64) string
func (c Color) Int() uint32
func (c Color) RGB() [3]uint8
func (c Color) RGBAValues(alpha float64) [4]float64
func (c Color) QuantizedBinIndex() int
```

## Errors

Exported sentinel errors:

- `ErrNilImage`
- `ErrInvalidCount`
- `ErrInvalidResize`
- `ErrOpenImage`
- `ErrDecodeImage`
- `ErrNoColors`

Behavior:

- `Extract*`, `Dominant*`, and `Accent*` return `ErrNoColors` when no extractable colors are found.
- `Quantize` treats `ErrNoColors` as non-fatal and returns a blank image with the source bounds.
- Wrapped errors support `errors.Is(err, ErrOpenImage)` and `errors.Is(err, ErrDecodeImage)`.

## Notes

- Fully transparent pixels are ignored during extraction.
- Colors are internally binned on a 5-5-5 quantization grid for speed.
- Nearby dominant bins are merged before truncating the palette, which reduces near-duplicate colors in photo-heavy inputs.
- Premultiplied `color.RGBA` values are unpremultiplied before processing.
- Invalid `Count` and `Resize` values return `ErrInvalidCount` and `ErrInvalidResize`.

## Development

```bash
go test ./...
golangci-lint run ./...
go test -coverprofile=cover.out ./...
go tool cover -func=cover.out
```

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE).
