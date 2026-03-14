# imgpalette

[![Go Reference](https://pkg.go.dev/badge/github.com/dgavrilov/imgpalette.svg)](https://pkg.go.dev/github.com/dgavrilov/imgpalette)
[![Go Report Card](https://goreportcard.com/badge/github.com/dgavrilov/imgpalette)](https://goreportcard.com/report/github.com/dgavrilov/imgpalette)
[![Go Version](https://img.shields.io/github/go-mod/go-version/dgavrilov/imgpalette)](https://github.com/dgavrilov/imgpalette/blob/master/go.mod)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

`imgpalette` is a small Go library for extracting dominant colors from images using only the Go standard library.

## Features

- No third-party runtime dependencies.
- Supports `png`, `jpeg`, and `gif` via `image.Decode`.
- Fast dominant-color extraction with configurable pixel sampling.
- Convenient color formatting helpers:
  - `#ffffdd`
  - `rgb(255,255,221)`
  - `rgba(255,255,221,0.25)`
  - packed integer (`0xffffdd` as numeric value)

## Installation

```bash
go get github.com/dgavrilov/imgpalette
```

## Quick Start

### From file path

```go
package main

import (
	"fmt"
	"log"

	"github.com/dgavrilov/imgpalette"
)

func main() {
	colors, err := imgpalette.ExtractDominantColorsFromPath("image.jpg", 5, 2)
	if err != nil {
		log.Fatal(err)
	}

	for _, c := range colors {
		fmt.Println(c.HexString(), c.RGBString(), c.Int())
	}
}
```

### From `io.Reader`

```go
package main

import (
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

	colors, err := imgpalette.ExtractDominantColorsFromReader(f, 4, 1)
	if err != nil {
		log.Fatal(err)
	}

	_ = colors
}
```

## API

```go
func ExtractDominantColorsFromReader(imageReader io.Reader, maxColors int, sampleStride int) ([]Color, error)
func ExtractDominantColorsFromPath(imagePath string, maxColors int, sampleStride int) ([]Color, error)
func ExtractDominantColorsFromImage(img image.Image, maxColors int, sampleStride int) []Color
```

### Parameters

- `maxColors`: maximum number of colors to return (`> 0`).
- `sampleStride`: pixel sampling step.
  - `1` = every pixel.
  - `2` = every second pixel on both axes.
  - `<= 1` is treated as `1`.

## `Color` helpers

```go
type Color struct {
	R uint8
	G uint8
	B uint8
}
```

Methods:

- `String() string` (same as `HexString()`)
- `HexString() string`
- `RGBString() string`
- `RGBAString(alpha float64) string`
- `Int() uint32`
- `RGB() [3]uint8`
- `RGBA(alpha float64) [4]float64`
- `QuantizedBinIndex() int`

## Notes

- Fully transparent pixels are ignored.
- `ExtractDominantColorsFromReader` and `ExtractDominantColorsFromPath` return `ErrInvalidMaxColors` if `maxColors <= 0`.
- Internally, colors are quantized to a 5-5-5 RGB grid for speed.

## Development

Run tests:

```bash
go test ./...
```

Run linter:

```bash
golangci-lint run ./...
```

Run coverage:

```bash
go test -coverprofile=cover.out ./...
go tool cover -func=cover.out
```

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE).
