# imgpalette

[![Go Reference](https://pkg.go.dev/badge/github.com/dgavrilov/imgpalette.svg)](https://pkg.go.dev/github.com/dgavrilov/imgpalette)
[![Go Report Card](https://goreportcard.com/badge/github.com/dgavrilov/imgpalette)](https://goreportcard.com/report/github.com/dgavrilov/imgpalette)
[![Go Version](https://img.shields.io/github/go-mod/go-version/dgavrilov/imgpalette)](https://github.com/dgavrilov/imgpalette/blob/master/go.mod)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

`imgpalette` is a small Go library for extracting dominant and accent colors from images using only the Go standard library.

Requires Go `1.21+`.

## Features

- No third-party runtime dependencies.
- Supports `png`, `jpeg`, and `gif` via `image.Decode`.
- Fast dominant-color extraction with configurable pixel sampling and color space.
- Accent color extraction (most saturated/noticeable color).
- Color quantization — redraws an image using a reduced palette.
- Palette preview rendering.
- Convenient color formatting helpers:
  - `#ffffdd`
  - `rgb(255,255,221)`
  - `rgba(255,255,221,0.25)`
  - packed integer (`0xffffdd` as a numeric value)

## Installation

```bash
go get github.com/dgavrilov/imgpalette
```

## Quick Start

### Extract palette from a file

```go
package main

import (
	"fmt"
	"log"

	"github.com/dgavrilov/imgpalette"
)

func main() {
	colors, err := imgpalette.ExtractFile("image.jpg", imgpalette.Count(5))
	if err != nil {
		log.Fatal(err)
	}

	for _, c := range colors {
		fmt.Println(c.HexString(), c.RGBString(), c.Int())
	}
}
```

### Extract accent color from an `io.Reader`

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

	accent, err := imgpalette.AccentReader(f,
		imgpalette.Count(8),
		imgpalette.FilterGray(true),
		imgpalette.MinSaturation(0.15),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(accent.HexString(), "text:", imgpalette.BestTextColor(accent.RGBA))
}
```

## API

### Palette extraction

```go
func Extract(img image.Image, opts ...Option) (Palette, error)
func ExtractReader(r io.Reader, opts ...Option) (Palette, error)
func ExtractFile(path string, opts ...Option) (Palette, error)
```

### Dominant color

```go
func Dominant(img image.Image, opts ...Option) (Color, error)
func DominantReader(r io.Reader, opts ...Option) (Color, error)
func DominantFile(path string, opts ...Option) (Color, error)
```

### Accent color

Returns the most saturated and visually noticeable color.

```go
func Accent(img image.Image, opts ...Option) (Color, error)
func AccentReader(r io.Reader, opts ...Option) (Color, error)
func AccentFile(path string, opts ...Option) (Color, error)
```

### Utilities

```go
func Quantize(img image.Image, opts ...Option) (image.Image, error)
func RenderPalette(p Palette, opts ...RenderOption) image.Image
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

| Option | Default | Description |
|---|---|---|
| `Count(n int)` | `5` | Maximum colors to return (`n > 0`) |
| `Resize(n int)` | `256` | Max image side before extraction; never upscales (`n > 0`) |
| `SampleStride(n int)` | `1` | Pixel sampling step (values ≤ 1 treated as 1) |
| `Space(s ColorSpace)` | `SpaceOKLab` | Color space for bin keying: `SpaceRGB`, `SpaceLab`, `SpaceOKLab` |
| `MinSaturation(v float64)` | `0` | Min HSV saturation threshold for accent selection `[0,1]` |
| `MinCoverage(v float64)` | `0` | Min pixel coverage ratio threshold for accent selection `[0,1]` |
| `FilterGray(enabled bool)` | `false` | Exclude near-gray colors from accent selection (uses `MinSaturation` as threshold) |

## Errors

Exported sentinel errors:

- `ErrNilImage`
- `ErrInvalidCount`
- `ErrInvalidResize`
- `ErrOpenImage`
- `ErrDecodeImage`
- `ErrNoColors`

Behavior notes:

- `Extract*`, `Dominant*`, and `Accent*` return `ErrNoColors` when no extractable colors are found.
- `Quantize` treats `ErrNoColors` as a non-fatal case and returns a blank image with source bounds.
- Wrapped errors support `errors.Is(err, ErrOpenImage)` and `errors.Is(err, ErrDecodeImage)`.

## `Palette` methods

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

## `Color` methods

```go
type Color struct {
	RGBA  color.RGBA
	Count int     // occurrences in the sampled image
	Ratio float64 // fraction of total sampled pixels
}
```

| Method | Example output |
|---|---|
| `String()` / `HexString()` | `#ffffdd` |
| `RGBString()` | `rgb(255,255,221)` |
| `RGBAString(0.5)` | `rgba(255,255,221,0.5)` |
| `Int()` | `16777181` (`0xffffdd`) |
| `RGB()` | `[3]uint8{255, 255, 221}` |
| `RGBAValues(0.5)` | `[4]float64{255, 255, 221, 0.5}` |

## `RenderPalette` options

```go
func RenderSwatchWidth(width int) RenderOption  // default: 64 px
func RenderHeight(height int) RenderOption      // default: 64 px
func RenderPadding(padding int) RenderOption    // default: 0 px
func RenderBackground(c color.Color) RenderOption
```

## Notes

- Fully transparent pixels are always ignored during extraction.
- Colors are internally binned on a 5-5-5 quantization grid for speed.
- `alpha-premultiplied` `color.RGBA` values are correctly unpremultiplied before processing.
- All extraction functions return `ErrInvalidCount` for `Count ≤ 0` and `ErrInvalidResize` for `Resize ≤ 0`.

## Development

```bash
go test ./...                                         # run tests
golangci-lint run ./...                               # lint
go test -coverprofile=cover.out ./...                 # coverage
go tool cover -func=cover.out
```

Current project status:

- test coverage: `100%` statements
- `golangci-lint`: clean

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE).
